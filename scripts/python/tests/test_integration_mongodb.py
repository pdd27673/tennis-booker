#!/usr/bin/env python3
"""
Integration test for LTA scraper + data standardization + MongoDB insertion
"""

import os
import logging
import time
from datetime import datetime, timedelta
from firecrawl import FirecrawlApp
from lta_clubspark_scraper import LTAClubsparkScraper
from data_standardizer import DataStandardizer


def test_full_integration():
    """Test the complete pipeline: scrape -> standardize -> store in MongoDB."""
    
    # Set up logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    logger = logging.getLogger(__name__)
    
    print("🚀 Testing Full Integration: LTA Scraper -> Data Standardizer -> MongoDB")
    print("=" * 80)
    
    # Get API key
    api_key = os.getenv("FIRECRAWL_API_KEY")
    if not api_key:
        print("❌ FIRECRAWL_API_KEY not found")
        return False
    
    # Initialize components
    firecrawl_client = FirecrawlApp(api_key=api_key)
    scraper = LTAClubsparkScraper(firecrawl_client=firecrawl_client)
    standardizer = DataStandardizer()
    
    # Test venue (using tomorrow's date)
    tomorrow = datetime.now() + timedelta(days=1)
    
    test_venue = {
        '_id': 'stratford_park_test',
        'name': 'Stratford Park (Test)',
        'provider': 'lta_clubspark',
        'url': 'https://stratford.newhamparkstennis.org.uk/Booking/BookByDate',
        'courts': [
            {'id': 'court_1', 'name': 'Court 1'},
            {'id': 'court_2', 'name': 'Court 2'}
        ]
    }
    
    print(f"📅 Testing for date: {tomorrow.date()}")
    print(f"🏟️  Venue: {test_venue['name']}")
    print(f"🔗 URL: {test_venue['url']}")
    
    try:
        # Step 1: Scrape venue data
        print("\n🔍 Step 1: Scraping venue data...")
        scrape_start_time = datetime.now()
        result = scraper.scrape_venue_availability(test_venue, target_date=tomorrow)
        scrape_end_time = datetime.now()
        scrape_duration_ms = int((scrape_end_time - scrape_start_time).total_seconds() * 1000)
        
        print(f"   ✅ Scraping completed in {scrape_duration_ms}ms")
        print(f"   📊 Status: {result['status']}")
        
        if result['status'] != 'success':
            print(f"   ❌ Scraping failed: {result.get('error', 'Unknown error')}")
            return False
        
        print(f"   🎾 Courts found: {len(result['court_availability'])}")
        print(f"   ⏰ Time slots: {len(result['time_slots'])}")
        
        # Step 2: Standardize data
        print("\n🔄 Step 2: Standardizing scraped data...")
        
        # Use a sample venue ObjectId (in real implementation, this would come from venues collection)
        sample_venue_id = "507f1f77bcf86cd799439011"
        
        standardized_log = standardizer.standardize_lta_data(
            result, 
            sample_venue_id, 
            scrape_start_time, 
            scrape_duration_ms
        )
        
        print(f"   ✅ Data standardized successfully")
        print(f"   📝 Venue: {standardized_log.venue_name}")
        print(f"   🏷️  Provider: {standardized_log.provider}")
        print(f"   ✅ Success: {standardized_log.success}")
        print(f"   🎾 Slots: {len(standardized_log.slots_found)}")
        print(f"   ⚠️  Errors: {len(standardized_log.errors)}")
        
        # Show sample slots
        available_slots = [slot for slot in standardized_log.slots_found if slot.available]
        booked_slots = [slot for slot in standardized_log.slots_found if not slot.available]
        
        print(f"   📊 Available: {len(available_slots)}, Booked: {len(booked_slots)}")
        
        if available_slots:
            print("   💰 Sample available slots:")
            for slot in available_slots[:3]:
                print(f"      - {slot.court} at {slot.time}: £{slot.price}")
        
        # Step 3: Connect to MongoDB
        print("\n🗄️  Step 3: Connecting to MongoDB...")
        
        if not standardizer.connect_to_mongodb():
            print("   ❌ Failed to connect to MongoDB")
            print("   ℹ️  Make sure MongoDB is running: docker-compose up -d")
            return False
        
        print("   ✅ Connected to MongoDB successfully")
        
        # Step 4: Insert data
        print("\n💾 Step 4: Inserting data into MongoDB...")
        
        success = standardizer.insert_scraping_log(standardized_log)
        
        if success:
            print("   ✅ Data inserted successfully into scraping_logs collection")
        else:
            print("   ❌ Failed to insert data into MongoDB")
            return False
        
        # Step 5: Verify insertion (optional)
        print("\n🔍 Step 5: Verifying data insertion...")
        
        # Query the database to verify the data was inserted
        collection = standardizer._db.scraping_logs
        recent_logs = list(collection.find().sort("created_at", -1).limit(1))
        
        if recent_logs:
            latest_log = recent_logs[0]
            print(f"   ✅ Found latest log: {latest_log.get('venue_name', 'Unknown')}")
            print(f"   📅 Scrape time: {latest_log.get('scrape_timestamp', 'Unknown')}")
            print(f"   🎾 Slots in DB: {len(latest_log.get('slots_found', []))}")
            print(f"   🏷️  Provider: {latest_log.get('provider', 'Unknown')}")
            print(f"   🆔 Log ID: {latest_log.get('_id', 'Unknown')}")
            print(f"   🔍 Available fields: {list(latest_log.keys())}")
        else:
            print("   ⚠️  No logs found in database")
        
        print("\n🎉 Integration test completed successfully!")
        print("=" * 80)
        
        return True
        
    except Exception as e:
        logger.error(f"Integration test failed: {e}")
        print(f"\n❌ Integration test failed: {e}")
        return False
        
    finally:
        # Clean up
        standardizer.close_connection()


def test_error_handling():
    """Test error handling in the integration pipeline."""
    
    print("\n🧪 Testing Error Handling...")
    
    standardizer = DataStandardizer()
    
    # Test with invalid scraper result
    invalid_result = {
        'status': 'error',
        'error': 'Test error message',
        'venue_name': 'Test Venue'
    }
    
    scrape_time = datetime.now()
    duration_ms = 1000
    
    log = standardizer.standardize_lta_data(
        invalid_result,
        "507f1f77bcf86cd799439011",
        scrape_time,
        duration_ms
    )
    
    print(f"   ✅ Error handling test: Success={log.success}, Errors={len(log.errors)}")
    
    if not log.success and len(log.errors) > 0:
        print("   ✅ Error handling working correctly")
        return True
    else:
        print("   ❌ Error handling not working as expected")
        return False


if __name__ == "__main__":
    # Run integration test
    success = test_full_integration()
    
    # Run error handling test
    error_test_success = test_error_handling()
    
    if success and error_test_success:
        print("\n🎉 All tests passed!")
        exit(0)
    else:
        print("\n❌ Some tests failed!")
        exit(1) 