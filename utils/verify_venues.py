#!/usr/bin/env python3

"""
Venue Verification Script

This script verifies that venues were successfully inserted into the database
and displays their details.
"""

import os
from dotenv import load_dotenv
from pymongo import MongoClient
import json
from bson import json_util

# Load environment variables
load_dotenv()

# MongoDB connection settings
MONGO_URI = os.getenv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
MONGO_DB_NAME = os.getenv("MONGO_DB_NAME", "tennis_booking")
VENUES_COLLECTION = "venues"

def parse_json(data):
    """Parse MongoDB BSON to JSON"""
    return json.loads(json_util.dumps(data))

def verify_venues():
    """Retrieve and display venues from the database"""
    print("\n🎾 TENNIS VENUE VERIFICATION 🎾")
    print("===============================\n")
    
    try:
        # Connect to MongoDB
        client = MongoClient(MONGO_URI, serverSelectionTimeoutMS=5000)
        
        # Test connection
        client.admin.command('ping')
        print("✅ MongoDB connection successful")
        
        # Get database and collection
        db = client[MONGO_DB_NAME]
        collection = db[VENUES_COLLECTION]
        
        # Count venues
        venue_count = collection.count_documents({})
        print(f"Found {venue_count} venues in the database\n")
        
        if venue_count == 0:
            print("❌ No venues found. Please run the seeding script first.")
            return
        
        # Get venues by provider
        lta_venues = list(collection.find({"provider": "lta_clubspark"}))
        courtsides_venues = list(collection.find({"provider": "courtsides"}))
        
        # Display venues
        print(f"📍 LTA/Clubspark Venues ({len(lta_venues)}):")
        for venue in lta_venues:
            print(f"  - {venue['name']}")
            print(f"    URL: {venue['url']}")
            print(f"    Location: {venue['location']['address']}, {venue['location']['city']}, {venue['location']['post_code']}")
            print(f"    Courts: {len(venue['courts'])}")
            print(f"    Active: {venue['is_active']}\n")
        
        print(f"📍 Courtsides Venues ({len(courtsides_venues)}):")
        for venue in courtsides_venues:
            print(f"  - {venue['name']}")
            print(f"    URL: {venue['url']}")
            print(f"    Location: {venue['location']['address']}, {venue['location']['city']}, {venue['location']['post_code']}")
            print(f"    Courts: {len(venue['courts'])}")
            print(f"    Active: {venue['is_active']}\n")
    
    except Exception as e:
        print(f"❌ Error: {e}")
    finally:
        # Close MongoDB connection
        client.close()
        print("✅ MongoDB connection closed")

if __name__ == "__main__":
    verify_venues() 