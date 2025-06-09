#!/usr/bin/env python3
"""
Data standardization module for tennis court scraping data.
Transforms raw scraped data into MongoDB ScrapingLogs schema format.
"""

import logging
import re
from datetime import datetime, timedelta
from typing import Dict, Any, List, Optional
from dataclasses import dataclass, asdict
import pymongo
from pymongo import MongoClient
from bson import ObjectId


@dataclass
class Slot:
    """Represents a court slot in standardized format."""
    date: str          # Format: "YYYY-MM-DD"
    time: str          # Format: "HH:MM-HH:MM"
    court: str         # Court name
    price: float       # Price in pounds (0.0 if free or unknown)
    available: bool    # True if available for booking
    court_id: str = "" # Optional court identifier
    url: str = ""      # Optional direct booking URL


@dataclass
class ScrapingLog:
    """Represents a scraping log entry for MongoDB."""
    venue_id: str             # MongoDB ObjectId as string
    scrape_timestamp: datetime
    slots_found: List[Slot]
    scrape_duration_ms: int
    errors: List[str]
    success: bool
    venue_name: str
    provider: str
    raw_response: str = ""
    scraper_version: str = "1.0.0"
    user_agent: str = ""
    ip_address: str = ""
    run_id: str = ""
    created_at: Optional[datetime] = None


