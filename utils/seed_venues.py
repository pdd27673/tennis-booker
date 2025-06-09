#!/usr/bin/env python3

"""
Venue Collection Seeding Script

This script seeds the MongoDB venues collection with initial data for:
- LTA/Clubspark venues
- courtsides.com/tennistowerhamlets venues

The script ensures collections exist or creates them, and properly handles MongoDB connection.
"""

import os
import sys
import json
import datetime
from dotenv import load_dotenv
from pymongo import MongoClient, ASCENDING
from pymongo.errors import DuplicateKeyError, ServerSelectionTimeoutError

# Load environment variables
load_dotenv()

# MongoDB connection settings
MONGO_URI = os.getenv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
MONGO_DB_NAME = os.getenv("MONGO_DB_NAME", "tennis_booking")
VENUES_COLLECTION = "venues"

# Define venue data
LTA_CLUBSPARK_VENUES = [
    {
        "name": "Stratford Park Tennis",
        "provider": "lta_clubspark",
        "url": "https://stratford.newhamparkstennis.org.uk/Booking/BookByDate",
        "location": {
            "address": "Stratford Park, Broadway",
            "city": "London",
            "post_code": "E15 4BQ",
            "latitude": 51.5436,
            "longitude": -0.0032
        },
        "courts": [
            {
                "id": "stratford_court_1",
                "name": "Stratford Tennis Court 1",
                "surface": "hard",
                "indoor": False,
                "floodlights": True,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            },
            {
                "id": "stratford_court_2",
                "name": "Stratford Tennis Court 2",
                "surface": "hard",
                "indoor": False,
                "floodlights": True,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            },
            {
                "id": "stratford_court_3",
                "name": "Stratford Tennis Court 3",
                "surface": "hard",
                "indoor": False,
                "floodlights": True,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            },
            {
                "id": "stratford_court_4",
                "name": "Stratford Tennis Court 4",
                "surface": "hard",
                "indoor": False,
                "floodlights": True,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            }
        ],
        "booking_window": 7,
        "scraper_config": {
            "type": "clubspark",
            "requires_login": False,  # Guest booking available
            "retry_count": 3,
            "timeout_seconds": 30,
            "wait_after_load_ms": 3000,
            "use_headless_browser": True,
            "selector_mappings": {
                "date_picker": ".fc-datepicker",
                "court_slots": ".session-grid",
                "available_slot": ".available",
                "price_element": ".price",
                "time_element": ".time"
            },
            "navigation_steps": [
                "Navigate to URL",
                "Select date",
                "Extract available slots"
            ]
        },
        "scraping_interval": 30,
        "is_active": True
    }
]

COURTSIDES_VENUES = [
    {
        "name": "Victoria Park",
        "provider": "courtsides",
        "url": "https://tennistowerhamlets.com/book/courts/victoria-park",
        "location": {
            "address": "Victoria Park, Old Ford Road",
            "city": "London",
            "post_code": "E9 7DE",
            "latitude": 51.536710,
            "longitude": -0.034430
        },
        "courts": [
            {
                "id": "victoria_park_1",
                "name": "Victoria Park Court 1",
                "surface": "hard",
                "indoor": False,
                "floodlights": False,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            },
            {
                "id": "victoria_park_2",
                "name": "Victoria Park Court 2",
                "surface": "hard",
                "indoor": False,
                "floodlights": False,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            },
            {
                "id": "victoria_park_3",
                "name": "Victoria Park Court 3",
                "surface": "hard",
                "indoor": False,
                "floodlights": False,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            },
            {
                "id": "victoria_park_4",
                "name": "Victoria Park Court 4",
                "surface": "hard",
                "indoor": False,
                "floodlights": False,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            }
        ],
        "booking_window": 7,
        "scraper_config": {
            "type": "courtsides",
            "requires_login": False,  # Public booking interface
            "retry_count": 3,
            "timeout_seconds": 30,
            "wait_after_load_ms": 2000,
            "use_headless_browser": True,
            "selector_mappings": {
                "date_selector": ".date-button",
                "slots_container": "table",
                "court_slot": "td",
                "time_row": "tr",
                "booking_status": ".booking-status"
            },
            "navigation_steps": [
                "Navigate to URL",
                "Select date",
                "Extract available slots"
            ]
        },
        "scraping_interval": 30,
        "is_active": True
    },
    {
        "name": "Ropemakers Field",
        "provider": "courtsides",
        "url": "https://tennistowerhamlets.com/book/courts/ropemakers-field",
        "location": {
            "address": "Ropemakers Field, Limehouse",
            "city": "London",
            "post_code": "E14",
            "latitude": 51.5115,
            "longitude": -0.0235
        },
        "courts": [
            {
                "id": "ropemakers_1",
                "name": "Ropemakers Field Court 1",
                "surface": "hard",
                "indoor": False,
                "floodlights": False,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            },
            {
                "id": "ropemakers_2",
                "name": "Ropemakers Field Court 2",
                "surface": "hard",
                "indoor": False,
                "floodlights": False,
                "court_type": "doubles",
                "tags": ["outdoor", "park"]
            }
        ],
        "booking_window": 7,
        "scraper_config": {
            "type": "courtsides",
            "requires_login": False,
            "retry_count": 3,
            "timeout_seconds": 30,
            "wait_after_load_ms": 2000,
            "use_headless_browser": True,
            "selector_mappings": {
                "date_selector": ".date-button",
                "slots_container": "table",
                "court_slot": "td",
                "time_row": "tr",
                "booking_status": ".booking-status"
            },
            "navigation_steps": [
                "Navigate to URL",
                "Select date",
                "Extract available slots"
            ]
        },
        "scraping_interval": 30,
        "is_active": True
    }
]

