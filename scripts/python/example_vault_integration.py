#!/usr/bin/env python3
"""
Example script demonstrating how Python scrapers and booking automation
scripts can integrate with HashiCorp Vault for credential management.

This script shows:
- How to load credentials from Vault
- Error handling for missing credentials
- Using credentials for authentication
- Best practices for secure credential handling

Usage:
    python example_vault_integration.py
"""

import os
import sys
import logging
from typing import Dict, Any

# Add the current directory to Python path for imports
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from vault_client import VaultClient, get_platform_credentials
from hvac.exceptions import VaultError, InvalidPath


def setup_logging():
    """Configure logging for the script."""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )


def mask_secret(secret: str, show_chars: int = 2) -> str:
    """Mask a secret for safe logging."""
    if len(secret) <= show_chars * 2:
        return '*' * len(secret)
    return secret[:show_chars] + '*' * (len(secret) - show_chars * 2) + secret[-show_chars:]


def demonstrate_vault_integration():
    """Demonstrate various Vault integration patterns."""
    print("🔐 Tennis Booking Bot - Python Vault Integration Example")
    print("=" * 60)
    
    # Method 1: Using the convenience function
    print("\n📋 Method 1: Using convenience function")
    try:
        lta_creds = get_platform_credentials('lta')
        print(f"✓ LTA Username: {lta_creds['username']}")
        print(f"✓ LTA Base URL: {lta_creds['base_url']}")
        print(f"✓ LTA API Key: {mask_secret(lta_creds['api_key'])}")
    except VaultError as e:
        print(f"✗ Failed to get LTA credentials: {e}")
        return False
    except Exception as e:
        print(f"✗ Unexpected error: {e}")
        return False
    
    # Method 2: Using VaultClient directly
    print("\n📋 Method 2: Using VaultClient directly")
    try:
        client = VaultClient()
        
        # Health check
        if not client.health_check():
            print("✗ Vault health check failed")
            return False
        print("✓ Vault health check passed")
        
        # Get Courtsides credentials
        courtsides_creds = client.get_secret("secret/data/tennis-bot/courtsides")
        print(f"✓ Courtsides Username: {courtsides_creds['username']}")
        print(f"✓ Courtsides Login URL: {courtsides_creds['login_url']}")
        print(f"✓ Courtsides Booking URL: {courtsides_creds['booking_url']}")
        
        # Get specific field
        lta_username = client.get_secret_field("secret/data/tennis-bot/lta", "username")
        print(f"✓ LTA Username (field access): {lta_username}")
        
        client.close()
        
    except VaultError as e:
        print(f"✗ Vault error: {e}")
        return False
    except Exception as e:
        print(f"✗ Unexpected error: {e}")
        return False
    
    return True


def simulate_scraper_workflow():
    """Simulate a typical scraper workflow using Vault credentials."""
    print("\n🤖 Simulating Scraper Workflow")
    print("-" * 40)
    
    try:
        # Load credentials for the platform we want to scrape
        platform = "lta"
        print(f"Loading credentials for {platform.upper()} platform...")
        
        creds = get_platform_credentials(platform)
        
        # Simulate authentication (don't actually make requests in this example)
        print(f"✓ Authenticating with {platform.upper()} using username: {creds['username']}")
        print(f"✓ Using base URL: {creds['base_url']}")
        
        # Simulate scraping process
        print("✓ Simulating court availability scraping...")
        print("✓ Simulating data processing...")
        print("✓ Simulating database storage...")
        
        print(f"✅ {platform.upper()} scraping workflow completed successfully")
        
    except VaultError as e:
        print(f"✗ Failed to load credentials: {e}")
        return False
    except Exception as e:
        print(f"✗ Scraper workflow failed: {e}")
        return False
    
    return True


def simulate_booking_workflow():
    """Simulate a booking automation workflow using Vault credentials."""
    print("\n📅 Simulating Booking Workflow")
    print("-" * 40)
    
    try:
        # Load credentials for booking platform
        platform = "courtsides"
        print(f"Loading credentials for {platform} platform...")
        
        creds = get_platform_credentials(platform)
        
        # Simulate booking process
        print(f"✓ Logging into {platform} using username: {creds['username']}")
        print(f"✓ Using login URL: {creds['login_url']}")
        print(f"✓ Navigating to booking URL: {creds['booking_url']}")
        
        print("✓ Simulating court search...")
        print("✓ Simulating booking attempt...")
        print("✓ Simulating confirmation...")
        
        print(f"✅ {platform} booking workflow completed successfully")
        
    except VaultError as e:
        print(f"✗ Failed to load credentials: {e}")
        return False
    except Exception as e:
        print(f"✗ Booking workflow failed: {e}")
        return False
    
    return True


def demonstrate_error_handling():
    """Demonstrate proper error handling for Vault operations."""
    print("\n⚠️  Demonstrating Error Handling")
    print("-" * 40)
    
    # Test with non-existent platform
    try:
        print("Attempting to load credentials for non-existent platform...")
        get_platform_credentials('nonexistent_platform')
        print("✗ This should not have succeeded")
    except ValueError as e:
        print(f"✓ Correctly caught ValueError: {e}")
    except Exception as e:
        print(f"✗ Unexpected error type: {e}")
    
    # Test with invalid Vault configuration
    try:
        print("Testing with invalid Vault token...")
        client = VaultClient(vault_token="invalid_token")
        client.get_secret("secret/data/tennis-bot/lta")
        print("✗ This should not have succeeded")
    except VaultError as e:
        print(f"✓ Correctly caught VaultError: {e}")
    except Exception as e:
        print(f"✗ Unexpected error type: {e}")


def main():
    """Main function demonstrating Python Vault integration."""
    setup_logging()
    
    # Set environment variables if not already set
    if not os.getenv('VAULT_ADDR'):
        os.environ['VAULT_ADDR'] = 'http://localhost:8200'
    if not os.getenv('VAULT_TOKEN'):
        os.environ['VAULT_TOKEN'] = 'dev-token'
    
    print("Environment:")
    print(f"  VAULT_ADDR: {os.getenv('VAULT_ADDR')}")
    print(f"  VAULT_TOKEN: {'***' if os.getenv('VAULT_TOKEN') else 'Not set'}")
    
    success = True
    
    # Run demonstrations
    success &= demonstrate_vault_integration()
    success &= simulate_scraper_workflow()
    success &= simulate_booking_workflow()
    
    # Always run error handling demo
    demonstrate_error_handling()
    
    if success:
        print("\n🎉 All Python Vault integration examples completed successfully!")
        print("\nKey takeaways for Python scripts:")
        print("  • Use get_platform_credentials() for simple credential loading")
        print("  • Use VaultClient directly for advanced operations")
        print("  • Always handle VaultError and other exceptions")
        print("  • Perform health checks before critical operations")
        print("  • Never log actual credentials - use masking")
        print("  • Set VAULT_ADDR and VAULT_TOKEN environment variables")
    else:
        print("\n❌ Some examples failed - check Vault configuration")
        sys.exit(1)


if __name__ == "__main__":
    main() 