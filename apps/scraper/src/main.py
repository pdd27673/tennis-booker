#!/usr/bin/env python3

"""
Main entry point for the Tennis Booker scraper service.

This module provides the primary interface for running the scraper,
either as a one-time run or in scheduled mode.
"""

import asyncio
import argparse
import sys
import logging
import os
from pathlib import Path

# Add src to path for imports
current_dir = Path(__file__).parent
sys.path.insert(0, str(current_dir))

from config.config import load_config
from scrapers.scraper_orchestrator import ScraperOrchestrator
from scheduler import ScrapingScheduler


def setup_logging():
    """Setup basic logging configuration"""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
        handlers=[logging.StreamHandler()]
    )


async def run_once(venue_names=None, days_ahead=None):
    """Run scraper once and exit"""
    logger = logging.getLogger(__name__)
    logger.info("🚀 Starting one-time scraping run")
    
    try:
        orchestrator = ScraperOrchestrator()
        
        if venue_names and len(venue_names) == 1:
            # Single venue test
            result = await orchestrator.test_single_venue(venue_names[0], days_ahead)
            if result:
                logger.info(f"✅ Single venue test completed: {len(result.slots_found)} slots found")
            else:
                logger.error("❌ Single venue test failed")
                return False
        else:
            # Full scraping session
            results = await orchestrator.run_scraping_session(venue_names)
            if results:
                successful = sum(1 for r in results if r.success)
                total_slots = sum(len(r.slots_found) for r in results)
                logger.info(f"✅ Scraping completed: {successful}/{len(results)} venues successful, {total_slots} total slots")
            else:
                logger.error("❌ Scraping session failed")
                return False
                
        return True
        
    except Exception as e:
        logger.error(f"❌ Scraping failed: {e}")
        return False


async def run_scheduler():
    """Run scraper in scheduled mode"""
    logger = logging.getLogger(__name__)
    logger.info("🕐 Starting scraper in scheduled mode")
    
    try:
        scheduler = ScrapingScheduler()
        await scheduler.start_scheduler()
        
    except KeyboardInterrupt:
        logger.info("⏹️ Scheduler stopped by user")
    except Exception as e:
        logger.error(f"❌ Scheduler failed: {e}")
        raise


def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(
        description='Tennis Court Scraper',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python -m src.main                          # Run once, scrape all venues
  python -m src.main --venue "Victoria Park"  # Test single venue
  python -m src.main --schedule               # Run in scheduled mode
  python -m src.main --days-ahead 10          # Scrape 10 days ahead
        """
    )
    
    parser.add_argument(
        '--schedule', 
        action='store_true',
        help='Run in scheduled mode (continuous scraping)'
    )
    
    parser.add_argument(
        '--venue',
        type=str,
        help='Scrape specific venue only (for testing)'
    )
    
    parser.add_argument(
        '--days-ahead',
        type=int,
        help='Number of days ahead to scrape (overrides config)'
    )
    
    parser.add_argument(
        '--log-level',
        choices=['DEBUG', 'INFO', 'WARNING', 'ERROR'],
        default='INFO',
        help='Set logging level'
    )
    
    args = parser.parse_args()
    
    # Setup logging
    setup_logging()
    logger = logging.getLogger(__name__)
    
    # Set log level
    logging.getLogger().setLevel(getattr(logging, args.log_level))
    
    # Load and validate configuration
    try:
        config = load_config()
        logger.info(f"📋 Configuration loaded: {config.get_app_name()} v{config.get_app_version()}")
        logger.info(f"🌍 Environment: {config.get_environment()}")
        
    except Exception as e:
        logger.error(f"❌ Failed to load configuration: {e}")
        sys.exit(1)
    
    # Determine venues to scrape
    venue_names = [args.venue] if args.venue else None
    
    # Run the appropriate mode
    try:
        if args.schedule:
            # Scheduled mode
            asyncio.run(run_scheduler())
        else:
            # One-time run
            success = asyncio.run(run_once(venue_names, args.days_ahead))
            sys.exit(0 if success else 1)
            
    except KeyboardInterrupt:
        logger.info("🛑 Interrupted by user")
        sys.exit(0)
    except Exception as e:
        logger.error(f"❌ Unexpected error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()