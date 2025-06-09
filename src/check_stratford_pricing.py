#!/usr/bin/env python3

import pymongo
from collections import Counter

# Connect to MongoDB with authentication
client = pymongo.MongoClient("mongodb://admin:YOUR_PASSWORD@localhost:27017")
db = client["tennis_booking"]

print("=== Focused Analysis: Stratford Park Pricing ===")

# Check all collections to see where Stratford Park data might be
collections = ['slots', 'ScrapingLogs', 'scraping_logs']

for collection_name in collections:
    print(f"\n--- Checking collection: {collection_name} ---")
    collection = db[collection_name]
    
    # Count total documents
    total_docs = collection.count_documents({})
    print(f"Total documents: {total_docs}")
    
    # Check for Stratford Park specifically
    stratford_docs = collection.count_documents({"venue_name": "Stratford Park"})
    stratford_docs_alt = collection.count_documents({"venueName": "Stratford Park"})
    
    print(f"Stratford Park (venue_name): {stratford_docs}")
    print(f"Stratford Park (venueName): {stratford_docs_alt}")
    
    # Sample document structure
    sample = collection.find_one()
    if sample:
        print(f"Sample document keys: {list(sample.keys())}")

print(f"\n--- Direct Stratford Park Slot Search ---")

# Try different field names for venue
venue_queries = [
    {"venue_name": "Stratford Park"},
    {"venueName": "Stratford Park"},
    {"venue": "Stratford Park"}
]

for query in venue_queries:
    print(f"\nQuery: {query}")
    slots = list(db.slots.find(query).limit(5))
    print(f"Found {len(slots)} slots")
    
    if slots:
        # Show first slot structure and pricing
        first_slot = slots[0]
        print(f"Sample slot structure:")
        for key, value in first_slot.items():
            if key != '_id':
                print(f"  {key}: {value}")
        
        # Show all prices found
        all_slots = list(db.slots.find(query))
        prices = [slot.get('price') for slot in all_slots if slot.get('price') is not None]
        if prices:
            price_counts = Counter(prices)
            print(f"Price distribution:")
            for price, count in sorted(price_counts.items()):
                print(f"  £{price}: {count} slots")
        break 