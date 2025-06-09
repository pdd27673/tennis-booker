#!/usr/bin/env python3
"""
Seed test venues for the scraper orchestrator
"""

import pymongo
from pymongo import MongoClient
from datetime import datetime
from bson import ObjectId

def seed_test_venues():
    """Seed the database with test venues."""
    
    # Connect to MongoDB
    client = MongoClient("mongodb://admin:YOUR_PASSWORD@localhost:27017/")
    db = client.tennis_booking_bot
    venues_collection = db.venues
    
    # Test venues
    test_venues = [
        {
            "_id": ObjectId("507f1f77bcf86cd799439011"),
            "name": "Stratford Park Tennis Centre (Test)",
            "provider": "lta",
            "url": "https://stratford.newhamparkstennis.org.uk/Booking/BookByDate#?date=2025-06-09&role=guest",
            "location": {
                "address": "Stratford Park, London",
                "city": "London",
                "post_code": "E15 2AA"
            },
            "courts": [
                {"id": "1", "name": "Court 1", "surface": "hard", "indoor": False, "floodlights": True},
                {"id": "2", "name": "Court 2", "surface": "hard", "indoor": False, "floodlights": True}
            ],
            "booking_window": 7,
            "scraper_config": {
                "type": "clubspark",
                "requires_login": False,
                "retry_count": 3,
                "timeout_seconds": 30,
                "wait_after_load_ms": 3000,
                "use_headless_browser": True
            },
            "created_at": datetime.utcnow(),
            "updated_at": datetime.utcnow(),
            "scraping_interval": 30,
            "is_active": True
        }
    ]
    
    # Clear existing test venues
    venues_collection.delete_many({"name": {"$regex": "Test"}})
    
    # Insert test venues
    result = venues_collection.insert_many(test_venues)
    print(f"✅ Inserted {len(result.inserted_ids)} test venues")
    
    # Verify insertion
    venues_count = venues_collection.count_documents({"is_active": True})
    print(f"📊 Total active venues in database: {venues_count}")
    
    client.close()

if __name__ == "__main__":
    seed_test_venues() 