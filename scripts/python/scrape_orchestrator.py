#!/usr/bin/env python3
"""
Tennis Court Scraper Orchestrator

Main script that orchestrates the scraping process for tennis court venues.
Handles venue configuration loading, scraper invocation based on venue type,
data storage, error handling, and logging.

Usage:
    python scrape_orchestrator.py                    # Scrape all active venues
    python scrape_orchestrator.py --venue-id 123     # Scrape specific venue
    python scrape_orchestrator.py --venue-ids 1,2,3  # Scrape multiple venues
    python scrape_orchestrator.py --provider lta     # Scrape by provider type
"""

import argparse
import logging
import sys
import time
import traceback
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional, Tuple
import pymongo
from pymongo import MongoClient
from bson import ObjectId
import os

# Import our scraping modules
from lta_clubspark_scraper import LTAClubsparkScraper
from data_standardizer import DataStandardizer, ScrapingLog

# Import Firecrawl for enhanced scraping
try:
    from firecrawl import FirecrawlApp
    FIRECRAWL_AVAILABLE = True
except ImportError:
    print("Warning: Firecrawl not available. Falling back to requests-only mode.")
    FIRECRAWL_AVAILABLE = False


class VenueLoader:
    """Handles loading venue configurations from MongoDB."""
    
    def __init__(self, mongo_uri: str = "mongodb://admin:YOUR_PASSWORD@localhost:27017/", db_name: str = "tennis_booking_bot"):
        self.mongo_uri = mongo_uri
        self.db_name = db_name
        self.client = None
        self.db = None
        self.logger = logging.getLogger(self.__class__.__name__)
    
    def connect(self) -> bool:
        """Connect to MongoDB."""
        try:
            self.client = MongoClient(self.mongo_uri)
            self.db = self.client[self.db_name]
            # Test connection
            self.client.admin.command('ismaster')
            self.logger.info(f"Connected to MongoDB: {self.db_name}")
            return True
        except Exception as e:
            self.logger.error(f"Failed to connect to MongoDB: {e}")
            return False
    
    def disconnect(self):
        """Disconnect from MongoDB."""
        if self.client:
            self.client.close()
            self.logger.info("Disconnected from MongoDB")
    
    def get_venues(self, venue_ids: Optional[List[str]] = None, provider: Optional[str] = None, 
                   active_only: bool = True) -> List[Dict[str, Any]]:
        """
        Retrieve venue configurations from MongoDB.
        
        Args:
            venue_ids: List of specific venue IDs to retrieve
            provider: Filter by provider type ('lta', 'courtsides', etc.)
            active_only: Only return active venues
            
        Returns:
            List of venue dictionaries
        """
        if self.db is None:
            self.logger.error("Not connected to MongoDB")
            return []
        
        try:
            query = {}
            
            if venue_ids:
                # Convert string IDs to ObjectId if they're valid
                object_ids = []
                for vid in venue_ids:
                    try:
                        object_ids.append(ObjectId(vid))
                    except:
                        # If not a valid ObjectId, try matching by string ID or name
                        query["$or"] = query.get("$or", [])
                        query["$or"].extend([
                            {"_id": vid},
                            {"name": {"$regex": vid, "$options": "i"}}
                        ])
                
                if object_ids:
                    if "$or" in query:
                        query["$or"].append({"_id": {"$in": object_ids}})
                    else:
                        query["_id"] = {"$in": object_ids}
            
            if provider:
                query["provider"] = provider.lower()
            
            if active_only:
                query["is_active"] = True
            
            venues = list(self.db.venues.find(query))
            self.logger.info(f"Retrieved {len(venues)} venues from MongoDB")
            
            return venues
            
        except Exception as e:
            self.logger.error(f"Failed to retrieve venues: {e}")
            return []


