#!/usr/bin/env python3

import os
import time
import json
import hvac
import redis
import logging
from datetime import datetime
from dotenv import load_dotenv
from pymongo import MongoClient
from pymongo.errors import ConnectionError as PyMongoConnectionError

# Set up logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

class TennisScraper:
    def __init__(self):
        self.mongo_client = None
        self.redis_client = None
        self.vault_client = None
        self.initialize_connections()
        
    def initialize_connections(self):
        """Initialize connections to MongoDB, Redis, and Vault"""
        try:
            # MongoDB connection
            mongo_uri = os.getenv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
            mongo_db = os.getenv("MONGO_DB_NAME", "tennis_booking")
            self.mongo_client = MongoClient(mongo_uri)
            self.db = self.mongo_client[mongo_db]
            logger.info("MongoDB connection established")
            
            # Redis connection
            redis_host = os.getenv("REDIS_ADDR", "localhost:6379").split(":")[0]
            redis_port = int(os.getenv("REDIS_ADDR", "localhost:6379").split(":")[1])
            redis_password = os.getenv("REDIS_PASSWORD", "password")
            redis_db = int(os.getenv("REDIS_DB", "0"))
            self.redis_client = redis.Redis(
                host=redis_host,
                port=redis_port,
                password=redis_password,
                db=redis_db,
                decode_responses=True
            )
            self.redis_client.ping()  # Test connection
            logger.info("Redis connection established")
            
            # Vault connection
            vault_addr = os.getenv("VAULT_ADDR", "http://localhost:8200")
            vault_token = os.getenv("VAULT_TOKEN", "dev-token")
            self.vault_client = hvac.Client(url=vault_addr, token=vault_token)
            if not self.vault_client.is_authenticated():
                raise Exception("Vault authentication failed")
            logger.info("Vault connection established")
            
        except PyMongoConnectionError as e:
            logger.error(f"MongoDB connection error: {e}")
            raise
        except redis.ConnectionError as e:
            logger.error(f"Redis connection error: {e}")
            raise
        except Exception as e:
            logger.error(f"Error initializing connections: {e}")
            raise
    
    def get_credentials(self, platform):
        """Get credentials from Vault for the specified platform"""
        try:
            secret_path = f"secret/booking/{platform}"
            response = self.vault_client.secrets.kv.v2.read_secret_version(
                path=secret_path
            )
            return response["data"]["data"]
        except Exception as e:
            logger.error(f"Error getting credentials from Vault: {e}")
            return None
    
    def log_scraping_result(self, platform, status, details=None):
        """Log scraping results to MongoDB"""
        try:
            log_entry = {
                "platform": platform,
                "status": status,
                "timestamp": datetime.utcnow(),
                "details": details or {}
            }
            self.db.scraping_logs.insert_one(log_entry)
            logger.info(f"Logged scraping result for {platform}: {status}")
        except Exception as e:
            logger.error(f"Error logging scraping result: {e}")
    
    def scrape_lta(self):
        """Placeholder for LTA/Clubspark scraping logic"""
        logger.info("Scraping LTA/Clubspark...")
        # Placeholder for actual scraping code
        time.sleep(2)  # Simulate scraping work
        
        # Simulate finding some slots
        slots = [
            {"date": "2025-06-10", "time": "18:00", "court": "Court 1", "price": "£10.00"},
            {"date": "2025-06-11", "time": "19:00", "court": "Court 2", "price": "£12.00"}
        ]
        
        self.log_scraping_result("lta", "success", {"slots_found": len(slots), "slots": slots})
        return slots
    
    def scrape_courtsides(self):
        """Placeholder for courtsides.com scraping logic"""
        logger.info("Scraping courtsides.com...")
        # Placeholder for actual scraping code
        time.sleep(2)  # Simulate scraping work
        
        # Simulate finding some slots
        slots = [
            {"date": "2025-06-10", "time": "17:00", "court": "Court A", "price": "£8.00"},
            {"date": "2025-06-12", "time": "20:00", "court": "Court B", "price": "£9.00"}
        ]
        
        self.log_scraping_result("courtsides", "success", {"slots_found": len(slots), "slots": slots})
        return slots
    
    def process_redis_task(self, task_data):
        """Process a task received from Redis"""
        try:
            task = json.loads(task_data)
            platform = task.get("platform")
            
            if platform == "lta":
                slots = self.scrape_lta()
            elif platform == "courtsides":
                slots = self.scrape_courtsides()
            else:
                logger.warning(f"Unknown platform: {platform}")
                return
            
            # Publish results back to Redis
            result = {
                "task_id": task.get("task_id"),
                "platform": platform,
                "timestamp": datetime.utcnow().isoformat(),
                "slots": slots
            }
            self.redis_client.publish("scraping_results", json.dumps(result))
            
        except Exception as e:
            logger.error(f"Error processing task: {e}")
    
    def listen_for_tasks(self):
        """Listen for scraping tasks from Redis"""
        pubsub = self.redis_client.pubsub()
        pubsub.subscribe("scraping_tasks")
        
        logger.info("Listening for scraping tasks...")
        for message in pubsub.listen():
            if message["type"] == "message":
                self.process_redis_task(message["data"])
    
    def cleanup(self):
        """Clean up resources"""
        if self.mongo_client:
            self.mongo_client.close()
        if self.redis_client:
            self.redis_client.close()

def main():
    scraper = TennisScraper()
    try:
        # For testing, just run the scrapers directly
        # In production, would use listen_for_tasks()
        scraper.scrape_lta()
        scraper.scrape_courtsides()
        
        # Uncomment to listen for tasks from Redis
        # scraper.listen_for_tasks()
    except KeyboardInterrupt:
        logger.info("Scraper stopped by user")
    except Exception as e:
        logger.error(f"Error in main: {e}")
    finally:
        scraper.cleanup()

if __name__ == "__main__":
    main() 