class DataStandardizer:
    """Handles standardization and storage of scraped tennis court data."""
    
    def __init__(self, mongo_uri: str = "mongodb://admin:YOUR_PASSWORD@localhost:27017", 
                 db_name: str = "tennis_booking"):
        """
        Initialize the data standardizer.
        
        Args:
            mongo_uri: MongoDB connection string
            db_name: Database name
        """
        self.logger = logging.getLogger(__name__)
        self.mongo_uri = mongo_uri
        self.db_name = db_name
        self._client = None
        self._db = None
        
    def connect_to_mongodb(self) -> bool:
        """
        Establish connection to MongoDB.
        
        Returns:
            True if connection successful, False otherwise
        """
        try:
            self._client = MongoClient(self.mongo_uri)
            self._db = self._client[self.db_name]
            
            # Test the connection
            self._client.admin.command('ping')
            self.logger.info("Successfully connected to MongoDB")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to connect to MongoDB: {e}")
            return False
    
    def standardize_lta_data(self, raw_scraper_result: Dict[str, Any], 
                           venue_id: str, scrape_start_time: datetime,
                           scrape_duration_ms: int) -> ScrapingLog:
        """
        Standardize LTA scraper output into ScrapingLog format.
        
        Args:
            raw_scraper_result: Output from LTAClubsparkScraper
            venue_id: MongoDB ObjectId as string
            scrape_start_time: When scraping started
            scrape_duration_ms: How long scraping took
            
        Returns:
            ScrapingLog object
        """
        slots = []
        errors = []
        
        try:
            if raw_scraper_result.get('status') == 'success':
                court_availability = raw_scraper_result.get('court_availability', [])
                target_date = raw_scraper_result.get('target_date', datetime.now().date().isoformat())
                
                for slot_data in court_availability:
                    try:
                        # Extract price
                        price = 0.0
                        if slot_data.get('price'):
                            price_str = str(slot_data['price']).replace('£', '').strip()
                            if price_str and price_str != '0':
                                price = float(price_str)
                        
                        # Normalize court name
                        court_name = self._normalize_court_name(slot_data.get('court_name', 'Unknown'))
                        
                        slot = Slot(
                            date=target_date,
                            time=slot_data.get('time_slot', 'Unknown'),
                            court=court_name,
                            price=price,
                            available=slot_data.get('available', False),
                            court_id=self._extract_court_id(court_name),
                            url=""  # LTA doesn't provide direct booking URLs
                        )
                        slots.append(slot)
                        
                    except Exception as e:
                        error_msg = f"Error processing slot {slot_data}: {e}"
                        self.logger.warning(error_msg)
                        errors.append(error_msg)
                
                success = True
                
            else:
                success = False
                error_msg = raw_scraper_result.get('error', 'Unknown scraping error')
                errors.append(error_msg)
                
        except Exception as e:
            success = False
            error_msg = f"Error standardizing LTA data: {e}"
            self.logger.error(error_msg)
            errors.append(error_msg)
        
        return ScrapingLog(
            venue_id=venue_id,
            scrape_timestamp=scrape_start_time,
            slots_found=slots,
            scrape_duration_ms=scrape_duration_ms,
            errors=errors,
            success=success,
            venue_name=raw_scraper_result.get('venue_name', 'Unknown LTA Venue'),
            provider='lta_clubspark',
            scraper_version="1.0.0",
            user_agent="LTA-Scraper/1.0",
            created_at=datetime.utcnow()
        )
    
    def standardize_courtsides_data(self, raw_scraper_result: Dict[str, Any], 
                                  venue_id: str, scrape_start_time: datetime,
                                  scrape_duration_ms: int) -> ScrapingLog:
        """
        Standardize courtsides scraper output into ScrapingLog format.
        
        Args:
            raw_scraper_result: Output from courtsides scraper
            venue_id: MongoDB ObjectId as string
            scrape_start_time: When scraping started
            scrape_duration_ms: How long scraping took
            
        Returns:
            ScrapingLog object
        """
        slots = []
        errors = []
        
        try:
            if raw_scraper_result.get('status') == 'success':
                # Process courtsides-specific data structure
                court_availability = raw_scraper_result.get('court_availability', [])
                target_date = raw_scraper_result.get('target_date', datetime.now().date().isoformat())
                
                for slot_data in court_availability:
                    try:
                        # Extract price (courtsides might have different format)
                        price = 0.0
                        if slot_data.get('price'):
                            price_str = str(slot_data['price']).replace('£', '').strip()
                            if price_str and price_str != '0':
                                price = float(price_str)
                        
                        slot = Slot(
                            date=target_date,
                            time=slot_data.get('time_slot', 'Unknown'),
                            court=slot_data.get('court_name', 'Unknown'),
                            price=price,
                            available=slot_data.get('available', False),
                            court_id=slot_data.get('court_id', ''),
                            url=slot_data.get('booking_url', '')
                        )
                        slots.append(slot)
                        
                    except Exception as e:
                        error_msg = f"Error processing courtsides slot {slot_data}: {e}"
                        self.logger.warning(error_msg)
                        errors.append(error_msg)
                
                success = True
                
            else:
                success = False
                error_msg = raw_scraper_result.get('error', 'Unknown courtsides scraping error')
                errors.append(error_msg)
                
        except Exception as e:
            success = False
            error_msg = f"Error standardizing courtsides data: {e}"
            self.logger.error(error_msg)
            errors.append(error_msg)
        
        return ScrapingLog(
            venue_id=venue_id,
            scrape_timestamp=scrape_start_time,
            slots_found=slots,
            scrape_duration_ms=scrape_duration_ms,
            errors=errors,
            success=success,
            venue_name=raw_scraper_result.get('venue_name', 'Unknown Courtsides Venue'),
            provider='courtsides',
            scraper_version="1.0.0",
            user_agent="Courtsides-Scraper/1.0",
            created_at=datetime.utcnow()
        )
    
    def _normalize_court_name(self, court_name: str) -> str:
        """
        Normalize court names to consistent format.
        
        Args:
            court_name: Raw court name from scraper
            
        Returns:
            Normalized court name
        """
        if not court_name or court_name == 'Unknown':
            return 'Unknown'
        
        # Remove extra descriptive text common in LTA sites
        # e.g., "Court 1Full, Outdoor, Incandescent Lighting, Tarmac" -> "Court 1"
        
        # First try to extract just "Court X" pattern
        court_match = re.search(r'Court\s*(\d+)', court_name, re.IGNORECASE)
        if court_match:
            return f"Court {court_match.group(1)}"
        
        # If no clear pattern, clean up the name
        cleaned = re.sub(r'[,\-]\s*(full|outdoor|indoor|lighting|tarmac|surface|court).*', '', court_name, flags=re.IGNORECASE)
        cleaned = cleaned.strip()
        
        return cleaned if cleaned else court_name
    
    def _extract_court_id(self, court_name: str) -> str:
        """
        Extract court ID from court name.
        
        Args:
            court_name: Normalized court name
            
        Returns:
            Court ID (e.g., "1", "2", etc.)
        """
        match = re.search(r'(\d+)', court_name)
        return match.group(1) if match else ""
    
    def insert_scraping_log(self, scraping_log: ScrapingLog) -> bool:
        """
        Insert a scraping log into MongoDB.
        
        Args:
            scraping_log: ScrapingLog object to insert
            
        Returns:
            True if insertion successful, False otherwise
        """
        if self._db is None:
            self.logger.error("Not connected to MongoDB")
            return False
        
        try:
            # Convert dataclass to dict
            log_dict = asdict(scraping_log)
            
            # Convert venue_id string to ObjectId
            log_dict['venue_id'] = ObjectId(scraping_log.venue_id)
            
            # Convert slots to list of dicts
            log_dict['slots_found'] = [asdict(slot) for slot in scraping_log.slots_found]
            
            # Ensure created_at is set
            if not log_dict['created_at']:
                log_dict['created_at'] = datetime.utcnow()
            
            # Insert into scraping_logs collection
            collection = self._db.scraping_logs
            result = collection.insert_one(log_dict)
            
            self.logger.info(f"Successfully inserted scraping log with ID: {result.inserted_id}")
            return True
            
        except Exception as e:
            self.logger.error(f"Error inserting scraping log: {e}")
            return False
    
    def process_and_store_scraping_result(self, raw_result: Dict[str, Any], 
                                        venue_id: str, provider: str,
                                        scrape_start_time: datetime,
                                        scrape_duration_ms: int) -> bool:
        """
        Main method to process and store scraping results.
        
        Args:
            raw_result: Raw scraper output
            venue_id: MongoDB ObjectId as string
            provider: Provider type ('lta_clubspark' or 'courtsides')
            scrape_start_time: When scraping started
            scrape_duration_ms: Scraping duration
            
        Returns:
            True if successful, False otherwise
        """
        try:
            # Standardize data based on provider
            if provider == 'lta_clubspark':
                scraping_log = self.standardize_lta_data(
                    raw_result, venue_id, scrape_start_time, scrape_duration_ms
                )
            elif provider == 'courtsides':
                scraping_log = self.standardize_courtsides_data(
                    raw_result, venue_id, scrape_start_time, scrape_duration_ms
                )
            else:
                self.logger.error(f"Unknown provider: {provider}")
                return False
            
            # Insert into MongoDB
            return self.insert_scraping_log(scraping_log)
            
        except Exception as e:
            self.logger.error(f"Error processing scraping result: {e}")
            return False
    
    def close_connection(self):
        """Close MongoDB connection."""
        if self._client:
            self._client.close()
            self.logger.info("MongoDB connection closed")


