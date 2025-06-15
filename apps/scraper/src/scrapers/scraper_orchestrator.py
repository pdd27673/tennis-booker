#!/usr/bin/env python3

"""
Scraper orchestrator that coordinates platform-specific scrapers,
loads venues from MongoDB, and stores scraping results.
"""

import asyncio
import logging
import os
import sys
import time
from datetime import datetime
from typing import List, Dict, Any
from pymongo import MongoClient
from bson import ObjectId

# Handle imports for both module and script execution
try:
    # Try relative imports first (when run as module)
    from ..redis_publisher import RedisPublisher
    from .base_scraper import ScrapedSlot, ScrapingResult
    from .courtside_scraper import CourtsideScraper
    from .clubspark_scraper import ClubSparkScraper
except ImportError:
    # Fallback for when running as script - add parent directories to path
    current_dir = os.path.dirname(os.path.abspath(__file__))
    parent_dir = os.path.dirname(current_dir)  # src directory
    sys.path.insert(0, parent_dir)
    
    from redis_publisher import RedisPublisher
    from scrapers.base_scraper import ScrapedSlot, ScrapingResult
    from scrapers.courtside_scraper import CourtsideScraper
    from scrapers.clubspark_scraper import ClubSparkScraper

# Import Redis deduplicator
try:
    from ..deduplication.redis_deduplicator import RedisDeduplicator
except ImportError:
    from deduplication.redis_deduplicator import RedisDeduplicator