class ScraperOrchestrator:
    """Main orchestrator for tennis court scraping operations."""
    
    def __init__(self):
        self.logger = self._setup_logging()
        self.venue_loader = VenueLoader()
        self.data_standardizer = DataStandardizer()
        self.scrapers = {}
        self.firecrawl_client = None
        
        # Initialize Firecrawl if available
        if FIRECRAWL_AVAILABLE:
            api_key = os.getenv('FIRECRAWL_API_KEY')
            if api_key:
                try:
                    self.firecrawl_client = FirecrawlApp(api_key=api_key)
                    self.logger.info("Firecrawl client initialized successfully")
                except Exception as e:
                    self.logger.warning(f"Failed to initialize Firecrawl: {e}")
        
        # Initialize scrapers
        self._initialize_scrapers()
    
    def _setup_logging(self) -> logging.Logger:
        """Set up logging configuration."""
        # Create logs directory if it doesn't exist
        os.makedirs('logs', exist_ok=True)
        
        # Configure logging
        log_format = '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        logging.basicConfig(
            level=logging.INFO,
            format=log_format,
            handlers=[
                logging.FileHandler(f'logs/scraper_orchestrator_{datetime.now().strftime("%Y%m%d")}.log'),
                logging.StreamHandler(sys.stdout)
            ]
        )
        
        return logging.getLogger(self.__class__.__name__)
    
    def _initialize_scrapers(self):
        """Initialize available scrapers."""
        try:
            # Initialize LTA/Clubspark scraper with Firecrawl client
            self.scrapers['lta'] = LTAClubsparkScraper(firecrawl_client=self.firecrawl_client)
            self.scrapers['clubspark'] = self.scrapers['lta']  # Alias
            self.logger.info("LTA/Clubspark scraper initialized")
            
        except Exception as e:
            self.logger.error(f"Failed to initialize scrapers: {e}")
    
    def connect_services(self) -> bool:
        """Connect to all required services."""
        try:
            # Connect to MongoDB for venue loading
            if not self.venue_loader.connect():
                return False
            
            # Connect data standardizer to MongoDB
            if not self.data_standardizer.connect_to_mongodb():
                return False
            
            self.logger.info("All services connected successfully")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to connect services: {e}")
            return False
    
    def disconnect_services(self):
        """Disconnect from all services."""
        try:
            self.venue_loader.disconnect()
            self.data_standardizer.close_connection()
            self.logger.info("All services disconnected")
        except Exception as e:
            self.logger.error(f"Error disconnecting services: {e}")
    
    def scrape_venue(self, venue: Dict[str, Any]) -> Tuple[bool, Optional[ScrapingLog]]:
        """
        Scrape a single venue.
        
        Args:
            venue: Venue configuration dictionary
            
        Returns:
            Tuple of (success, scraping_log)
        """
        venue_name = venue.get('name', 'Unknown')
        venue_id = str(venue.get('_id', ''))
        provider = venue.get('provider', '').lower()
        
        self.logger.info(f"Starting scrape for venue: {venue_name} (Provider: {provider})")
        
        start_time = datetime.utcnow()
        
        try:
            # Get the appropriate scraper
            if provider not in self.scrapers:
                raise ValueError(f"No scraper available for provider: {provider}")
            
            scraper = self.scrapers[provider]
            
            # Prepare scraping parameters
            venue_url = venue.get('url')
            if not venue_url:
                raise ValueError(f"No URL configured for venue: {venue_name}")
            
            # Determine target date (default to tomorrow)
            target_date = datetime.now() + timedelta(days=1)
            
            # Perform the scrape
            slots_data = scraper.scrape_venue_availability(
                venue_config={'_id': venue_id, 'name': venue_name, 'url': venue_url},
                target_date=target_date
            )
            
            # Standardize the data
            end_time = datetime.utcnow()
            duration_ms = int((end_time - start_time).total_seconds() * 1000)
            
            if provider in ['lta', 'clubspark']:
                scraping_log = self.data_standardizer.standardize_lta_data(
                    raw_scraper_result=slots_data,
                    venue_id=venue_id,
                    scrape_start_time=start_time,
                    scrape_duration_ms=duration_ms
                )
            else:
                # For other providers, use courtsides method as fallback
                scraping_log = self.data_standardizer.standardize_courtsides_data(
                    raw_scraper_result=slots_data,
                    venue_id=venue_id,
                    scrape_start_time=start_time,
                    scrape_duration_ms=duration_ms
                )
            
            # Store in MongoDB
            if scraping_log and self.data_standardizer.insert_scraping_log(scraping_log):
                duration = (end_time - start_time).total_seconds()
                
                self.logger.info(
                    f"✅ Successfully scraped {venue_name}: "
                    f"{len(scraping_log.slots_found)} slots found "
                    f"(Duration: {duration:.2f}s)"
                )
                
                return True, scraping_log
            else:
                raise Exception("Failed to store scraping results")
            
        except Exception as e:
            end_time = datetime.utcnow()
            duration = (end_time - start_time).total_seconds()
            
            self.logger.error(
                f"❌ Failed to scrape {venue_name}: {str(e)} "
                f"(Duration: {duration:.2f}s)"
            )
            
            # Create a failed scraping log
            failed_log = ScrapingLog(
                venue_id=venue_id,
                venue_name=venue_name,
                provider=provider,
                scrape_timestamp=start_time,
                slots_found=[],
                scrape_duration_ms=int(duration * 1000),
                errors=[str(e)],
                success=False
            )
            
            # Try to store the failed log
            try:
                self.data_standardizer.insert_scraping_log(failed_log)
            except Exception as store_error:
                self.logger.error(f"Failed to store error log: {store_error}")
            
            return False, failed_log
    
    def run_scraping_session(self, venue_ids: Optional[List[str]] = None, 
                           provider: Optional[str] = None) -> Dict[str, Any]:
        """
        Run a complete scraping session.
        
        Args:
            venue_ids: List of specific venue IDs to scrape
            provider: Filter by provider type
            
        Returns:
            Dictionary with session results
        """
        session_start = datetime.utcnow()
        self.logger.info("=" * 80)
        self.logger.info("🚀 Starting Tennis Court Scraping Session")
        self.logger.info("=" * 80)
        
        results = {
            'session_start': session_start,
            'venues_attempted': 0,
            'venues_successful': 0,
            'venues_failed': 0,
            'total_slots_found': 0,
            'venues_results': [],
            'errors': []
        }
        
        try:
            # Connect to services
            if not self.connect_services():
                raise Exception("Failed to connect to required services")
            
            # Load venues
            venues = self.venue_loader.get_venues(
                venue_ids=venue_ids,
                provider=provider,
                active_only=True
            )
            
            if not venues:
                self.logger.warning("No venues found matching criteria")
                return results
            
            self.logger.info(f"Found {len(venues)} venues to scrape")
            results['venues_attempted'] = len(venues)
            
            # Scrape each venue
            for venue in venues:
                venue_name = venue.get('name', 'Unknown')
                
                try:
                    success, scraping_log = self.scrape_venue(venue)
                    
                    venue_result = {
                        'venue_id': str(venue.get('_id', '')),
                        'venue_name': venue_name,
                        'provider': venue.get('provider', ''),
                        'success': success,
                        'slots_found': len(scraping_log.slots_found) if scraping_log else 0,
                        'error': scraping_log.errors[0] if scraping_log and scraping_log.errors else None
                    }
                    
                    results['venues_results'].append(venue_result)
                    
                    if success:
                        results['venues_successful'] += 1
                        if scraping_log:
                            results['total_slots_found'] += len(scraping_log.slots_found)
                    else:
                        results['venues_failed'] += 1
                        if scraping_log and scraping_log.errors:
                            results['errors'].append(f"{venue_name}: {scraping_log.errors[0]}")
                
                except Exception as e:
                    error_msg = f"Critical error scraping {venue_name}: {str(e)}"
                    self.logger.error(error_msg)
                    results['venues_failed'] += 1
                    results['errors'].append(error_msg)
                
                # Small delay between venues to be respectful
                time.sleep(2)
            
        except Exception as e:
            error_msg = f"Session failed: {str(e)}"
            self.logger.error(error_msg)
            results['errors'].append(error_msg)
        
        finally:
            self.disconnect_services()
        
        # Log session summary
        session_end = datetime.utcnow()
        session_duration = (session_end - session_start).total_seconds()
        
        results['session_end'] = session_end
        results['session_duration_seconds'] = session_duration
        
        self.logger.info("=" * 80)
        self.logger.info("📊 Scraping Session Summary")
        self.logger.info("=" * 80)
        self.logger.info(f"📅 Session Duration: {session_duration:.2f} seconds")
        self.logger.info(f"🎯 Venues Attempted: {results['venues_attempted']}")
        self.logger.info(f"✅ Venues Successful: {results['venues_successful']}")
        self.logger.info(f"❌ Venues Failed: {results['venues_failed']}")
        self.logger.info(f"🎾 Total Slots Found: {results['total_slots_found']}")
        
        if results['errors']:
            self.logger.info("🚨 Errors:")
            for error in results['errors']:
                self.logger.info(f"   - {error}")
        
        success_rate = (results['venues_successful'] / results['venues_attempted'] * 100) if results['venues_attempted'] > 0 else 0
        self.logger.info(f"📈 Success Rate: {success_rate:.1f}%")
        
        return results


