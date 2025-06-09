#!/usr/bin/env python3

"""
Test Venue Seeding Script

This script tests the venue seeding functionality without actually inserting data
into the main database. It creates a separate test database and tests all
seeding functions.
"""

import os
import unittest
import datetime
from seed_venues import (
    LTA_CLUBSPARK_VENUES,
    COURTSIDES_VENUES,
    connect_to_db,
    create_venue_collection,
    seed_venue
)
from pymongo import MongoClient, ASCENDING
from pymongo.errors import DuplicateKeyError, ServerSelectionTimeoutError
from dotenv import load_dotenv

# Load environment variables
load_dotenv()

class TestVenueSeeding(unittest.TestCase):
    
    @classmethod
    def setUpClass(cls):
        """Set up test database and connection"""
        # Use a test database
        cls.test_db_name = "tennis_booking_test"
        
        # Get MongoDB URI
        cls.mongo_uri = os.getenv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
        
        # Connect to the database
        cls.client = MongoClient(cls.mongo_uri)
        cls.db = cls.client[cls.test_db_name]
        
        # Create collection and indexes
        cls.collection = cls.db["venues"]
        
    @classmethod
    def tearDownClass(cls):
        """Clean up after all tests"""
        # Drop test database
        cls.client.drop_database(cls.test_db_name)
        cls.client.close()
    
    def setUp(self):
        """Run before each test"""
        # Clear collection before each test
        self.collection.delete_many({})
        
    def test_connect_to_db(self):
        """Test database connection"""
        # Use the actual function
        client, db = connect_to_db()
        
        # Check client and db
        self.assertIsNotNone(client)
        self.assertIsNotNone(db)
        
        # Clean up
        client.close()
    
    def test_create_venue_collection(self):
        """Test venue collection creation with indexes"""
        # Create collection and get indexes
        collection = create_venue_collection(self.db)
        indexes = [idx for idx in collection.list_indexes()]
        
        # Check index count (should be 4: _id, name, provider, is_active)
        self.assertEqual(len(indexes), 4)
        
        # Check for name index with unique constraint
        name_index_found = False
        for idx in indexes:
            if 'name' in idx['key']:
                name_index_found = True
                self.assertTrue(idx.get('unique', False))
                break
        
        self.assertTrue(name_index_found, "Name index not found")
    
    def test_seed_venue(self):
        """Test venue seeding"""
        # Use first LTA venue for testing
        test_venue = LTA_CLUBSPARK_VENUES[0]
        
        # Seed venue
        result_id = seed_venue(self.collection, test_venue)
        
        # Verify venue was inserted
        self.assertIsNotNone(result_id)
        
        # Retrieve venue from DB
        venue = self.collection.find_one({"_id": result_id})
        
        # Verify fields
        self.assertEqual(venue["name"], test_venue["name"])
        self.assertEqual(venue["provider"], test_venue["provider"])
        self.assertEqual(len(venue["courts"]), len(test_venue["courts"]))
        
        # Verify timestamps were added
        self.assertIn("created_at", venue)
        self.assertIn("updated_at", venue)
        
        # Test duplicate prevention
        duplicate_id = seed_venue(self.collection, test_venue)
        self.assertIsNone(duplicate_id)
        
    def test_all_venues_valid(self):
        """Test that all venue definitions are valid"""
        # Test LTA venues
        for venue in LTA_CLUBSPARK_VENUES:
            self._validate_venue_structure(venue)
        
        # Test Courtsides venues
        for venue in COURTSIDES_VENUES:
            self._validate_venue_structure(venue)
    
    def _validate_venue_structure(self, venue):
        """Helper method to validate venue structure"""
        # Check required fields
        self.assertIn("name", venue)
        self.assertIn("provider", venue)
        self.assertIn("url", venue)
        self.assertIn("location", venue)
        self.assertIn("courts", venue)
        self.assertIn("booking_window", venue)
        self.assertIn("scraper_config", venue)
        
        # Check location
        location = venue["location"]
        self.assertIn("address", location)
        self.assertIn("city", location)
        self.assertIn("post_code", location)
        
        # Check courts
        self.assertGreaterEqual(len(venue["courts"]), 1)
        court = venue["courts"][0]
        self.assertIn("id", court)
        self.assertIn("name", court)
        
        # Check scraper config
        config = venue["scraper_config"]
        self.assertIn("type", config)
        self.assertIn("requires_login", config)
        
if __name__ == "__main__":
    unittest.main() 