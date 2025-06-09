#!/usr/bin/env python3
"""
Tennis Court Availability Scraper

This script monitors tennis court availability from various booking platforms using Firecrawl
and stores the data in MongoDB. It supports multiple venue types including Courtside/Tower Hamlets
and LTA/Clubspark venues.

Key features:
- Uses Firecrawl for robust web scraping with JavaScript support
- Integrates with HashiCorp Vault for secure credential management
- Stores structured data in MongoDB ScrapingLogs collection
- Handles multiple venue types with different scraping strategies
"""

import os
import sys
import logging
import argparse
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
import json

# Add the current directory to Python path for imports
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from firecrawl import FirecrawlApp
from pymongo import MongoClient
from pymongo.errors import ServerSelectionTimeoutError, DuplicateKeyError
from vault_client import VaultClient, get_platform_credentials
from hvac.exceptions import VaultError
from courtsides_scraper import CourtsidesVenueScraper
from lta_clubspark_scraper import LTAClubsparkScraper


class TennisCourtScraper:
    """
    Tennis court availability scraper using Firecrawl and MongoDB.
    
    This scraper handles multiple venue types and platforms:
    - Courtside/Tower Hamlets venues
    - LTA/Clubspark venues
    - Future platform integrations
    """
    
    def __init__(self, mongo_uri: Optional[str] = None, firecrawl_api_key: Optional[str] = None):
        """
        Initialize the scraper with MongoDB and Firecrawl connections.
        
        Args:
            mongo_uri: MongoDB connection string (defaults to env var)
            firecrawl_api_key: Firecrawl API key (defaults to env var)
        """
        self.logger = logging.getLogger(__name__)
        
        # Initialize MongoDB connection
        self.mongo_uri = mongo_uri or os.getenv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
        self.db_name = os.getenv("MONGO_DB_NAME", "tennis_booking")
        self.mongo_client = None
        self.db = None
        
        # Initialize Firecrawl (optional, for testing only)
        self.firecrawl_api_key = firecrawl_api_key or os.getenv("FIRECRAWL_API_KEY")
        self.firecrawl = None
        if self.firecrawl_api_key and self.firecrawl_api_key != "test_key_fallback":
            try:
                self.firecrawl = FirecrawlApp(api_key=self.firecrawl_api_key)
            except Exception as e:
                self.logger.warning(f"Failed to initialize Firecrawl: {e}")
                self.firecrawl = None
        
        # Initialize Vault client for credentials
        try:
            self.vault_client = VaultClient()
            self.logger.info("Vault client initialized successfully")
        except VaultError as e:
            self.logger.warning(f"Vault client initialization failed: {e}")
            self.vault_client = None
        
        # Initialize venue-specific scrapers
        self.courtsides_scraper = CourtsidesVenueScraper(firecrawl_client=None)  # Courtside works with requests
        self.lta_scraper = LTAClubsparkScraper(firecrawl_client=self.firecrawl)  # LTA needs Firecrawl for JS
        
        self.logger.info("Tennis court scraper initialized")
    
    def connect_to_mongodb(self) -> bool:
        """
        Establish connection to MongoDB.
        
        Returns:
            True if connection successful, False otherwise
        """
        try:
            self.logger.info(f"Connecting to MongoDB: {self.mongo_uri}")
            self.mongo_client = MongoClient(self.mongo_uri, serverSelectionTimeoutMS=5000)
            
            # Test the connection
            self.mongo_client.admin.command('ping')
            self.db = self.mongo_client[self.db_name]
            
            self.logger.info(f"Successfully connected to MongoDB database: {self.db_name}")
            return True
            
        except ServerSelectionTimeoutError as e:
            self.logger.error(f"Failed to connect to MongoDB: {e}")
            return False
        except Exception as e:
            self.logger.error(f"Unexpected error connecting to MongoDB: {e}")
            return False
    
    def get_venues_from_db(self, venue_ids: Optional[List[str]] = None) -> List[Dict[str, Any]]:
        """
        Retrieve venue configurations from MongoDB.
        
        Args:
            venue_ids: Optional list of specific venue IDs to retrieve
            
        Returns:
            List of venue configuration dictionaries
        """
        if self.db is None:
            self.logger.error("No database connection available")
            return []
        
        try:
            venues_collection = self.db.venues
            
            # Build query filter
            query = {"is_active": True}
            if venue_ids:
                query["_id"] = {"$in": venue_ids}
            
            venues = list(venues_collection.find(query))
            self.logger.info(f"Retrieved {len(venues)} active venues from database")
            
            return venues
            
        except Exception as e:
            self.logger.error(f"Failed to retrieve venues from database: {e}")
            return []
    
    def get_test_venues(self) -> List[Dict[str, Any]]:
        """
        Get test venue configurations for the Courtside venues.
        
        Returns:
            List of test venue dictionaries
        """
        return [
            {
                "_id": "ropemakers_field",
                "name": "Ropemakers Field",
                "provider": "courtsides",
                "url": "https://tennistowerhamlets.com/book/courts/ropemakers-field#book",
                "location": {
                    "address": "Ropemaker's Field, Limehouse",
                    "city": "London",
                    "post_code": "E14 7JE"
                },
                "courts": [
                    {"id": "court_1", "name": "Court 1"},
                    {"id": "court_2", "name": "Court 2"}
                ],
                "scraper_config": {
                    "type": "courtsides",
                    "requires_login": False,
                    "selector_mappings": {
                        "time_slots": "table tr",
                        "court_cells": "td",
                        "booking_status": "[class*='booked'], [class*='Reserved']"
                    }
                },
                "is_active": True
            },
            {
                "_id": "victoria_park",
                "name": "Victoria Park",
                "provider": "courtsides", 
                "url": "https://tennistowerhamlets.com/book/courts/victoria-park#book",
                "location": {
                    "address": "Victoria Park, Old Ford Road",
                    "city": "London",
                    "post_code": "E9 7DE"
                },
                "courts": [
                    {"id": "court_1", "name": "Court 1"},
                    {"id": "court_2", "name": "Court 2"},
                    {"id": "court_3", "name": "Court 3"},
                    {"id": "court_4", "name": "Court 4"}
                ],
                "scraper_config": {
                    "type": "courtsides",
                    "requires_login": False,
                    "selector_mappings": {
                        "time_slots": "table tr",
                        "court_cells": "td",
                        "booking_status": "[class*='booked'], [class*='Reserved']"
                    }
                },
                "is_active": True
            },
            {
                "_id": "wimbledon_park_lta",
                "name": "Wimbledon Park Tennis Club",
                "provider": "lta_clubspark",
                "url": "https://clubspark.lta.org.uk/WimbledonParkTennisClub/Booking/BookByDate#?date=2024-06-09",
                "location": {
                    "address": "Home Park Road",
                    "city": "London", 
                    "post_code": "SW19 7HX"
                },
                "courts": [
                    {"id": "court_1", "name": "Court 1"},
                    {"id": "court_2", "name": "Court 2"},
                    {"id": "court_3", "name": "Court 3"},
                    {"id": "court_4", "name": "Court 4"}
                ],
                "scraper_config": {
                    "type": "lta_clubspark",
                    "requires_login": True,
                    "selector_mappings": {
                        "time_slots": ".booking-slot",
                        "court_info": ".court-details",
                        "availability": ".availability-status"
                    }
                },
                "is_active": True
            },
            {
                "_id": "putney_tennis_lta",
                "name": "Putney Tennis Club", 
                "provider": "lta_clubspark",
                "url": "https://clubspark.lta.org.uk/PutneyTennisClub/Booking/BookByDate",
                "location": {
                    "address": "Putney Lower Common",
                    "city": "London",
                    "post_code": "SW15 1TW"
                },
                "courts": [
                    {"id": "court_1", "name": "Court 1"},
                    {"id": "court_2", "name": "Court 2"}
                ],
                "scraper_config": {
                    "type": "lta_clubspark",
                    "requires_login": True,
                    "selector_mappings": {
                        "time_slots": ".time-slot",
                        "court_info": ".court-name",
                        "booking_button": ".book-now"
                    }
                },
                "is_active": True
            },
            {
                "_id": "stratford_park_test",
                "name": "Stratford Park (Test)",
                "provider": "lta_clubspark",
                "url": "https://stratford.newhamparkstennis.org.uk/Booking/BookByDate",
                "location": {
                    "address": "Stratford Park, Newham",
                    "city": "London",
                    "post_code": "E15 1DA"
                },
                "courts": [
                    {"id": "court_1", "name": "Court 1"},
                    {"id": "court_2", "name": "Court 2"},
                    {"id": "court_3", "name": "Court 3"},
                    {"id": "court_4", "name": "Court 4"}
                ],
                "scraper_config": {
                    "type": "lta_clubspark",
                    "requires_login": False,
                    "selector_mappings": {
                        "sessions": ".resource-session",
                        "book_intervals": ".book-interval",
                        "availability": "[data-availability]"
                    }
                },
                "is_active": True
            }
        ]
    
    def test_firecrawl_connection(self) -> bool:
        """
        Test Firecrawl API connection.
        
        Returns:
            True if connection successful, False otherwise
        """
        try:
            # Test with a simple webpage
            test_url = "https://httpbin.org/html"
            result = self.firecrawl.scrape_url(test_url, params={'formats': ['markdown']})
            
            if result and 'markdown' in result:
                self.logger.info("Firecrawl API connection test successful")
                return True
            else:
                self.logger.error("Firecrawl API test failed - no content returned")
                return False
                
        except Exception as e:
            self.logger.error(f"Firecrawl API connection test failed: {e}")
            return False
    
    def close_connections(self):
        """Close all open connections."""
        if self.mongo_client:
            self.mongo_client.close()
            self.logger.info("MongoDB connection closed")
        
        if self.vault_client:
            self.vault_client.close()
            self.logger.info("Vault client closed")
    
    def scrape_venue(self, venue_config: Dict[str, Any], 
                    target_date: Optional[datetime] = None) -> Dict[str, Any]:
        """
        Scrape a single venue for court availability.
        
        Args:
            venue_config: Venue configuration dictionary
            target_date: Date to check (defaults to today)
            
        Returns:
            Scraping result dictionary
        """
        venue_provider = venue_config.get('provider', '').lower()
        venue_id = venue_config.get('_id')
        venue_name = venue_config.get('name')
        
        self.logger.info(f"Scraping venue {venue_name} (provider: {venue_provider})")
        
        try:
            if venue_provider in ['courtsides', 'tower_hamlets']:
                result = self.courtsides_scraper.scrape_venue_availability(venue_config, target_date)
            elif venue_provider in ['lta_clubspark', 'lta', 'clubspark']:
                result = self.lta_scraper.scrape_venue_availability(venue_config, target_date)
            else:
                self.logger.error(f"Unsupported venue provider: {venue_provider}")
                result = {
                    'venue_id': venue_id,
                    'status': 'error',
                    'error': f'Unsupported provider: {venue_provider}',
                    'scrape_date': datetime.now().isoformat()
                }
            
            return result
            
        except Exception as e:
            self.logger.error(f"Error scraping venue {venue_name}: {e}")
            return {
                'venue_id': venue_id,
                'status': 'error',
                'error': str(e),
                'scrape_date': datetime.now().isoformat()
            }
    
    def scrape_all_venues(self, venue_ids: Optional[List[str]] = None,
                         target_date: Optional[datetime] = None) -> List[Dict[str, Any]]:
        """
        Scrape all configured venues for court availability.
        
        Args:
            venue_ids: Optional list of specific venue IDs to scrape
            target_date: Date to check (defaults to today)
            
        Returns:
            List of scraping results
        """
        # Get venues from database or use test venues
        venues = []
        if self.db is not None:
            venues = self.get_venues_from_db(venue_ids)
        
        if not venues:
            self.logger.info("No venues found in database, using test venues")
            venues = self.get_test_venues()
            if venue_ids:
                venues = [v for v in venues if v['_id'] in venue_ids]
        
        results = []
        
        for venue in venues:
            venue_name = venue.get('name')
            self.logger.info(f"Processing venue: {venue_name}")
            
            try:
                result = self.scrape_venue(venue, target_date)
                results.append(result)
                
                # Store result in database if available
                if self.db is not None:
                    self.store_scraping_result(result)
                
            except Exception as e:
                self.logger.error(f"Failed to process venue {venue_name}: {e}")
                error_result = {
                    'venue_id': venue.get('_id'),
                    'status': 'error',
                    'error': str(e),
                    'scrape_date': datetime.now().isoformat()
                }
                results.append(error_result)
        
        self.logger.info(f"Completed scraping {len(results)} venues")
        return results
    
    def store_scraping_result(self, result: Dict[str, Any]) -> bool:
        """
        Store scraping result in MongoDB ScrapingLogs collection.
        
        Args:
            result: Scraping result dictionary
            
        Returns:
            True if stored successfully, False otherwise
        """
        if self.db is None:
            self.logger.error("No database connection available")
            return False
        
        try:
            scraping_logs = self.db.scraping_logs
            
            # Add some metadata
            result['_id'] = f"{result['venue_id']}_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
            result['created_at'] = datetime.now()
            
            # Insert the document
            scraping_logs.insert_one(result)
            
            self.logger.debug(f"Stored scraping result for venue {result['venue_id']}")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to store scraping result: {e}")
            return False


