#!/usr/bin/env python3

"""
Scraper orchestrator that coordinates platform-specific scrapers,
loads venues from MongoDB, and stores scraping results.
"""

import asyncio
import logging
import os
import time
from datetime import datetime
from typing import List, Dict, Any, Optional
from pymongo import MongoClient
from bson import ObjectId
from redis_publisher import RedisPublisher
import copy

from .base_scraper import ScrapedSlot, ScrapingResult
from .courtside_scraper import CourtsideScraper
from .clubspark_scraper import ClubSparkScraper

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
        
        # Scraper registry
        self.scrapers = {
            'courtside': CourtsideScraper,
            'clubspark': ClubSparkScraper
        }
        
    def setup_logging(self):
        """Configure logging"""
        log_level = os.getenv("LOG_LEVEL", "INFO")
        logging.basicConfig(
            level=getattr(logging, log_level),
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler('scraper_orchestrator.log'),
                logging.StreamHandler()
            ]
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
            target_dates = scraper.get_target_dates(days_ahead=7)
            
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
                await asyncio.sleep(3)
                
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
        """Store scraping result and slots in MongoDB"""
        try:
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
                    
                # Upsert slots (replace existing slots for same venue/date)
                slots_collection = self.db.slots
                for slot_doc in slots_data:
                    filter_query = {
                        "venue_id": slot_doc["venue_id"],
                        "court_id": slot_doc["court_id"],
                        "date": slot_doc["date"],
                        "start_time": slot_doc["start_time"]
                    }
                    slots_collection.replace_one(filter_query, slot_doc, upsert=True)
                    
                self.logger.info(f"Stored {len(slots_data)} slots for {result.venue_name}")
                
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
            
    async def run_scraping_session(self, venue_names: List[str] = None, 
                                  target_dates: List[str] = None):
        """Run a complete scraping session"""
        session_start = time.time()
        
        try:
            self.logger.info("Starting scraping session")
            
            results = await self.scrape_all_venues(venue_names, target_dates)
            
            # Summary statistics
            total_venues = len(results)
            successful_venues = sum(1 for r in results if r.success)
            total_slots = sum(len(r.slots_found) for r in results)
            total_errors = sum(len(r.errors) for r in results)
            
            session_duration = time.time() - session_start
            
            self.logger.info(f"Scraping session completed in {session_duration:.2f}s")
            self.logger.info(f"Venues: {successful_venues}/{total_venues} successful")
            self.logger.info(f"Total slots found: {total_slots}")
            self.logger.info(f"Total errors: {total_errors}")
            
            return results
            
        except Exception as e:
            self.logger.error(f"Scraping session failed: {e}")
            raise
        finally:
            self.disconnect_mongodb()
            
    async def test_single_venue(self, venue_name: str, days_ahead: int = 1):
        """Test scraping a single venue for debugging"""
        self.logger.info(f"Testing scrape for venue: {venue_name}")
        
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

    def store_slots(self, venue_id, venue_name, platform, slots_data):
        """Store slots in MongoDB and send notifications for new slots"""
        if not slots_data:
            self.logger.info(f"No slots to store for {venue_name}")
            return
        
        collection = self.db.slots
        stored_count = 0
        new_slots = []  # Track new slots for notifications
        
        for slot in slots_data:
            try:
                # Create slot document
                slot_doc = {
                    'venueId': ObjectId(venue_id),
                    'venueName': venue_name,
                    'platform': platform,
                    'courtId': slot.get('courtId', ''),
                    'courtName': slot.get('courtName', ''),
                    'date': slot.get('date', ''),
                    'startTime': slot.get('startTime', ''),
                    'endTime': slot.get('endTime', ''),
                    'price': float(slot.get('price', 0)),
                    'isAvailable': slot.get('isAvailable', True),
                    'bookingUrl': slot.get('bookingUrl', ''),
                    'sourceUrl': slot.get('sourceUrl', ''),
                    'scrapedAt': datetime.utcnow()
                }
                
                # Create unique identifier for upsert
                unique_filter = {
                    'venueId': ObjectId(venue_id),
                    'courtId': slot.get('courtId', ''),
                    'date': slot.get('date', ''),
                    'startTime': slot.get('startTime', '')
                }
                
                # Check if this is a new slot
                existing_slot = collection.find_one(unique_filter)
                is_new_slot = existing_slot is None
                
                # Upsert the slot
                result = collection.update_one(
                    unique_filter,
                    {'$set': slot_doc},
                    upsert=True
                )
                
                if result.upserted_id or result.modified_count > 0:
                    stored_count += 1
                    
                    # If this is a new available slot, add to notifications
                    if is_new_slot and slot.get('isAvailable', True):
                        notification_slot = {
                            'venueId': str(venue_id),
                            'venueName': venue_name,
                            'platform': platform,
                            'courtId': slot.get('courtId', ''),
                            'courtName': slot.get('courtName', ''),
                            'date': slot.get('date', ''),
                            'startTime': slot.get('startTime', ''),
                            'endTime': slot.get('endTime', ''),
                            'price': float(slot.get('price', 0)),
                            'isAvailable': slot.get('isAvailable', True),
                            'bookingUrl': slot.get('bookingUrl', ''),
                            'scrapedAt': datetime.utcnow().isoformat()
                        }
                        new_slots.append(notification_slot)
                        self.logger.info(f"New slot detected: {venue_name} - {slot.get('courtName', '')} on {slot.get('date', '')} at {slot.get('startTime', '')}")
                        
            except Exception as e:
                self.logger.error(f"Error storing slot {slot}: {e}")
                
        self.logger.info(f"Stored {stored_count} slots for {venue_name}")
        
        # Send notifications for new slots
        if new_slots:
            try:
                if not self.redis_publisher.client:
                    self.connect_to_redis()
                    
                published_count = self.redis_publisher.publish_new_slots(new_slots)
                self.logger.info(f"Published {published_count} new slot notifications for {venue_name}")
            except Exception as e:
                self.logger.error(f"Failed to publish notifications: {e}")

    def publish_new_slots_for_notifications(self, venue_id, venue_name, slots, platform):
        """Publish new slots to Redis for notification processing"""
        try:
            for slot in slots:
                # Only publish available slots
                if slot.get('isAvailable', False):
                    notification_data = {
                        'venueId': venue_id,
                        'venueName': venue_name,
                        'platform': platform,
                        'courtId': slot.get('courtId', ''),
                        'courtName': slot.get('courtName', ''),
                        'date': slot.get('date', ''),
                        'startTime': slot.get('startTime', ''),
                        'endTime': slot.get('endTime', ''),
                        'price': slot.get('price', 0.0),
                        'isAvailable': True,
                        'bookingUrl': slot.get('bookingUrl', ''),
                        'scrapedAt': slot.get('scrapedAt', datetime.utcnow()).isoformat() if isinstance(slot.get('scrapedAt'), datetime) else slot.get('scrapedAt', datetime.utcnow().isoformat())
                    }
                    
                    # Publish to Redis queue
                    success = self.redis_publisher.publish_slot(notification_data)
                    if success:
                        self.logger.info(f"📧 Published slot notification: {venue_name} - {slot.get('courtName')} at {slot.get('startTime')}")
                    else:
                        self.logger.warning(f"⚠️ Failed to publish slot notification for {venue_name}")
                        
        except Exception as e:
            self.logger.error(f"Error publishing notifications for {venue_name}: {e}")

    def store_venue_slots(self, venue_id, venue_name, slots):
        """Store scraped slots in MongoDB and return count of new slots"""
        if not slots:
            return 0
        
        try:
            slots_collection = self.db.slots
            new_slots_count = 0
            
            for slot in slots:
                # Add venue information to slot
                slot_doc = copy.deepcopy(slot)
                slot_doc['venueId'] = venue_id
                slot_doc['venueName'] = venue_name
                
                # Create unique identifier for upsert
                query = {
                    'venueId': venue_id,
                    'courtId': slot.get('courtId'),
                    'date': slot.get('date'),
                    'startTime': slot.get('startTime')
                }
                
                # Check if this is a new slot
                existing_slot = slots_collection.find_one(query)
                is_new_slot = existing_slot is None
                
                # Upsert the slot
                slots_collection.update_one(
                    query,
                    {'$set': slot_doc},
                    upsert=True
                )
                
                if is_new_slot:
                    new_slots_count += 1
            
            self.logger.info(f"📊 Stored {len(slots)} slots for {venue_name} ({new_slots_count} new)")
            return new_slots_count
            
        except Exception as e:
            self.logger.error(f"Error storing slots for {venue_name}: {e}")
            return 0 