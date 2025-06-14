"""
Unit tests for Redis-based slot deduplication.

Tests the RedisDeduplicator class functionality including connection handling,
key generation, duplicate detection, and metrics tracking.
"""

import unittest
from unittest.mock import Mock, patch, MagicMock
import redis
from datetime import datetime, timedelta
import hashlib

# Add the src directory to the path for imports
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

from deduplication.redis_deduplicator import RedisDeduplicator


class TestRedisDeduplicator(unittest.TestCase):
    """Test cases for RedisDeduplicator class."""
    
    def setUp(self):
        """Set up test fixtures."""
        self.deduplicator = RedisDeduplicator(
            redis_host='localhost',
            redis_port=6379,
            redis_password=None,
            redis_db=0,
            expiry_hours=48
        )
        
        # Sample slot data for testing
        self.sample_slot = {
            'venue_id': '507f1f77bcf86cd799439011',
            'date': '2024-06-15',
            'start_time': '18:00',
            'court_id': 'court_1',
            'court_name': 'Court 1',
            'end_time': '19:00',
            'price': 25.0,
            'available': True
        }
        
        self.sample_slot_2 = {
            'venue_id': '507f1f77bcf86cd799439012',
            'date': '2024-06-15',
            'start_time': '19:00',
            'court_id': 'court_2',
            'court_name': 'Court 2',
            'end_time': '20:00',
            'price': 30.0,
            'available': True
        }
    
    def test_initialization(self):
        """Test RedisDeduplicator initialization."""
        dedup = RedisDeduplicator(
            redis_host='test-host',
            redis_port=1234,
            redis_password='test-pass',
            redis_db=5,
            expiry_hours=24
        )
        
        self.assertEqual(dedup.redis_host, 'test-host')
        self.assertEqual(dedup.redis_port, 1234)
        self.assertEqual(dedup.redis_password, 'test-pass')
        self.assertEqual(dedup.redis_db, 5)
        self.assertEqual(dedup.expiry_seconds, 24 * 3600)
        self.assertIsNone(dedup.client)
        
        # Check initial metrics
        expected_metrics = {
            'total_checks': 0,
            'duplicates_found': 0,
            'new_slots': 0,
            'redis_errors': 0,
            'connection_failures': 0
        }
        self.assertEqual(dedup.metrics, expected_metrics)
    
    def test_generate_slot_key_normal(self):
        """Test slot key generation for normal-length keys."""
        expected_key = "dedupe:slot:507f1f77bcf86cd799439011:2024-06-15:18:00:court_1"
        actual_key = self.deduplicator.generate_slot_key(self.sample_slot)
        self.assertEqual(actual_key, expected_key)
    
    def test_generate_slot_key_long(self):
        """Test slot key generation for very long keys (should use hash)."""
        long_slot = {
            'venue_id': 'a' * 100,
            'date': 'b' * 50,
            'start_time': 'c' * 50,
            'court_id': 'd' * 50
        }
        
        key = self.deduplicator.generate_slot_key(long_slot)
        self.assertTrue(key.startswith("dedupe:slot:hash:"))
        
        # Verify it's a valid MD5 hash
        hash_part = key.replace("dedupe:slot:hash:", "")
        self.assertEqual(len(hash_part), 32)  # MD5 hash length
    
    def test_generate_slot_key_missing_fields(self):
        """Test slot key generation with missing fields."""
        incomplete_slot = {
            'venue_id': '507f1f77bcf86cd799439011',
            'date': '2024-06-15'
            # Missing start_time and court_id
        }
        
        expected_key = "dedupe:slot:507f1f77bcf86cd799439011:2024-06-15::"
        actual_key = self.deduplicator.generate_slot_key(incomplete_slot)
        self.assertEqual(actual_key, expected_key)
    
    @patch('redis.Redis')
    def test_connect_success_no_auth(self, mock_redis_class):
        """Test successful Redis connection without authentication."""
        mock_client = Mock()
        mock_redis_class.return_value = mock_client
        mock_client.ping.return_value = True
        
        result = self.deduplicator.connect()
        
        self.assertTrue(result)
        self.assertEqual(self.deduplicator.client, mock_client)
        mock_redis_class.assert_called_once_with(
            host='localhost',
            port=6379,
            db=0,
            decode_responses=True,
            socket_timeout=5,
            socket_connect_timeout=5
        )
        mock_client.ping.assert_called_once()
    
    @patch('redis.Redis')
    def test_connect_success_with_auth(self, mock_redis_class):
        """Test successful Redis connection with authentication."""
        self.deduplicator.redis_password = 'test-password'
        
        mock_client_no_auth = Mock()
        mock_client_with_auth = Mock()
        
        # First call (no auth) fails, second call (with auth) succeeds
        mock_redis_class.side_effect = [mock_client_no_auth, mock_client_with_auth]
        mock_client_no_auth.ping.side_effect = redis.ConnectionError("Auth required")
        mock_client_with_auth.ping.return_value = True
        
        result = self.deduplicator.connect()
        
        self.assertTrue(result)
        self.assertEqual(self.deduplicator.client, mock_client_with_auth)
        self.assertEqual(mock_redis_class.call_count, 2)
    
    @patch('redis.Redis')
    def test_connect_failure(self, mock_redis_class):
        """Test Redis connection failure."""
        mock_client = Mock()
        mock_redis_class.return_value = mock_client
        mock_client.ping.side_effect = redis.ConnectionError("Connection failed")
        
        result = self.deduplicator.connect()
        
        self.assertFalse(result)
        self.assertEqual(self.deduplicator.metrics['connection_failures'], 1)
    
    def test_is_duplicate_slot_new_slot(self):
        """Test duplicate check for a new slot."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        mock_client.set.return_value = True  # Key was set (new slot)
        
        is_duplicate, redis_key = self.deduplicator.is_duplicate_slot(self.sample_slot)
        
        self.assertFalse(is_duplicate)
        self.assertTrue(redis_key.startswith("dedupe:slot:"))
        self.assertEqual(self.deduplicator.metrics['new_slots'], 1)
        self.assertEqual(self.deduplicator.metrics['total_checks'], 1)
        
        # Verify Redis SET call
        mock_client.set.assert_called_once()
        call_args = mock_client.set.call_args
        self.assertEqual(call_args[1]['ex'], 48 * 3600)  # 48 hours
        self.assertTrue(call_args[1]['nx'])  # Only if not exists
    
    def test_is_duplicate_slot_duplicate(self):
        """Test duplicate check for an existing slot."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        mock_client.set.return_value = False  # Key already existed (duplicate)
        
        is_duplicate, redis_key = self.deduplicator.is_duplicate_slot(self.sample_slot)
        
        self.assertTrue(is_duplicate)
        self.assertTrue(redis_key.startswith("dedupe:slot:"))
        self.assertEqual(self.deduplicator.metrics['duplicates_found'], 1)
        self.assertEqual(self.deduplicator.metrics['total_checks'], 1)
    
    @patch.object(RedisDeduplicator, 'connect')
    def test_is_duplicate_slot_no_connection(self, mock_connect):
        """Test duplicate check when Redis is not available."""
        mock_connect.return_value = False
        self.deduplicator.client = None
        
        is_duplicate, redis_key = self.deduplicator.is_duplicate_slot(self.sample_slot)
        
        self.assertFalse(is_duplicate)  # Treat as new when Redis unavailable
        self.assertEqual(redis_key, "")
        mock_connect.assert_called_once()
    
    def test_is_duplicate_slot_redis_error(self):
        """Test duplicate check when Redis operation fails."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        mock_client.set.side_effect = redis.RedisError("Redis operation failed")
        
        is_duplicate, redis_key = self.deduplicator.is_duplicate_slot(self.sample_slot)
        
        self.assertFalse(is_duplicate)  # Treat as new on error
        self.assertEqual(redis_key, "")
        self.assertEqual(self.deduplicator.metrics['redis_errors'], 1)
    
    def test_check_multiple_slots(self):
        """Test checking multiple slots for duplicates."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        
        # First slot is new, second is duplicate
        mock_client.set.side_effect = [True, False]
        
        slots = [self.sample_slot, self.sample_slot_2]
        new_slots, duplicate_slots = self.deduplicator.check_multiple_slots(slots)
        
        self.assertEqual(len(new_slots), 1)
        self.assertEqual(len(duplicate_slots), 1)
        self.assertEqual(new_slots[0], self.sample_slot)
        self.assertEqual(duplicate_slots[0]['slot'], self.sample_slot_2)
        self.assertTrue('redis_key' in duplicate_slots[0])
    
    def test_get_slot_info_exists(self):
        """Test getting slot info for an existing slot."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        
        timestamp = "2024-06-14T10:30:00"
        ttl = 3600  # 1 hour remaining
        
        mock_client.get.return_value = timestamp
        mock_client.ttl.return_value = ttl
        
        info = self.deduplicator.get_slot_info(self.sample_slot)
        
        self.assertIsNotNone(info)
        self.assertEqual(info['first_seen'], timestamp)
        self.assertEqual(info['ttl_seconds'], ttl)
        self.assertIsNotNone(info['expires_at'])
        self.assertTrue('redis_key' in info)
    
    def test_get_slot_info_not_exists(self):
        """Test getting slot info for a non-existent slot."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        mock_client.get.return_value = None
        
        info = self.deduplicator.get_slot_info(self.sample_slot)
        
        self.assertIsNone(info)
    
    def test_remove_slot_success(self):
        """Test successful slot removal."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        mock_client.delete.return_value = 1  # Key was deleted
        
        result = self.deduplicator.remove_slot(self.sample_slot)
        
        self.assertTrue(result)
        mock_client.delete.assert_called_once()
    
    def test_remove_slot_not_found(self):
        """Test slot removal when slot doesn't exist."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        mock_client.delete.return_value = 0  # Key not found
        
        result = self.deduplicator.remove_slot(self.sample_slot)
        
        self.assertFalse(result)
    
    def test_get_metrics(self):
        """Test metrics calculation."""
        # Set some test metrics
        self.deduplicator.metrics = {
            'total_checks': 100,
            'duplicates_found': 30,
            'new_slots': 70,
            'redis_errors': 5,
            'connection_failures': 2
        }
        
        metrics = self.deduplicator.get_metrics()
        
        self.assertEqual(metrics['total_checks'], 100)
        self.assertEqual(metrics['duplicates_found'], 30)
        self.assertEqual(metrics['new_slots'], 70)
        self.assertEqual(metrics['redis_errors'], 5)
        self.assertEqual(metrics['connection_failures'], 2)
        self.assertEqual(metrics['duplicate_rate'], 0.3)
        self.assertEqual(metrics['new_slot_rate'], 0.7)
        self.assertEqual(metrics['error_rate'], 0.05)
    
    def test_get_metrics_no_checks(self):
        """Test metrics calculation when no checks have been performed."""
        metrics = self.deduplicator.get_metrics()
        
        self.assertEqual(metrics['duplicate_rate'], 0.0)
        self.assertEqual(metrics['new_slot_rate'], 0.0)
        self.assertEqual(metrics['error_rate'], 0.0)
    
    def test_reset_metrics(self):
        """Test metrics reset."""
        # Set some test metrics
        self.deduplicator.metrics['total_checks'] = 100
        self.deduplicator.metrics['duplicates_found'] = 30
        
        self.deduplicator.reset_metrics()
        
        expected_metrics = {
            'total_checks': 0,
            'duplicates_found': 0,
            'new_slots': 0,
            'redis_errors': 0,
            'connection_failures': 0
        }
        self.assertEqual(self.deduplicator.metrics, expected_metrics)
    
    def test_get_cache_stats(self):
        """Test cache statistics retrieval."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        
        # Mock Redis responses
        mock_client.keys.return_value = ['dedupe:slot:key1', 'dedupe:slot:key2']
        mock_client.info.return_value = {
            'used_memory_human': '1.5M',
            'used_memory_peak_human': '2.0M'
        }
        
        stats = self.deduplicator.get_cache_stats()
        
        self.assertIsNotNone(stats)
        self.assertEqual(stats['dedupe_keys_count'], 2)
        self.assertEqual(stats['redis_memory_used'], '1.5M')
        self.assertEqual(stats['redis_memory_peak'], '2.0M')
        self.assertEqual(stats['expiry_hours'], 48)
    
    def test_cleanup_expired_keys(self):
        """Test expired keys cleanup."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        
        # Mock keys and TTL responses
        mock_client.keys.return_value = ['key1', 'key2', 'key3']
        mock_client.ttl.side_effect = [3600, -2, 1800]  # active, expired, active
        
        expired_count = self.deduplicator.cleanup_expired_keys()
        
        self.assertEqual(expired_count, 1)
        self.assertEqual(mock_client.ttl.call_count, 3)
    
    def test_close_connection(self):
        """Test closing Redis connection."""
        mock_client = Mock()
        self.deduplicator.client = mock_client
        
        self.deduplicator.close()
        
        mock_client.close.assert_called_once()


if __name__ == '__main__':
    unittest.main() 