def main():
    """Main entry point for the scraper orchestrator."""
    parser = argparse.ArgumentParser(
        description="Tennis Court Scraper Orchestrator",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s                              # Scrape all active venues
  %(prog)s --venue-id 507f1f77bcf86cd799439011  # Scrape specific venue by ID
  %(prog)s --venue-ids 507f1f77bcf86cd799439011,507f1f77bcf86cd799439012  # Multiple venues
  %(prog)s --provider lta               # Scrape only LTA venues
  %(prog)s --provider courtsides        # Scrape only Courtsides venues
  %(prog)s --debug                      # Enable debug logging
        """
    )
    
    parser.add_argument(
        '--venue-id',
        type=str,
        help='Scrape a specific venue by ID'
    )
    
    parser.add_argument(
        '--venue-ids',
        type=str,
        help='Scrape multiple venues by IDs (comma-separated)'
    )
    
    parser.add_argument(
        '--provider',
        type=str,
        choices=['lta', 'clubspark', 'courtsides'],
        help='Filter venues by provider type'
    )
    
    parser.add_argument(
        '--debug',
        action='store_true',
        help='Enable debug logging'
    )
    
    parser.add_argument(
        '--dry-run',
        action='store_true',
        help='Show what would be scraped without actually scraping'
    )
    
    args = parser.parse_args()
    
    # Parse venue IDs
    venue_ids = None
    if args.venue_id:
        venue_ids = [args.venue_id]
    elif args.venue_ids:
        venue_ids = [vid.strip() for vid in args.venue_ids.split(',')]
    
    # Set logging level
    if args.debug:
        logging.getLogger().setLevel(logging.DEBUG)
    
    try:
        orchestrator = ScraperOrchestrator()
        
        if args.dry_run:
            # Just show what would be scraped
            if orchestrator.connect_services():
                venues = orchestrator.venue_loader.get_venues(
                    venue_ids=venue_ids,
                    provider=args.provider,
                    active_only=True
                )
                print(f"\n🔍 Dry Run - Would scrape {len(venues)} venues:")
                for venue in venues:
                    print(f"  - {venue.get('name', 'Unknown')} ({venue.get('provider', 'Unknown')})")
                orchestrator.disconnect_services()
            return
        
        # Run the actual scraping session
        results = orchestrator.run_scraping_session(
            venue_ids=venue_ids,
            provider=args.provider
        )
        
        # Exit with error code if there were failures
        if results['venues_failed'] > 0:
            sys.exit(1)
        
    except KeyboardInterrupt:
        print("\n⚠️  Scraping interrupted by user")
        sys.exit(130)
    except Exception as e:
        print(f"💥 Critical error: {e}")
        logging.getLogger().error(f"Critical error: {e}", exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    main() 