class ScraperOrchestrator:
    """Main orchestrator for tennis court scraping operations"""
    
    def __init__(self, mongo_uri: str = None, db_name: str = None):
        self.setup_logging()
        
        # MongoDB connection
        self.mongo_uri = mongo_uri or os.getenv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
        self.db_name = db_name or os.getenv("MONGO_DB_NAME", "tennis_booking")
        self.mongo_client = None
        self.db = None
        
        # Redis publisher
        self.redis_publisher = RedisPublisher()
        
        # Redis deduplicator for slot deduplication
        self.redis_deduplicator = RedisDeduplicator(
            redis_host=os.getenv("REDIS_HOST", "localhost"),
            redis_port=int(os.getenv("REDIS_PORT", "6379")),
            redis_password=os.getenv("REDIS_PASSWORD"),
            redis_db=int(os.getenv("REDIS_DEDUPE_DB", "1")),  # Use different DB for deduplication
            expiry_hours=int(os.getenv("REDIS_DEDUPE_EXPIRY_HOURS", "48"))
        )
        
        # Scraper registry - enable both platforms by default
        self.scrapers = {}
        if os.getenv('COURTSIDE_ENABLED', 'true').lower() == 'true':
            self.scrapers['courtside'] = CourtsideScraper
        if os.getenv('CLUBSPARK_ENABLED', 'true').lower() == 'true':
            self.scrapers['clubspark'] = ClubSparkScraper
        
    def setup_logging(self):
        """Configure logging"""
        log_level = os.getenv('LOG_LEVEL', 'INFO').upper()
        
        # Use simple console logging
        format_str = '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        
        logging.basicConfig(
            level=getattr(logging, log_level),
            format=format_str,
            handlers=[logging.StreamHandler()]
        )
        self.logger = logging.getLogger(__name__)
        
    def connect_mongodb(self):
        """Connect to MongoDB"""
        try:
            self.mongo_client = MongoClient(self.mongo_uri)
            self.db = self.mongo_client[self.db_name]
            
            # Test connection
            self.mongo_client.admin.command('ping')
            self.logger.info("Connected to MongoDB")
            
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to connect to MongoDB: {e}")
            return False
            
    def disconnect_mongodb(self):
        """Disconnect from MongoDB"""
        if self.mongo_client:
            self.mongo_client.close()
            self.logger.info("Disconnected from MongoDB")
            
    def connect_to_redis(self):
        """Connect to Redis for notifications"""
        try:
            self.redis_publisher.connect()
            self.logger.info("Connected to Redis for notifications")
            return True
        except Exception as e:
            self.logger.error(f"Failed to connect to Redis: {e}")
            return False
            
    def load_venues(self, venue_names: List[str] = None) -> List[Dict[str, Any]]:
        """Load venue configurations from MongoDB"""
        try:
            venues_collection = self.db.venues
            
            # Build query filter
            query = {}
            if venue_names:
                query["name"] = {"$in": venue_names}
                
            venues = list(venues_collection.find(query))
            
            # Convert ObjectId to string for JSON serialization
            for venue in venues:
                venue['_id'] = str(venue['_id'])
                
            self.logger.info(f"Loaded {len(venues)} venues from MongoDB")
            return venues
            
        except Exception as e:
            self.logger.error(f"Failed to load venues: {e}")
            return []
            
    async def scrape_venue(self, venue_config: Dict[str, Any], target_dates: List[str] = None) -> ScrapingResult:
        """Scrape a single venue using the appropriate platform scraper"""
        
        platform_type = venue_config['scraper_config']['type']
        venue_name = venue_config['name']
        
        if platform_type not in self.scrapers:
            error_msg = f"No scraper available for platform: {platform_type}"
            self.logger.error(error_msg)
            return ScrapingResult(
                venue_id=venue_config['_id'],
                venue_name=venue_name,
                platform=platform_type,
                success=False,
                slots_found=[],
                errors=[error_msg],
                duration_ms=0,
                scraped_at=datetime.now()
            )
            
        # Create platform-specific scraper
        scraper_class = self.scrapers[platform_type]
        scraper = scraper_class(venue_config)
        
        # Use default dates if none provided
        if not target_dates:
            days_ahead = int(os.getenv('SCRAPER_DAYS_AHEAD', '7'))
            target_dates = scraper.get_target_dates(days_ahead=days_ahead)
            
        self.logger.info(f"Starting scrape for {venue_name} ({platform_type}) - {len(target_dates)} dates")
        
        # Run the scraper
        result = await scraper.scrape_availability(target_dates)
        
        self.logger.info(f"Completed scrape for {venue_name}: {len(result.slots_found)} slots, "
                        f"success={result.success}, duration={result.duration_ms}ms")
        
        return result
        
    async def scrape_all_venues(self, venue_names: List[str] = None, 
                               target_dates: List[str] = None) -> List[ScrapingResult]:
        """Scrape all active venues"""
        
        if not self.mongo_client:
            self.connect_mongodb()
            
        venues = self.load_venues(venue_names)
        if not venues:
            self.logger.warning("No venues to scrape")
            return []
            
        results = []
        
        # Scrape venues sequentially to avoid overwhelming target sites
        for venue in venues:
            try:
                result = await self.scrape_venue(venue, target_dates)
                results.append(result)
                
                # Store results in MongoDB
                await self.store_scraping_result(result)
                
                # Rate limiting between venues
                interval = int(os.getenv('SCRAPER_INTERVAL', '600'))
                # Convert interval to a reasonable delay between venues (use 1/10th of interval)
                delay = max(1, interval // 10)
                await asyncio.sleep(delay)
                
            except Exception as e:
                error_msg = f"Error scraping venue {venue['name']}: {e}"
                self.logger.error(error_msg)
                
                # Create error result
                error_result = ScrapingResult(
                    venue_id=venue['_id'],
                    venue_name=venue['name'],
                    platform=venue['scraper_config']['type'],
                    success=False,
                    slots_found=[],
                    errors=[error_msg],
                    duration_ms=0,
                    scraped_at=datetime.now()
                )
                results.append(error_result)
                await self.store_scraping_result(error_result)
                
        return results
        
    async def store_scraping_result(self, result: ScrapingResult):
        """Store scraping result and slots in MongoDB, and publish new slots to Redis for notifications"""
        try:
            new_slots_for_notification = []
            duplicate_slots_count = 0
            
            # Store slots in slots collection
            if result.slots_found:
                slots_data = []
                for slot in result.slots_found:
                    slot_doc = {
                        "venue_id": ObjectId(slot.venue_id),
                        "venue_name": slot.venue_name,
                        "court_id": slot.court_id,
                        "court_name": slot.court_name,
                        "date": slot.date,
                        "start_time": slot.start_time,
                        "end_time": slot.end_time,
                        "price": slot.price,
                        "currency": slot.currency,
                        "available": slot.available,
                        "booking_url": slot.booking_url,
                        "scraped_at": slot.scraped_at,
                        "platform": result.platform
                    }
                    slots_data.append(slot_doc)
                
                # Use Redis deduplication to filter out recently seen slots
                self.logger.debug(f"Checking {len(slots_data)} slots for duplicates using Redis deduplication")
                new_slots, duplicate_slots = self.redis_deduplicator.check_multiple_slots(slots_data)
                duplicate_slots_count = len(duplicate_slots)
                
                if duplicate_slots_count > 0:
                    self.logger.info(f"🔄 Skipped {duplicate_slots_count} duplicate slots for {result.venue_name} (recently seen)")
                
                # Process only new slots (not seen in Redis cache)
                slots_collection = self.db.slots
                for slot_doc in new_slots:
                    filter_query = {
                        "venue_id": slot_doc["venue_id"],
                        "court_id": slot_doc["court_id"],
                        "date": slot_doc["date"],
                        "start_time": slot_doc["start_time"]
                    }
                    
                    # Check if this slot already exists in MongoDB (fallback check)
                    existing_slot = slots_collection.find_one(filter_query)
                    is_new_slot = existing_slot is None
                    
                    # Upsert the slot
                    slots_collection.replace_one(filter_query, slot_doc, upsert=True)
                    
                    # If it's a new available slot, add to notification queue
                    if is_new_slot and slot_doc["available"]:
                        notification_slot = {
                            'venueId': str(slot_doc["venue_id"]),
                            'venueName': slot_doc["venue_name"],
                            'platform': slot_doc["platform"],
                            'courtId': slot_doc["court_id"],
                            'courtName': slot_doc["court_name"],
                            'date': slot_doc["date"],
                            'startTime': slot_doc["start_time"],
                            'endTime': slot_doc["end_time"],
                            'price': float(slot_doc["price"]),
                            'isAvailable': slot_doc["available"],
                            'bookingUrl': slot_doc["booking_url"],
                            'scrapedAt': slot_doc["scraped_at"].strftime('%Y-%m-%dT%H:%M:%SZ')
                        }
                        new_slots_for_notification.append(notification_slot)
                        self.logger.info(f"🆕 New slot detected: {result.venue_name} - {slot_doc['court_name']} on {slot_doc['date']} at {slot_doc['start_time']}")
                    
                self.logger.info(f"Processed {len(new_slots)} new slots for {result.venue_name} (skipped {duplicate_slots_count} duplicates)")
                
            # Publish new slots to Redis for immediate notifications
            if new_slots_for_notification:
                self.logger.info(f"🔔 Found {len(new_slots_for_notification)} new slots to publish for {result.venue_name}")
                try:
                    if not self.redis_publisher.client:
                        self.logger.info("🔗 Connecting to Redis for notifications...")
                        self.connect_to_redis()
                        
                    published_count = self.redis_publisher.publish_new_slots(new_slots_for_notification)
                    self.logger.info(f"📧 Published {published_count} new slot notifications for {result.venue_name}")
                except Exception as e:
                    self.logger.error(f"Failed to publish notifications: {e}")
            else:
                self.logger.info(f"ℹ️ No new slots detected for {result.venue_name} - all slots already exist")
                
            # Store scraping log
            log_doc = {
                "venue_id": ObjectId(result.venue_id),
                "venue_name": result.venue_name,
                "platform": result.platform,
                "scrape_timestamp": result.scraped_at,
                "success": result.success,
                "slots_found": len(result.slots_found),
                "scrape_duration_ms": result.duration_ms,
                "errors": result.errors,
                "created_at": datetime.now()
            }
            
            logs_collection = self.db.scraping_logs
            logs_collection.insert_one(log_doc)
            
            self.logger.info(f"Stored scraping log for {result.venue_name}")
            
        except Exception as e:
            self.logger.error(f"Failed to store scraping result: {e}")
            
    def update_last_scrape_time(self):
        """Update the last scrape time in the database"""
        try:
            # Update or create system status document with last scrape time
            system_collection = self.db.system_status
            current_time = datetime.now()
            
            system_collection.update_one(
                {"_id": "scraper_status"},
                {
                    "$set": {
                        "last_scrape_time": current_time,
                        "updated_at": current_time
                    },
                    "$setOnInsert": {
                        "created_at": current_time
                    }
                },
                upsert=True
            )
            
            self.logger.info(f"Updated last scrape time: {current_time.strftime('%Y-%m-%d %H:%M:%S')}")
            
        except Exception as e:
            self.logger.error(f"Failed to update last scrape time: {e}")

    async def run_scraping_session(self, venue_names: List[str] = None, 
                                  target_dates: List[str] = None):
        """Run a complete scraping session"""
        session_start = time.time()
        
        try:
            self.logger.info("Starting scraping session")
            
            # Ensure MongoDB connection is established
            self.connect_mongodb()
            
            results = await self.scrape_all_venues(venue_names, target_dates)
            
            # Summary statistics
            total_venues = len(results)
            successful_venues = sum(1 for r in results if r.success)
            total_slots = sum(len(r.slots_found) for r in results)
            total_errors = sum(len(r.errors) for r in results)
            
            session_duration = time.time() - session_start
            
            # Get deduplication metrics
            dedupe_metrics = self.redis_deduplicator.get_metrics()
            
            self.logger.info(f"Scraping session completed in {session_duration:.2f}s")
            self.logger.info(f"Venues: {successful_venues}/{total_venues} successful")
            self.logger.info(f"Total slots found: {total_slots}")
            self.logger.info(f"Total errors: {total_errors}")
            self.logger.info(f"Deduplication: {dedupe_metrics['duplicates_found']}/{dedupe_metrics['total_checks']} duplicates found "
                           f"({dedupe_metrics['duplicate_rate']:.1%} duplicate rate)")
            
            # Update last scrape time at the end of successful session
            self.update_last_scrape_time()
            
            return results
            
        except Exception as e:
            self.logger.error(f"Scraping session failed: {e}")
            raise
        finally:
            self.disconnect_mongodb()
            self.redis_deduplicator.close()
            
    async def test_single_venue(self, venue_name: str, days_ahead: int = None):
        """Test scraping a single venue for debugging"""
        self.logger.info(f"Testing scrape for venue: {venue_name}")
        
        # Use config default if not specified
        if days_ahead is None:
            days_ahead = int(os.getenv('SCRAPER_DAYS_AHEAD', '7'))
        
        try:
            self.connect_mongodb()
            venues = self.load_venues([venue_name])
            
            if not venues:
                self.logger.error(f"Venue '{venue_name}' not found")
                return None
                
            venue = venues[0]
            
            # Create scraper and get target dates
            platform_type = venue['scraper_config']['type']
            scraper_class = self.scrapers[platform_type]
            scraper = scraper_class(venue)
            target_dates = scraper.get_target_dates(days_ahead=days_ahead)
            
            result = await self.scrape_venue(venue, target_dates)
            
            # Store result
            await self.store_scraping_result(result)
            
            self.logger.info(f"Test completed: {len(result.slots_found)} slots found")
            return result
            
        except Exception as e:
            self.logger.error(f"Test failed: {e}")
            return None
        finally:
            self.disconnect_mongodb()
            self.redis_deduplicator.close()


async def main():
    """Main entry point for the scraper orchestrator"""
    orchestrator = ScraperOrchestrator()
    
    # Run a scraping session for all venues
    try:
        await orchestrator.run_scraping_session()
    except KeyboardInterrupt:
        orchestrator.logger.info("Scraping interrupted by user")
    except Exception as e:
        orchestrator.logger.error(f"Scraping failed: {e}")
        raise


if __name__ == "__main__":
    # Run the scraper when executed as a script
    asyncio.run(main())

 