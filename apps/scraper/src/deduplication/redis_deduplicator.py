"""
Redis-based deduplication service for court slots.

This module provides efficient deduplication of scraped court slots using Redis
with expiring keys to prevent reprocessing of recently seen slots.
"""

import redis
import logging
import hashlib
import os
from typing import Dict, Any, Optional, Tuple
from datetime import datetime, timedelta


class RedisDeduplicator:
    """
    Redis-based deduplication service for court slots.
    
    Uses Redis SET with EX (expiry) and NX (only if not exists) options to track
    recently seen slots and prevent duplicate processing.
    """
    
    def __init__(self, 
                 redis_host: Optional[str] = None, 
                 redis_port: Optional[int] = None, 
                 redis_password: Optional[str] = None, 
                 redis_db: int = 0,
                 expiry_hours: int = 48):
        """
        Initialize the Redis deduplicator.
        
        Args:
            redis_host: Redis server hostname (uses REDIS_HOST env var or 'tennis-redis' if None)
            redis_port: Redis server port (uses REDIS_PORT env var or 6379 if None)
            redis_password: Redis password (uses REDIS_PASSWORD env var if None)
            redis_db: Redis database number
            expiry_hours: Hours to keep deduplication keys (default: 48)
        """
        # Use environment variables with fallback to Docker service names
        if redis_host is None:
            redis_host = os.getenv('REDIS_HOST', 'tennis-redis')
        if redis_port is None:
            redis_port = int(os.getenv('REDIS_PORT', '6379'))
        if redis_password is None:
            redis_password = os.getenv('REDIS_PASSWORD')
            
        # Initialize with resolved values
        self.redis_host = redis_host
        self.redis_port = redis_port
        self.redis_password = redis_password
        self.redis_db = redis_db
        self.expiry_seconds = expiry_hours * 3600
        self.client = None
        self.logger = logging.getLogger(__name__)
        
        # Metrics tracking
        self.metrics = {
            'total_checks': 0,
            'duplicates_found': 0,
            'new_slots': 0,
            'redis_errors': 0,
            'connection_failures': 0
        }
    
    def connect(self) -> bool:
        """
        Connect to Redis server.
        
        Returns:
            bool: True if connection successful, False otherwise
        """
        try:
            # Try connection without password first
            self.client = redis.Redis(
                host=self.redis_host,
                port=self.redis_port,
                db=self.redis_db,
                decode_responses=True,
                socket_timeout=5,
                socket_connect_timeout=5
            )
            
            # Test connection
            self.client.ping()
            self.logger.info(f"✅ Connected to Redis deduplicator at {self.redis_host}:{self.redis_port} (no auth)")
            return True
            
        except Exception as e:
            # If no-auth fails, try with password
            if self.redis_password:
                try:
                    self.client = redis.Redis(
                        host=self.redis_host,
                        port=self.redis_port,
                        password=self.redis_password,
                        db=self.redis_db,
                        decode_responses=True,
                        socket_timeout=5,
                        socket_connect_timeout=5
                    )
                    
                    # Test connection
                    self.client.ping()
                    self.logger.info(f"✅ Connected to Redis deduplicator at {self.redis_host}:{self.redis_port} (with auth)")
                    return True
                    
                except Exception as e2:
                    self.logger.error(f"❌ Failed to connect to Redis with auth: {e2}")
                    self.metrics['connection_failures'] += 1
                    return False
            else:
                self.logger.error(f"❌ Failed to connect to Redis: {e}")
                self.metrics['connection_failures'] += 1
                return False
    
    def generate_slot_key(self, slot_data: Dict[str, Any]) -> str:
        """
        Generate a unique Redis key for a court slot.
        
        Key format: dedupe:slot:<venueId>:<date>:<startTime>:<courtId>
        
        Args:
            slot_data: Dictionary containing slot information
            
        Returns:
            str: Unique Redis key for the slot
        """
        # Extract required fields
        venue_id = str(slot_data.get('venue_id', ''))
        date = slot_data.get('date', '')
        start_time = slot_data.get('start_time', '')
        court_id = slot_data.get('court_id', '')
        
        # Create a deterministic key
        key_parts = [venue_id, date, start_time, court_id]
        key_suffix = ':'.join(key_parts)
        
        # Use hash for very long keys to avoid Redis key length limits
        if len(key_suffix) > 200:
            key_hash = hashlib.md5(key_suffix.encode()).hexdigest()
            return f"dedupe:slot:hash:{key_hash}"
        
        return f"dedupe:slot:{key_suffix}"
    
    def is_duplicate_slot(self, slot_data: Dict[str, Any]) -> Tuple[bool, str]:
        """
        Check if a slot is a duplicate using Redis deduplication.
        
        Uses Redis SET with EX (expiry) and NX (only if not exists) to atomically
        check and mark a slot as seen.
        
        Args:
            slot_data: Dictionary containing slot information
            
        Returns:
            Tuple[bool, str]: (is_duplicate, redis_key)
                - is_duplicate: True if slot was already seen, False if new
                - redis_key: The Redis key used for this slot
        """
        self.metrics['total_checks'] += 1
        
        if not self.client:
            if not self.connect():
                self.logger.warning("Redis not available, treating slot as new")
                return False, ""
        
        try:
            # Generate unique key for this slot
            redis_key = self.generate_slot_key(slot_data)
            
            # Use SET with EX (expiry) and NX (only if not exists)
            # Returns True if key was set (new slot), False if key already existed (duplicate)
            slot_timestamp = datetime.now().isoformat()
            was_set = self.client.set(
                redis_key, 
                slot_timestamp, 
                ex=self.expiry_seconds, 
                nx=True
            )
            
            if was_set:
                # Key was set, this is a new slot
                self.metrics['new_slots'] += 1
                self.logger.debug(f"🆕 New slot marked in Redis: {redis_key}")
                return False, redis_key
            else:
                # Key already existed, this is a duplicate
                self.metrics['duplicates_found'] += 1
                self.logger.debug(f"🔄 Duplicate slot detected: {redis_key}")
                return True, redis_key
                
        except Exception as e:
            self.logger.error(f"Redis deduplication error: {e}")
            self.metrics['redis_errors'] += 1
            # On Redis error, treat as new slot to avoid losing data
            return False, ""
    
    def check_multiple_slots(self, slots_data: list) -> Tuple[list, list]:
        """
        Check multiple slots for duplicates efficiently.
        
        Args:
            slots_data: List of slot dictionaries
            
        Returns:
            Tuple[list, list]: (new_slots, duplicate_slots)
        """
        new_slots = []
        duplicate_slots = []
        
        for slot in slots_data:
            is_duplicate, redis_key = self.is_duplicate_slot(slot)
            
            if is_duplicate:
                duplicate_slots.append({
                    'slot': slot,
                    'redis_key': redis_key
                })
            else:
                new_slots.append(slot)
        
        return new_slots, duplicate_slots
    
    def get_slot_info(self, slot_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """
        Get information about a slot from Redis (if it exists).
        
        Args:
            slot_data: Dictionary containing slot information
            
        Returns:
            Optional[Dict[str, Any]]: Slot info from Redis or None if not found
        """
        if not self.client:
            if not self.connect():
                return None
        
        try:
            redis_key = self.generate_slot_key(slot_data)
            value = self.client.get(redis_key)
            
            if value:
                ttl = self.client.ttl(redis_key)
                return {
                    'redis_key': redis_key,
                    'first_seen': value,
                    'ttl_seconds': ttl,
                    'expires_at': datetime.now() + timedelta(seconds=ttl) if ttl > 0 else None
                }
            
            return None
            
        except Exception as e:
            self.logger.error(f"Error getting slot info: {e}")
            return None
    
    def remove_slot(self, slot_data: Dict[str, Any]) -> bool:
        """
        Manually remove a slot from deduplication cache.
        
        Args:
            slot_data: Dictionary containing slot information
            
        Returns:
            bool: True if slot was removed, False otherwise
        """
        if not self.client:
            if not self.connect():
                return False
        
        try:
            redis_key = self.generate_slot_key(slot_data)
            result = self.client.delete(redis_key)
            
            if result:
                self.logger.debug(f"🗑️ Removed slot from deduplication: {redis_key}")
                return True
            else:
                self.logger.debug(f"🤷 Slot not found in deduplication cache: {redis_key}")
                return False
                
        except Exception as e:
            self.logger.error(f"Error removing slot: {e}")
            return False
    
    def get_metrics(self) -> Dict[str, Any]:
        """
        Get deduplication metrics.
        
        Returns:
            Dict[str, Any]: Metrics dictionary with counts and rates
        """
        total_checks = self.metrics['total_checks']
        
        metrics = self.metrics.copy()
        
        if total_checks > 0:
            metrics['duplicate_rate'] = self.metrics['duplicates_found'] / total_checks
            metrics['new_slot_rate'] = self.metrics['new_slots'] / total_checks
            metrics['error_rate'] = self.metrics['redis_errors'] / total_checks
        else:
            metrics['duplicate_rate'] = 0.0
            metrics['new_slot_rate'] = 0.0
            metrics['error_rate'] = 0.0
        
        return metrics
    
    def reset_metrics(self):
        """Reset all metrics counters."""
        self.metrics = {
            'total_checks': 0,
            'duplicates_found': 0,
            'new_slots': 0,
            'redis_errors': 0,
            'connection_failures': 0
        }
    
    def get_cache_stats(self) -> Optional[Dict[str, Any]]:
        """
        Get Redis cache statistics for deduplication keys.
        
        Returns:
            Optional[Dict[str, Any]]: Cache statistics or None if Redis unavailable
        """
        if not self.client:
            if not self.connect():
                return None
        
        try:
            # Count deduplication keys
            dedupe_keys = self.client.keys("dedupe:slot:*")
            
            # Get memory usage info
            info = self.client.info('memory')
            
            return {
                'dedupe_keys_count': len(dedupe_keys),
                'redis_memory_used': info.get('used_memory_human', 'unknown'),
                'redis_memory_peak': info.get('used_memory_peak_human', 'unknown'),
                'expiry_hours': self.expiry_seconds / 3600
            }
            
        except Exception as e:
            self.logger.error(f"Error getting cache stats: {e}")
            return None
    
    def cleanup_expired_keys(self) -> int:
        """
        Manually cleanup expired deduplication keys (Redis handles this automatically).
        
        This is mainly for testing and monitoring purposes.
        
        Returns:
            int: Number of keys that were already expired
        """
        if not self.client:
            if not self.connect():
                return 0
        
        try:
            dedupe_keys = self.client.keys("dedupe:slot:*")
            expired_count = 0
            
            for key in dedupe_keys:
                ttl = self.client.ttl(key)
                if ttl == -2:  # Key doesn't exist (expired)
                    expired_count += 1
            
            self.logger.info(f"Found {expired_count} expired deduplication keys out of {len(dedupe_keys)} total")
            return expired_count
            
        except Exception as e:
            self.logger.error(f"Error during cleanup: {e}")
            return 0
    
    def close(self):
        """Close Redis connection."""
        if self.client:
            try:
                self.client.close()
                self.logger.info("Redis deduplicator connection closed")
            except Exception as e:
                self.logger.error(f"Error closing Redis connection: {e}") 