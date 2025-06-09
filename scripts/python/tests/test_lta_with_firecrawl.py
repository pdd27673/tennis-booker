#!/usr/bin/env python3
"""
Test the LTA Clubspark scraper with Firecrawl integration
"""

import os
import logging
from datetime import datetime, timedelta
from firecrawl import FirecrawlApp
from lta_clubspark_scraper import LTAClubsparkScraper

def test_lta_with_firecrawl():
    """Test the LTA scraper with Firecrawl."""
    
    # Set up logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    logger = logging.getLogger(__name__)
    
    print("🚀 Testing LTA Clubspark scraper with Firecrawl...")
    print("=" * 60)
    
    # Get API key
    api_key = os.getenv("FIRECRAWL_API_KEY")
    if not api_key:
        print("❌ FIRECRAWL_API_KEY not found")
        return
    
    # Initialize Firecrawl
    firecrawl_client = FirecrawlApp(api_key=api_key)
    
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
    
    # Initialize scraper with Firecrawl
    scraper = LTAClubsparkScraper(firecrawl_client=firecrawl_client)
    result = scraper.scrape_venue_availability(test_venue, target_date=tomorrow)
    
    print("\n📊 Results:")
    print(f"Status: {result['status']}")
    
    if result['status'] == 'success':
        print(f"✅ Courts found: {len(result['court_availability'])}")
        print(f"⏰ Time slots: {result['time_slots']}")
        
        if result['court_availability']:
            print(f"\n🎾 Court Availability (first 15):")
            
            # Group by status for better display
            available_slots = [slot for slot in result['court_availability'] if slot['available']]
            booked_slots = [slot for slot in result['court_availability'] if not slot['available']]
            
            print(f"\n✅ Available slots ({len(available_slots)}):")
            for slot in available_slots[:10]:
                price_info = f"(£{slot['price']})" if slot['price'] != '0' else ""
                print(f"  - {slot['court_name']} at {slot['time_slot']} {price_info}")
            
            print(f"\n❌ Booked/Reserved slots ({len(booked_slots)}):")
            for slot in booked_slots[:5]:
                session_info = f" - {slot['session_info']}" if slot['session_info'] else ""
                print(f"  - {slot['court_name']} at {slot['time_slot']}: {slot['status']}{session_info}")
                
        else:
            print("⚠️  No court availability data found")
            
    elif result['status'] == 'error':
        print(f"❌ Error: {result['error']}")
    
    return result

if __name__ == "__main__":
    test_lta_with_firecrawl() 