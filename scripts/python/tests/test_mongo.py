#!/usr/bin/env python3

import os
from dotenv import load_dotenv
from pymongo import MongoClient

# Load environment variables
load_dotenv()

def test_mongodb_connection():
    """Test connection to MongoDB"""
    try:
        # Get MongoDB connection details from environment variables
        mongo_uri = os.getenv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
        mongo_db = os.getenv("MONGO_DB_NAME", "tennis_booking")
        
        # Connect to MongoDB
        client = MongoClient(mongo_uri)
        db = client[mongo_db]
        
        # Ping the database to verify connection
        client.admin.command('ping')
        
        print("MongoDB connection successful!")
        print(f"Connected to database: {mongo_db}")
        
        # List all collections
        collections = db.list_collection_names()
        print(f"Collections: {collections}")
        
        # Create a test collection and insert a document
        test_collection = db.test_collection
        test_document = {
            "name": "Test Document",
            "type": "Connection Test",
            "status": "success"
        }
        result = test_collection.insert_one(test_document)
        print(f"Inserted document with ID: {result.inserted_id}")
        
        # Retrieve the document
        retrieved = test_collection.find_one({"name": "Test Document"})
        print(f"Retrieved document: {retrieved}")
        
        # Clean up - delete the test document
        test_collection.delete_one({"_id": result.inserted_id})
        print("Test document deleted")
        
        return True
    except Exception as e:
        print(f"MongoDB connection failed: {e}")
        return False
    finally:
        if 'client' in locals():
            client.close()
            print("MongoDB connection closed")

if __name__ == "__main__":
    test_mongodb_connection() 