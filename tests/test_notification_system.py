#!/usr/bin/env python3
"""
Test script for the Tennis Court Notification System
"""

import subprocess
import json
import time
from datetime import datetime

def test_notification_service_compilation():
    """Test that the notification service compiles successfully"""
    print("Testing notification service compilation...")
    
    try:
        result = subprocess.run(
            ["go", "build", "-o", "/tmp/notification-service-test", "./cmd/notification-service"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✅ Notification service compiles successfully")
            return True
        else:
            print(f"❌ Compilation failed: {result.stderr}")
            return False
            
    except subprocess.TimeoutExpired:
        print("❌ Compilation timed out")
        return False
    except Exception as e:
        print(f"❌ Compilation error: {e}")
        return False

def test_api_service_compilation():
    """Test that the API service still compiles after refactoring"""
    print("Testing API service compilation...")
    
    try:
        result = subprocess.run(
            ["go", "build", "-o", "/tmp/api-test", "./cmd/api"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✅ API service compiles successfully")
            return True
        else:
            print(f"❌ API compilation failed: {result.stderr}")
            return False
            
    except subprocess.TimeoutExpired:
        print("❌ API compilation timed out")
        return False
    except Exception as e:
        print(f"❌ API compilation error: {e}")
        return False

def test_scheduler_service_compilation():
    """Test that the scheduler service still compiles after refactoring"""
    print("Testing scheduler service compilation...")
    
    try:
        result = subprocess.run(
            ["go", "build", "-o", "/tmp/scheduler-test", "./cmd/scheduler"],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0:
            print("✅ Scheduler service compiles successfully")
            return True
        else:
            print(f"❌ Scheduler compilation failed: {result.stderr}")
            return False
            
    except subprocess.TimeoutExpired:
        print("❌ Scheduler compilation timed out")
        return False
    except Exception as e:
        print(f"❌ Scheduler compilation error: {e}")
        return False

def test_makefile_build():
    """Test that the Makefile build target works"""
    print("Testing Makefile build...")
    
    try:
        result = subprocess.run(
            ["make", "build"],
            capture_output=True,
            text=True,
            timeout=60
        )
        
        if result.returncode == 0:
            print("✅ Makefile build successful")
            return True
        else:
            print(f"❌ Makefile build failed: {result.stderr}")
            return False
            
    except subprocess.TimeoutExpired:
        print("❌ Makefile build timed out")
        return False
    except Exception as e:
        print(f"❌ Makefile build error: {e}")
        return False

def test_removed_files():
    """Test that old booking-related files have been properly removed"""
    print("Testing removal of old booking files...")
    
    import os
    
    removed_files = [
        "cmd/booking-engine/main.go",
        "internal/booking/engine.go", 
        "internal/booking/python_caller.go",
        "scripts/python/court_booker.py"
    ]
    
    all_removed = True
    for file_path in removed_files:
        if os.path.exists(file_path):
            print(f"❌ File still exists: {file_path}")
            all_removed = False
        else:
            print(f"✅ File properly removed: {file_path}")
    
    return all_removed

def test_new_files_exist():
    """Test that new notification-related files exist"""
    print("Testing existence of new notification files...")
    
    import os
    
    new_files = [
        "cmd/notification-service/main.go",
        "internal/booking/notification_engine.go",
        "internal/models/notification.go"
    ]
    
    all_exist = True
    for file_path in new_files:
        if os.path.exists(file_path):
            print(f"✅ New file exists: {file_path}")
        else:
            print(f"❌ New file missing: {file_path}")
            all_exist = False
    
    return all_exist

def main():
    """Run all refactoring tests"""
    print("🔧 Testing Tennis Court Notification System Refactoring")
    print("=" * 60)
    
    tests = [
        ("File Removal", test_removed_files),
        ("New Files", test_new_files_exist),
        ("Makefile Build", test_makefile_build),
        ("API Service", test_api_service_compilation),
        ("Scheduler Service", test_scheduler_service_compilation),
        ("Notification Service", test_notification_service_compilation),
    ]
    
    results = []
    
    for test_name, test_func in tests:
        print(f"\n📋 Running {test_name} test...")
        try:
            success = test_func()
            results.append((test_name, success))
        except Exception as e:
            print(f"❌ {test_name} test failed with exception: {e}")
            results.append((test_name, False))
    
    # Summary
    print("\n" + "=" * 60)
    print("📊 REFACTORING TEST SUMMARY")
    print("=" * 60)
    
    passed = 0
    total = len(results)
    
    for test_name, success in results:
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} - {test_name}")
        if success:
            passed += 1
    
    print(f"\nResults: {passed}/{total} tests passed ({passed/total*100:.1f}%)")
    
    if passed == total:
        print("🎉 All refactoring tests passed! Ready for Task 8.")
        return True
    else:
        print("⚠️  Some tests failed. Please review the issues above.")
        return False

if __name__ == "__main__":
    success = main()
    exit(0 if success else 1) 