def main():
    """Test the data standardizer."""
    logging.basicConfig(level=logging.INFO)
    
    # Sample LTA scraper result (from our successful test)
    sample_lta_result = {
        'venue_id': 'stratford_park_test',
        'venue_name': 'Stratford Park (Test)',
        'status': 'success',
        'scrape_date': '2025-06-09T17:15:00.000Z',
        'target_date': '2025-06-10',
        'time_slots': ['08:00-09:00', '11:00-13:00', '15:00-17:00'],
        'court_availability': [
            {
                'court_name': 'Court 1',
                'time_slot': '08:00-09:00',
                'status': 'available',
                'available': True,
                'price': '6.00',
                'session_info': None
            },
            {
                'court_name': 'Court 2',
                'time_slot': '11:00-13:00',
                'status': 'booked',
                'available': False,
                'price': '0',
                'session_info': None
            }
        ],
        'total_slots': 2
    }
    
    # Test standardization
    standardizer = DataStandardizer()
    
    scrape_time = datetime.now()
    duration_ms = 1500
    
    log = standardizer.standardize_lta_data(
        sample_lta_result, 
        "507f1f77bcf86cd799439011",  # Sample ObjectId
        scrape_time,
        duration_ms
    )
    
    print("Standardized ScrapingLog:")
    print(f"- Venue ID: {log.venue_id}")
    print(f"- Venue Name: {log.venue_name}")
    print(f"- Provider: {log.provider}")
    print(f"- Success: {log.success}")
    print(f"- Slots Found: {len(log.slots_found)}")
    print(f"- Errors: {log.errors}")
    
    for i, slot in enumerate(log.slots_found):
        print(f"  Slot {i+1}: {slot.court} at {slot.time} - £{slot.price} ({'Available' if slot.available else 'Booked'})")


if __name__ == "__main__":
    main() 