def connect_to_db():
    """Connect to MongoDB and return db instance"""
    try:
        client = MongoClient(MONGO_URI, serverSelectionTimeoutMS=5000)
        
        # Test connection
        client.admin.command('ping')
        print("✅ MongoDB connection successful")
        
        # Get database
        db = client[MONGO_DB_NAME]
        return client, db
    
    except ServerSelectionTimeoutError as e:
        print(f"❌ Failed to connect to MongoDB: {e}")
        sys.exit(1)

def create_venue_collection(db):
    """Create venues collection if it doesn't exist and set up indexes"""
    # Check if collection exists
    if VENUES_COLLECTION in db.list_collection_names():
        print(f"✅ Collection '{VENUES_COLLECTION}' already exists")
    else:
        # Create collection
        db.create_collection(VENUES_COLLECTION)
        print(f"✅ Collection '{VENUES_COLLECTION}' created")
    
    # Set up indexes (similar to CreateIndexes in Go repository)
    collection = db[VENUES_COLLECTION]
    
    # Create a unique index on the name field
    collection.create_index([("name", ASCENDING)], unique=True)
    
    # Create an index on the provider field
    collection.create_index([("provider", ASCENDING)])
    
    # Create an index on the is_active field
    collection.create_index([("is_active", ASCENDING)])
    
    print("✅ Indexes created")
    
    return collection

def seed_venue(collection, venue_data):
    """Insert venue data with proper formatting"""
    try:
        # Add timestamps
        now = datetime.datetime.utcnow()
        venue_data["created_at"] = now
        venue_data["updated_at"] = now
        
        # Insert the venue
        result = collection.insert_one(venue_data)
        print(f"✅ Inserted venue '{venue_data['name']}' with ID: {result.inserted_id}")
        return result.inserted_id
    
    except DuplicateKeyError:
        print(f"⚠️ Venue '{venue_data['name']}' already exists, skipping")
        return None
    
    except Exception as e:
        print(f"❌ Error inserting venue '{venue_data['name']}': {e}")
        return None

def seed_venues():
    """Main function to seed venues collection"""
    print("\n🎾 TENNIS VENUE SEEDING SCRIPT 🎾")
    print("================================\n")
    
    # Connect to MongoDB
    client, db = connect_to_db()
    
    try:
        # Create collection and indexes
        collection = create_venue_collection(db)
        
        # Seed LTA/Clubspark venues
        print("\n📍 Seeding LTA/Clubspark venues...")
        for venue in LTA_CLUBSPARK_VENUES:
            seed_venue(collection, venue)
            
        # Seed Courtsides venues
        print("\n📍 Seeding Courtsides venues...")
        for venue in COURTSIDES_VENUES:
            seed_venue(collection, venue)
        
        # Print summary
        count = collection.count_documents({})
        print(f"\n✅ Seeding complete: {count} venues in database")
        
        # Print breakdown by provider
        lta_count = collection.count_documents({"provider": "lta_clubspark"})
        courtsides_count = collection.count_documents({"provider": "courtsides"})
        
        print(f"  - LTA/Clubspark venues: {lta_count}")
        print(f"  - Courtsides venues: {courtsides_count}")
        
    finally:
        # Close MongoDB connection
        client.close()
        print("\n✅ MongoDB connection closed")

if __name__ == "__main__":
    seed_venues() 