#!/usr/bin/env python3
"""
Redis Worker for Tennis Court Scraping Tasks

This worker listens for scraping tasks from Redis queues and processes them
using the scrape orchestrator.
"""

import json
import logging
import os
import signal
import sys
import time
from datetime import datetime
from typing import Dict, Optional

import redis
from dotenv import load_dotenv

from scrape_orchestrator import ScraperOrchestrator


class RedisWorker:
    """Redis worker that processes tennis court scraping tasks"""
    
    def __init__(self, redis_url: str = "redis://localhost:6379", queue_name: str = "tennis_scraping_tasks"):
        self.redis_url = redis_url
        self.queue_name = queue_name
        self.high_priority_queue = f"{queue_name}:high"
        self.running = False
        
        # Initialize Redis connection
        self.redis_client = redis.from_url(redis_url, decode_responses=True)
        
        # Initialize scraper orchestrator
        self.orchestrator = ScraperOrchestrator()
        
        # Setup logging
        self.logger = logging.getLogger(__name__)
        handler = logging.StreamHandler()
        formatter = logging.Formatter(
            '[%(asctime)s] [WORKER] %(levelname)s: %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )
        handler.setFormatter(formatter)
        self.logger.addHandler(handler)
        self.logger.setLevel(logging.INFO)
        
        # Statistics
        self.stats = {
            "tasks_processed": 0,
            "tasks_failed": 0,
            "start_time": None,
            "last_task_time": None
        }
        
        # Setup graceful shutdown
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)
    
    def _signal_handler(self, signum, frame):
        """Handle shutdown signals gracefully"""
        self.logger.info(f"Received signal {signum}, shutting down gracefully...")
        self.running = False
    
    def start(self):
        """Start the worker and begin processing tasks"""
        self.logger.info("🚀 Starting Redis Worker for Tennis Court Scraping")
        self.logger.info(f"📡 Listening to queues: {self.high_priority_queue} (high), {self.queue_name} (normal)")
        
        self.running = True
        self.stats["start_time"] = datetime.now()
        
        try:
            # Test Redis connection
            self.redis_client.ping()
            self.logger.info("✅ Connected to Redis")
            
            # Test orchestrator initialization
            self.orchestrator.close_connection()  # Test and close
            self.logger.info("✅ Scraper orchestrator initialized")
            
        except Exception as e:
            self.logger.error(f"❌ Failed to initialize worker: {e}")
            return
        
        # Main processing loop
        while self.running:
            try:
                # Process tasks with priority (high priority first)
                task_processed = (
                    self._process_queue(self.high_priority_queue, "HIGH") or
                    self._process_queue(self.queue_name, "NORMAL")
                )
                
                if not task_processed:
                    # No tasks available, sleep briefly
                    time.sleep(1)
                    
            except Exception as e:
                self.logger.error(f"❌ Error in main loop: {e}")
                time.sleep(5)  # Wait before retrying
        
        self.logger.info("🛑 Worker stopped")
        self._print_final_stats()
    
    def _process_queue(self, queue_name: str, priority: str) -> bool:
        """
        Process a single task from the specified queue
        
        Returns:
            bool: True if a task was processed, False if queue was empty
        """
        try:
            # Try to get a task (non-blocking with timeout)
            task_data = self.redis_client.brpop(queue_name, timeout=1)
            
            if not task_data:
                return False  # No task available
            
            _, task_json = task_data
            task = json.loads(task_json)
            
            self.logger.info(f"📋 [{priority}] Processing task: {task.get('task_id')} for venue {task.get('venue_name')}")
            
            # Process the task
            success = self._process_scraping_task(task)
            
            # Update statistics
            self.stats["last_task_time"] = datetime.now()
            if success:
                self.stats["tasks_processed"] += 1
                self.logger.info(f"✅ [{priority}] Task completed successfully: {task.get('task_id')}")
            else:
                self.stats["tasks_failed"] += 1
                self.logger.error(f"❌ [{priority}] Task failed: {task.get('task_id')}")
            
            return True
            
        except json.JSONDecodeError as e:
            self.logger.error(f"❌ Invalid JSON in task: {e}")
            return True  # Count as processed (bad task)
        except redis.ConnectionError as e:
            self.logger.error(f"❌ Redis connection error: {e}")
            time.sleep(5)
            return False
        except Exception as e:
            self.logger.error(f"❌ Unexpected error processing queue {queue_name}: {e}")
            return False
    
    def _process_scraping_task(self, task: Dict) -> bool:
        """
        Process a single scraping task
        
        Args:
            task: Task dictionary containing venue information
            
        Returns:
            bool: True if successful, False otherwise
        """
        try:
            venue_id = task.get('venue_id')
            venue_name = task.get('venue_name')
            venue_url = task.get('venue_url')
            provider = task.get('provider')
            task_id = task.get('task_id')
            
            if not all([venue_id, venue_name, venue_url, provider]):
                self.logger.error(f"❌ Invalid task data: missing required fields")
                return False
            
            # Record task start
            start_time = time.time()
            self.logger.info(f"🎯 Starting scrape for venue: {venue_name} ({provider})")
            
            # Execute scraping using orchestrator
            # For now, we'll use a simple single-venue scrape
            success, results = self._execute_scraping(venue_id, venue_name, venue_url, provider)
            
            # Record results
            duration = time.time() - start_time
            
            if success:
                slots_found = results.get('slots_found', 0) if results else 0
                self.logger.info(f"✅ Scraping completed: {venue_name} - {slots_found} slots found in {duration:.2f}s")
                return True
            else:
                error_msg = results.get('error', 'Unknown error') if results else 'Unknown error'
                self.logger.error(f"❌ Scraping failed: {venue_name} - {error_msg} (took {duration:.2f}s)")
                return False
                
        except Exception as e:
            self.logger.error(f"❌ Exception during task processing: {e}")
            return False
    
    def _execute_scraping(self, venue_id: str, venue_name: str, venue_url: str, provider: str) -> tuple[bool, Optional[Dict]]:
        """
        Execute the actual scraping operation
        
        Returns:
            tuple: (success: bool, results: Optional[Dict])
        """
        try:
            # Use the orchestrator to handle the scraping
            venue_config = {
                'id': venue_id,
                'name': venue_name,
                'url': venue_url,
                'provider': provider,
                'is_active': True
            }
            
            # Use the orchestrator's scrape_venue method
            success, scraping_log = self.orchestrator.scrape_venue(venue_config)
            
            if success and scraping_log:
                return True, {
                    'slots_found': len(scraping_log.slots) if scraping_log.slots else 0,
                    'venue_id': venue_id,
                    'scraped_at': datetime.now().isoformat(),
                    'scraping_log_id': str(scraping_log.id) if hasattr(scraping_log, 'id') else None
                }
            else:
                error_msg = 'Unknown error'
                if scraping_log and hasattr(scraping_log, 'errors') and scraping_log.errors:
                    error_msg = scraping_log.errors[0]
                return False, {'error': error_msg}
                
        except Exception as e:
            return False, {'error': str(e)}
    
    def _print_final_stats(self):
        """Print final worker statistics"""
        if self.stats["start_time"]:
            runtime = datetime.now() - self.stats["start_time"]
            total_tasks = self.stats["tasks_processed"] + self.stats["tasks_failed"]
            success_rate = (self.stats["tasks_processed"] / total_tasks * 100) if total_tasks > 0 else 0
            
            self.logger.info("📊 Worker Statistics:")
            self.logger.info(f"   • Runtime: {runtime}")
            self.logger.info(f"   • Tasks processed: {self.stats['tasks_processed']}")
            self.logger.info(f"   • Tasks failed: {self.stats['tasks_failed']}")
            self.logger.info(f"   • Success rate: {success_rate:.1f}%")
            if self.stats["last_task_time"]:
                self.logger.info(f"   • Last task: {self.stats['last_task_time']}")
    
    def get_queue_stats(self) -> Dict[str, int]:
        """Get current queue statistics"""
        try:
            return {
                "normal_queue": self.redis_client.llen(self.queue_name),
                "high_priority_queue": self.redis_client.llen(self.high_priority_queue),
                "total_pending": self.redis_client.llen(self.queue_name) + self.redis_client.llen(self.high_priority_queue)
            }
        except Exception as e:
            self.logger.error(f"❌ Failed to get queue stats: {e}")
            return {}


def main():
    """Main entry point for the Redis worker"""
    # Load environment variables
    load_dotenv()
    
    # Configuration
    redis_url = os.getenv('REDIS_URL', 'redis://localhost:6379')
    queue_name = os.getenv('TASK_QUEUE_NAME', 'tennis_scraping_tasks')
    
    # Check if Redis password is set
    redis_password = os.getenv('REDIS_PASSWORD')
    if redis_password and redis_password.strip():
        # Parse the URL and add password
        if '://' in redis_url and '@' not in redis_url:
            protocol, rest = redis_url.split('://', 1)
            redis_url = f"{protocol}://:{redis_password}@{rest}"
    
    print("🚀 Tennis Court Scraping Redis Worker")
    print(f"📡 Redis URL: {redis_url}")
    print(f"📋 Queue Name: {queue_name}")
    print("Press Ctrl+C to stop the worker...")
    print("-" * 50)
    
    # Create and start worker
    worker = RedisWorker(redis_url=redis_url, queue_name=queue_name)
    
    try:
        worker.start()
    except KeyboardInterrupt:
        print("\n🛑 Interrupted by user")
    except Exception as e:
        print(f"❌ Worker failed: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main() 