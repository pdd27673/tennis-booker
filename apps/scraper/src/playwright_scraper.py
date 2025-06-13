#!/usr/bin/env python3

"""
🎾 Playwright-Based Tennis Court Scraper

Main CLI script for running the Playwright-based scraping system.
Uses platform-specific scrapers for Courtside and ClubSpark platforms.

Usage:
    python playwright_scraper.py --test victoria                    # Test single venue
    python playwright_scraper.py --venues victoria,stratford       # Scrape specific venues  
    python playwright_scraper.py --all                             # Scrape all venues
    python playwright_scraper.py --all --days 3                    # Scrape 3 days ahead
"""

import asyncio
import argparse
import logging
import os
import sys
from datetime import datetime

# Add the src directory to Python path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from scrapers.scraper_orchestrator import ScraperOrchestrator
from config import get_config

def setup_logging(log_level: str = None):
    """Setup logging configuration"""
    if log_level is None:
        config = get_config()
        log_level = config.get_log_level()
    
    logging.basicConfig(
        level=getattr(logging, log_level.upper()),
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler('playwright_scraper.log'),
            logging.StreamHandler()
        ]
    )

async def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='Playwright Tennis Court Scraper')
    
    # Scraping mode options
    parser.add_argument('--test', type=str, help='Test scrape a single venue by name')
    parser.add_argument('--venues', type=str, help='Comma-separated list of venue names to scrape')
    parser.add_argument('--all', action='store_true', help='Scrape all active venues')
    
    # Configuration options
    parser.add_argument('--days', type=int, help='Number of days ahead to scrape (uses config default if not specified)')
    parser.add_argument('--log-level', type=str, 
                       choices=['DEBUG', 'INFO', 'WARNING', 'ERROR'],
                       help='Logging level (uses config default if not specified)')
    
    # MongoDB connection options
    parser.add_argument('--mongo-uri', type=str, help='MongoDB connection URI')
    parser.add_argument('--db-name', type=str, help='MongoDB database name')
    
    args = parser.parse_args()
    
    # Setup logging
    setup_logging(args.log_level)
    logger = logging.getLogger(__name__)
    
    # Load configuration for defaults
    config = get_config()
    days_ahead = args.days if args.days is not None else config.get_scraper_days_ahead()
    
    # Validate arguments
    if not any([args.test, args.venues, args.all]):
        parser.error("Must specify one of: --test, --venues, or --all")
        
    # Create orchestrator
    orchestrator = ScraperOrchestrator(
        mongo_uri=args.mongo_uri,
        db_name=args.db_name
    )
    
    try:
        if args.test:
            # Test mode - single venue
            logger.info(f"🧪 Testing venue: {args.test}")
            result = await orchestrator.test_single_venue(args.test, days_ahead=days_ahead)
            
            if result:
                print(f"\n✅ Test Results for {args.test}:")
                print(f"   Platform: {result.platform}")
                print(f"   Success: {result.success}")
                print(f"   Slots found: {len(result.slots_found)}")
                print(f"   Duration: {result.duration_ms}ms")
                
                if result.errors:
                    print(f"   Errors: {len(result.errors)}")
                    for error in result.errors:
                        print(f"     - {error}")
                        
                if result.slots_found:
                    print(f"\n📅 Sample slots:")
                    for i, slot in enumerate(result.slots_found[:5]):  # Show first 5
                        print(f"     {slot.date} {slot.start_time}-{slot.end_time} "
                              f"{slot.court_name} £{slot.price or 'N/A'}")
                    if len(result.slots_found) > 5:
                        print(f"     ... and {len(result.slots_found) - 5} more")
            else:
                print(f"❌ Test failed for {args.test}")
                
        elif args.venues:
            # Specific venues mode
            venue_list = [v.strip() for v in args.venues.split(',')]
            logger.info(f"🎾 Scraping venues: {venue_list}")
            
            results = await orchestrator.run_scraping_session(
                venue_names=venue_list,
                target_dates=None  # Use default date range
            )
            
            print(f"\n✅ Scraping Results:")
            for result in results:
                status = "✅" if result.success else "❌"
                print(f"   {status} {result.venue_name}: {len(result.slots_found)} slots "
                      f"({result.duration_ms}ms)")
                
        elif args.all:
            # All venues mode
            logger.info(f"🎾 Scraping all active venues ({days_ahead} days ahead)")
            
            results = await orchestrator.run_scraping_session(
                venue_names=None,  # All venues
                target_dates=None  # Use default date range
            )
            
            print(f"\n✅ Scraping Session Complete:")
            total_slots = sum(len(r.slots_found) for r in results)
            successful = sum(1 for r in results if r.success)
            
            print(f"   Venues processed: {len(results)}")
            print(f"   Successful: {successful}")
            print(f"   Total slots found: {total_slots}")
            
            print(f"\n📊 Venue Results:")
            for result in results:
                status = "✅" if result.success else "❌"
                print(f"   {status} {result.venue_name} ({result.platform}): "
                      f"{len(result.slots_found)} slots")
                      
    except KeyboardInterrupt:
        logger.info("Scraping interrupted by user")
        print("\n⏹️  Scraping interrupted")
        
    except Exception as e:
        logger.error(f"Scraping failed: {e}")
        print(f"\n❌ Scraping failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    # Install Playwright browsers if needed
    try:
        import playwright
        asyncio.run(main())
    except ImportError:
        print("❌ Playwright not installed. Please run: pip install playwright")
        print("   Then install browsers: playwright install")
        sys.exit(1) 