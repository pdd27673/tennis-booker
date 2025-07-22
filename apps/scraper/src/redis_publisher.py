import redis
import json
import logging
import os
from typing import Dict, Any, Optional
from datetime import datetime

class RedisPublisher:
    """Redis publisher for sending slot notifications to the notification service"""
    
    def __init__(self, redis_host=None, redis_port=None, redis_password=None, redis_db=0):
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
        self.client = None
        self.queue_name = 'court_slots'
        self.logger = logging.getLogger(__name__)
        
    def connect(self):
        """Connect to Redis"""
        try:
            # First try without password
            self.client = redis.Redis(
                host=self.redis_host,
                port=self.redis_port,
                db=self.redis_db,
                decode_responses=True
            )
            
            # Test connection
            self.client.ping()
            self.logger.info(f"✅ Connected to Redis at {self.redis_host}:{self.redis_port} (no auth)")
            return True
            
        except Exception as e:
            # If no-auth fails, try with password
            try:
                self.client = redis.Redis(
                    host=self.redis_host,
                    port=self.redis_port,
                    password=self.redis_password,
                    db=self.redis_db,
                    decode_responses=True
                )
                
                # Test connection
                self.client.ping()
                self.logger.info(f"✅ Connected to Redis at {self.redis_host}:{self.redis_port} (with auth)")
                return True
                
            except Exception as e2:
                self.logger.error(f"❌ Failed to connect to Redis: {e2}")
                return False
    
    def publish_slot(self, slot_data: Dict[str, Any]) -> bool:
        """
        Publish a slot notification to Redis queue
        
        Args:
            slot_data: Dictionary containing slot information
            
        Returns:
            bool: True if successful, False otherwise
        """
        if not self.client:
            if not self.connect():
                return False
                
        try:
            # Ensure required fields are present
            required_fields = ['venueId', 'venueName', 'courtId', 'courtName', 
                             'date', 'startTime', 'endTime', 'price', 'isAvailable', 'bookingUrl']
            
            for field in required_fields:
                if field not in slot_data:
                    self.logger.warning(f"Missing required field '{field}' in slot data")
                    return False
            
            # Add timestamp if not present
            if 'scrapedAt' not in slot_data:
                slot_data['scrapedAt'] = datetime.now().isoformat()
            
            # Convert to JSON
            slot_json = json.dumps(slot_data, default=str)
            
            # Push to Redis queue
            result = self.client.lpush(self.queue_name, slot_json)
            
            if result:
                self.logger.info(f"Published slot notification: {slot_data['venueName']} - {slot_data['courtName']} on {slot_data['date']} at {slot_data['startTime']}")
                return True
            else:
                self.logger.error("Failed to push to Redis queue")
                return False
                
        except Exception as e:
            self.logger.error(f"Error publishing slot notification: {e}")
            return False
    
    def publish_new_slots(self, new_slots: list) -> int:
        """
        Publish multiple new slot notifications
        
        Args:
            new_slots: List of slot dictionaries
            
        Returns:
            int: Number of successfully published notifications
        """
        if not new_slots:
            return 0
            
        published_count = 0
        
        for slot in new_slots:
            if self.publish_slot(slot):
                published_count += 1
                
        self.logger.info(f"Published {published_count}/{len(new_slots)} slot notifications")
        return published_count
    
    def close(self):
        """Close Redis connection"""
        if self.client:
            self.client.close()
            self.logger.info("Redis connection closed")


def test_redis_publisher():
    """Test function for Redis publisher"""
    import os
    
    # Configure logging
    logging.basicConfig(level=logging.INFO)
    
    # Create publisher
    publisher = RedisPublisher()
    
    if not publisher.connect():
        print("Failed to connect to Redis")
        return
    
    # Test slot data
    test_slot = {
        'venueId': '64f8a123b456789012345678',
        'venueName': 'Victoria Park',
        'platform': 'courtside',
        'courtId': '1',
        'courtName': 'Court 1',
        'date': '2025-06-10',
        'startTime': '19:00',
        'endTime': '20:00',
        'price': 8.0,
        'isAvailable': True,
        'bookingUrl': 'https://tennistowerhamlets.com/book/courts/victoria-park#book',
        'scrapedAt': datetime.now().isoformat()
    }
    
    # Publish test slot
    if publisher.publish_slot(test_slot):
        print("✅ Test slot published successfully!")
    else:
        print("❌ Failed to publish test slot")
    
    publisher.close()


if __name__ == "__main__":
    test_redis_publisher() 