#!/usr/bin/env python3

"""
Scheduler for running the scraper orchestrator periodically.
This ensures continuous scraping at configured intervals.
"""

import asyncio
import logging
import os
import signal
import sys
import time
from datetime import datetime, timedelta
from typing import Optional

# Handle imports for both module and script execution
try:
    from .scrapers.scraper_orchestrator import ScraperOrchestrator
except ImportError:
    # Fallback for when running as script
    current_dir = os.path.dirname(os.path.abspath(__file__))
    sys.path.insert(0, current_dir)
    
    from scrapers.scraper_orchestrator import ScraperOrchestrator

class ScrapingScheduler:
    """Scheduler for periodic scraping operations"""
    
    def __init__(self):
        self.setup_logging()
        self.running = False
        self.next_run_time: Optional[datetime] = None
        
        # Get interval from environment (in minutes)
        self.interval_minutes = int(os.getenv("SCRAPER_INTERVAL_MINUTES", 
                                            os.getenv("SCRAPER_INTERVAL", 30)))
        
        self.logger.info(f"Scraping scheduler initialized with {self.interval_minutes}-minute intervals")
        
    def setup_logging(self):
        """Configure logging for the scheduler"""
        log_level = os.getenv('LOG_LEVEL', 'INFO').upper()
        
        # Configure format
        format_str = '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        
        logging.basicConfig(
            level=getattr(logging, log_level),
            format=format_str,
            handlers=[logging.StreamHandler()]
        )
        self.logger = logging.getLogger(__name__)
        
    def calculate_next_run_time(self) -> datetime:
        """Calculate the next scheduled run time"""
        return datetime.now() + timedelta(minutes=self.interval_minutes)
        
    async def run_scraping_session(self):
        """Run a single scraping session"""
        session_start = time.time()
        
        try:
            self.logger.info("🚀 Starting scheduled scraping session")
            
            # Create orchestrator and run scraping
            orchestrator = ScraperOrchestrator()
            results = await orchestrator.run_scraping_session()
            
            session_duration = time.time() - session_start
            
            # Log summary
            if results:
                total_venues = len(results)
                successful_venues = sum(1 for r in results if r.success)
                total_slots = sum(len(r.slots_found) for r in results)
                
                self.logger.info(f"✅ Scraping session completed in {session_duration:.2f}s")
                self.logger.info(f"📊 Results: {successful_venues}/{total_venues} venues successful, {total_slots} total slots")
            else:
                self.logger.warning("⚠️ No scraping results returned")
                
        except Exception as e:
            self.logger.error(f"❌ Scraping session failed: {e}")
            raise
            
    async def start_scheduler(self):
        """Start the periodic scraping scheduler"""
        self.running = True
        self.logger.info(f"🕐 Starting scraping scheduler (interval: {self.interval_minutes} minutes)")
        
        # Run initial scraping session immediately
        self.logger.info("🎯 Running initial scraping session...")
        try:
            await self.run_scraping_session()
        except Exception as e:
            self.logger.error(f"Initial scraping session failed: {e}")
        
        # Schedule next run
        self.next_run_time = self.calculate_next_run_time()
        self.logger.info(f"⏰ Next scraping session scheduled for: {self.next_run_time.strftime('%Y-%m-%d %H:%M:%S')}")
        
        # Main scheduler loop
        while self.running:
            try:
                current_time = datetime.now()
                
                # Check if it's time to run
                if current_time >= self.next_run_time:
                    await self.run_scraping_session()
                    
                    # Schedule next run
                    self.next_run_time = self.calculate_next_run_time()
                    self.logger.info(f"⏰ Next scraping session scheduled for: {self.next_run_time.strftime('%Y-%m-%d %H:%M:%S')}")
                
                # Sleep for 30 seconds before checking again
                await asyncio.sleep(30)
                
            except Exception as e:
                self.logger.error(f"Error in scheduler loop: {e}")
                # Wait a bit before retrying
                await asyncio.sleep(60)
                
    def stop_scheduler(self):
        """Stop the scheduler gracefully"""
        self.logger.info("🛑 Stopping scraping scheduler...")
        self.running = False
        
    def get_status(self) -> dict:
        """Get current scheduler status"""
        return {
            "running": self.running,
            "interval_minutes": self.interval_minutes,
            "next_run_time": self.next_run_time.isoformat() if self.next_run_time else None,
            "time_until_next_run": str(self.next_run_time - datetime.now()) if self.next_run_time else None
        }

# Global scheduler instance
scheduler = None

def signal_handler(signum, frame):
    """Handle shutdown signals gracefully"""
    global scheduler
    if scheduler:
        scheduler.stop_scheduler()
    sys.exit(0)

async def main():
    """Main entry point for the scheduler"""
    global scheduler
    
    # Set up signal handlers for graceful shutdown
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    scheduler = ScrapingScheduler()
    
    try:
        await scheduler.start_scheduler()
    except KeyboardInterrupt:
        scheduler.logger.info("Scheduler interrupted by user")
    except Exception as e:
        scheduler.logger.error(f"Scheduler failed: {e}")
        raise
    finally:
        if scheduler:
            scheduler.stop_scheduler()

if __name__ == "__main__":
    # Run the scheduler when executed as a script
    asyncio.run(main()) 