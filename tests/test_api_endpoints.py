#!/usr/bin/env python3
"""
Test script for Tennis Booking API preference endpoints
"""

import requests
import json
import time
from bson import ObjectId

API_BASE_URL = "http://localhost:8080"
TEST_USER_ID = str(ObjectId())  # Generate a test user ID

def test_health_endpoint():
    """Test the health endpoint"""
    print("Testing health endpoint...")
    try:
        response = requests.get(f"{API_BASE_URL}/health")
        if response.status_code == 200:
            print("✅ Health endpoint working")
            print(f"   Response: {response.json()}")
            return True
        else:
            print(f"❌ Health endpoint failed: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Health endpoint error: {e}")
        return False

def test_get_preferences():
    """Test GET /api/v1/preferences"""
    print(f"\nTesting GET preferences for user {TEST_USER_ID}...")
    try:
        response = requests.get(f"{API_BASE_URL}/api/v1/preferences", 
                               params={"user_id": TEST_USER_ID})
        if response.status_code == 200:
            print("✅ GET preferences working")
            data = response.json()
            print(f"   Response: {json.dumps(data, indent=2)}")
            return True, data
        else:
            print(f"❌ GET preferences failed: {response.status_code}")
            print(f"   Response: {response.text}")
            return False, None
    except Exception as e:
        print(f"❌ GET preferences error: {e}")
        return False, None

def test_update_preferences():
    """Test PUT /api/v1/preferences"""
    print(f"\nTesting PUT preferences for user {TEST_USER_ID}...")
    
    test_preferences = {
        "times": [
            {"start": "08:00", "end": "10:00"},
            {"start": "18:00", "end": "20:00"}
        ],
        "max_price": 25.0,
        "preferred_venues": ["venue1", "venue2"],
        "excluded_venues": ["venue3"],
        "preferred_days": ["monday", "wednesday", "friday"],
        "notification_settings": {
            "email": True,
            "sms": False,
            "minutes_before_play": 30,
            "email_address": "test@example.com"
        }
    }
    
    try:
        response = requests.put(f"{API_BASE_URL}/api/v1/preferences",
                               params={"user_id": TEST_USER_ID},
                               json=test_preferences)
        if response.status_code == 200:
            print("✅ PUT preferences working")
            data = response.json()
            print(f"   Response: {json.dumps(data, indent=2)}")
            return True, data
        else:
            print(f"❌ PUT preferences failed: {response.status_code}")
            print(f"   Response: {response.text}")
            return False, None
    except Exception as e:
        print(f"❌ PUT preferences error: {e}")
        return False, None

def test_add_venue():
    """Test POST /api/v1/preferences/venues"""
    print(f"\nTesting POST add venue for user {TEST_USER_ID}...")
    
    test_venue = {
        "venue_id": "test_venue_123",
        "venue_type": "preferred"
    }
    
    try:
        response = requests.post(f"{API_BASE_URL}/api/v1/preferences/venues",
                                params={"user_id": TEST_USER_ID},
                                json=test_venue)
        if response.status_code == 201:
            print("✅ POST add venue working")
            data = response.json()
            print(f"   Response: {json.dumps(data, indent=2)}")
            return True, data
        else:
            print(f"❌ POST add venue failed: {response.status_code}")
            print(f"   Response: {response.text}")
            return False, None
    except Exception as e:
        print(f"❌ POST add venue error: {e}")
        return False, None

def test_remove_venue():
    """Test DELETE /api/v1/preferences/venues/:venueId"""
    print(f"\nTesting DELETE remove venue for user {TEST_USER_ID}...")
    
    venue_id = "test_venue_123"
    
    try:
        response = requests.delete(f"{API_BASE_URL}/api/v1/preferences/venues/{venue_id}",
                                  params={"user_id": TEST_USER_ID, "list_type": "preferred"})
        if response.status_code == 200:
            print("✅ DELETE remove venue working")
            data = response.json()
            print(f"   Response: {json.dumps(data, indent=2)}")
            return True, data
        else:
            print(f"❌ DELETE remove venue failed: {response.status_code}")
            print(f"   Response: {response.text}")
            return False, None
    except Exception as e:
        print(f"❌ DELETE remove venue error: {e}")
        return False, None

def test_validation_errors():
    """Test input validation"""
    print(f"\nTesting input validation...")
    
    # Test invalid time format
    invalid_preferences = {
        "times": [
            {"start": "25:00", "end": "10:00"}  # Invalid hour
        ]
    }
    
    try:
        response = requests.put(f"{API_BASE_URL}/api/v1/preferences",
                               params={"user_id": TEST_USER_ID},
                               json=invalid_preferences)
        if response.status_code == 400:
            print("✅ Time validation working")
            print(f"   Error response: {response.json()}")
        else:
            print(f"❌ Time validation not working: {response.status_code}")
            
        # Test invalid user ID
        response = requests.get(f"{API_BASE_URL}/api/v1/preferences",
                               params={"user_id": "invalid_id"})
        if response.status_code == 400:
            print("✅ User ID validation working")
        else:
            print(f"❌ User ID validation not working: {response.status_code}")
            
        return True
    except Exception as e:
        print(f"❌ Validation testing error: {e}")
        return False

def main():
    """Run all API tests"""
    print("=" * 60)
    print("Tennis Booking API - Preferences Endpoint Tests")
    print("=" * 60)
    print(f"Testing against: {API_BASE_URL}")
    print(f"Test User ID: {TEST_USER_ID}")
    print("=" * 60)
    
    # Give server time to start up if needed
    time.sleep(1)
    
    test_results = []
    
    # Run all tests
    test_results.append(test_health_endpoint())
    test_results.append(test_get_preferences()[0])
    test_results.append(test_update_preferences()[0])
    test_results.append(test_add_venue()[0])
    test_results.append(test_remove_venue()[0])
    test_results.append(test_validation_errors())
    
    # Summary
    print("\n" + "=" * 60)
    print("TEST SUMMARY")
    print("=" * 60)
    passed = sum(test_results)
    total = len(test_results)
    print(f"Tests passed: {passed}/{total}")
    
    if passed == total:
        print("🎉 All tests passed!")
        return 0
    else:
        print("⚠️  Some tests failed")
        return 1

if __name__ == "__main__":
    exit(main()) 