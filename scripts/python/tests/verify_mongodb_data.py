#!/usr/bin/env python3
"""
Verify what data was inserted into MongoDB
"""

import logging
from datetime import datetime
from data_standardizer import DataStandardizer


def verify_mongodb_data():
    """Check what data is in the MongoDB scraping_logs collection."""
    
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger(__name__)
    
    print("🔍 Verifying MongoDB Data")
    print("=" * 50)
    
    # Connect to MongoDB
    standardizer = DataStandardizer()
    
    if not standardizer.connect_to_mongodb():
        print("❌ Failed to connect to MongoDB")
        return
    
    print("✅ Connected to MongoDB")
    
    # Get the collection
    collection = standardizer._db.scraping_logs
    
    # Count total documents
    total_count = collection.count_documents({})
    print(f"📊 Total documents in scraping_logs: {total_count}")
    
    # Get recent documents
    recent_docs = list(collection.find().sort("created_at", -1).limit(5))
    
    print(f"\n📋 Recent {len(recent_docs)} documents:")
    
    for i, doc in enumerate(recent_docs, 1):
        print(f"\n{i}. Document ID: {doc.get('_id')}")
        print(f"   Venue: {doc.get('venue_name', 'Unknown')}")
        print(f"   Provider: {doc.get('provider', 'Unknown')}")
        print(f"   Created: {doc.get('created_at', 'Unknown')}")
        print(f"   Success: {doc.get('success', 'Unknown')}")
        print(f"   Slots: {len(doc.get('slots_found', []))}")
        print(f"   Fields: {list(doc.keys())}")
        
        # If this looks like our standardized format, show more details
        if 'slots_found' in doc and doc.get('provider') == 'lta_clubspark':
            slots = doc.get('slots_found', [])
            if slots:
                print(f"   Sample slot: {slots[0]}")
    
    # Look specifically for our test venue
    test_docs = list(collection.find({"venue_name": "Stratford Park (Test)"}))
    
    if test_docs:
        print(f"\n🎯 Found {len(test_docs)} documents for 'Stratford Park (Test)':")
        for doc in test_docs:
            print(f"   - ID: {doc.get('_id')}")
            print(f"   - Created: {doc.get('created_at')}")
            print(f"   - Provider: {doc.get('provider')}")
            print(f"   - Slots: {len(doc.get('slots_found', []))}")
    else:
        print("\n⚠️  No documents found for 'Stratford Park (Test)'")
    
    # Close connection
    standardizer.close_connection()


if __name__ == "__main__":
    verify_mongodb_data() 