def setup_logging(level: str = "INFO"):
    """Configure logging for the scraper."""
    log_level = getattr(logging, level.upper(), logging.INFO)
    
    logging.basicConfig(
        level=log_level,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.StreamHandler(),
            logging.FileHandler('tennis_scraper.log')
        ]
    )


def main():
    """Main function for testing the scraper initialization and running scraping jobs."""
    parser = argparse.ArgumentParser(description="Tennis Court Availability Scraper")
    parser.add_argument("--log-level", default="INFO", help="Logging level")
    parser.add_argument("--test-init", action="store_true", help="Test initialization only")
    parser.add_argument("--scrape", action="store_true", help="Run scraping for all venues")
    parser.add_argument("--venue-ids", nargs="+", help="Specific venue IDs to scrape")
    parser.add_argument("--date", help="Target date (YYYY-MM-DD) to scrape, defaults to today")
    
    args = parser.parse_args()
    
    setup_logging(args.log_level)
    logger = logging.getLogger(__name__)
    
    logger.info("Starting Tennis Court Scraper")
    
    try:
        # Initialize scraper (Firecrawl is optional)
        scraper = TennisCourtScraper()
        
        # Test connections
        if not scraper.connect_to_mongodb():
            logger.error("Failed to connect to MongoDB")
            return 1
        
        # Test Firecrawl connection (optional, for testing only)
        if scraper.firecrawl:
            if not scraper.test_firecrawl_connection():
                logger.warning("Firecrawl API connection failed, will use requests fallback")
        else:
            logger.info("Firecrawl not configured, using requests-based scraping")
        
        # Test venue retrieval
        test_venues = scraper.get_test_venues()
        logger.info(f"Test venues loaded: {len(test_venues)}")
        for venue in test_venues:
            logger.info(f"  - {venue['name']}: {venue['url']}")
        
        if args.test_init:
            logger.info("Initialization test completed successfully")
            return 0
        
        # Parse target date if provided
        target_date = None
        if args.date:
            try:
                target_date = datetime.strptime(args.date, '%Y-%m-%d')
                logger.info(f"Using target date: {target_date.strftime('%Y-%m-%d')}")
            except ValueError:
                logger.error(f"Invalid date format: {args.date}. Use YYYY-MM-DD")
                return 1
        
        # Run scraping if requested
        if args.scrape:
            logger.info("Starting venue scraping...")
            
            # Use specific venue IDs if provided
            venue_ids = args.venue_ids
            if venue_ids:
                logger.info(f"Scraping specific venues: {venue_ids}")
            else:
                logger.info("Scraping all available venues")
            
            results = scraper.scrape_all_venues(venue_ids, target_date)
            
            # Summary report
            logger.info(f"\n{'='*60}")
            logger.info("SCRAPING SUMMARY")
            logger.info(f"{'='*60}")
            
            total_venues = len(results)
            successful = len([r for r in results if r.get('status') == 'success'])
            failed = total_venues - successful
            
            logger.info(f"Total venues processed: {total_venues}")
            logger.info(f"Successful: {successful}")
            logger.info(f"Failed: {failed}")
            
            for result in results:
                venue_name = result.get('venue_name', result.get('venue_id', 'Unknown'))
                status = result.get('status', 'unknown')
                
                if status == 'success':
                    court_count = len(result.get('court_availability', []))
                    logger.info(f"✅ {venue_name}: {court_count} court slots scraped")
                else:
                    error = result.get('error', 'Unknown error')
                    logger.info(f"❌ {venue_name}: {error}")
            
            logger.info(f"{'='*60}")
            return 0 if failed == 0 else 1
        
        logger.info("Scraper initialization completed successfully")
        logger.info("Use --scrape to run venue scraping or --test-init to test connections only")
        return 0
        
    except Exception as e:
        logger.error(f"Scraper failed: {e}")
        return 1
    
    finally:
        if 'scraper' in locals():
            scraper.close_connections()


if __name__ == "__main__":
    sys.exit(main()) 