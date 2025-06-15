#!/usr/bin/env python3
"""
End-to-end test for Python configuration system
"""

import sys
import os
sys.path.append('src')

from config.config import Config

def test_python_configuration():
    print("=== TESTING PYTHON SCRAPER CONFIGURATION ===")
    
    print("\n--- Testing Default Configuration ---")
    config = Config()
    print(f"App Name: {config.get('app.name')}")
    print(f"App Version: {config.get('app.version')}")
    print(f"Environment: {config.get('app.environment')}")
    print(f"Scraper Interval: {config.get('scraper.interval')}")
    print(f"Log Level: {config.get('logging.level')}")
    print(f"ClubSpark Enabled: {config.get('scraper.platforms.clubspark.enabled')}")
    print(f"Analytics Enabled: {config.get('scraper.platforms.analytics.enabled')}")
    
    print("\n--- Testing Feature Flags ---")
    print(f"Analytics Feature: {config.is_feature_enabled('analytics')}")
    print(f"Notifications Feature: {config.is_feature_enabled('notifications')}")
    
    print("\n--- Testing Duration Parsing ---")
    print(f"Scraper Timeout (parsed): {config.get('scraper.timeout')} seconds")
    print(f"API Timeout (parsed): {config.get('api.timeout')} seconds")
    
    print("\n--- Testing Environment Variable Overrides ---")
    # Set environment variables
    os.environ['APP_ENV'] = 'test'
    os.environ['SCRAPER_INTERVAL'] = '30'
    os.environ['LOG_LEVEL'] = 'warn'
    os.environ['SCRAPER_PLATFORMS_CLUBSPARK_ENABLED'] = 'false'
    os.environ['SCRAPER_TIMEOUT'] = '2m'  # Test duration parsing
    
    # Create new config instance to pick up env vars
    config2 = Config()
    print(f"Environment (should be test): {config2.get('app.environment')}")
    print(f"Scraper Interval (should be 30): {config2.get('scraper.interval')}")
    print(f"Log Level (should be warn): {config2.get('logging.level')}")
    print(f"ClubSpark Enabled (should be False): {config2.get('scraper.platforms.clubspark.enabled')}")
    print(f"Scraper Timeout (should be 120): {config2.get('scraper.timeout')}")
    
    print("\n--- Testing Different Environments ---")
    environments = ['development', 'production', 'test']
    
    for env in environments:
        print(f"\n  Environment: {env}")
        os.environ['APP_ENV'] = env
        config_env = Config()
        print(f"    Log Level: {config_env.get('logging.level')}")
        print(f"    Scraper Interval: {config_env.get('scraper.interval')}")
        print(f"    Analytics Enabled: {config_env.get('scraper.platforms.analytics.enabled')}")
    
    print("\n✅ Python configuration system test completed successfully!")

if __name__ == "__main__":
    test_python_configuration() 