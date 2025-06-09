"""
HashiCorp Vault client utility for Python scripts.

This module provides a simple interface for retrieving secrets from Vault
for use in Python-based scrapers and booking automation scripts.
"""

import os
import logging
from typing import Dict, Any, Optional

import hvac
from hvac.exceptions import VaultError, InvalidPath


class VaultClient:
    """
    A client for interacting with HashiCorp Vault.
    
    This client handles authentication and secret retrieval from Vault,
    with support for both environment variable and explicit configuration.
    """
    
    def __init__(self, vault_url: Optional[str] = None, vault_token: Optional[str] = None):
        """
        Initialize the Vault client.
        
        Args:
            vault_url: Vault server URL. If None, uses VAULT_ADDR env var or default.
            vault_token: Vault authentication token. If None, uses VAULT_TOKEN env var.
            
        Raises:
            ValueError: If vault_token is not provided and VAULT_TOKEN env var is not set.
            VaultError: If unable to connect to Vault or authentication fails.
        """
        self.vault_url = vault_url or os.getenv('VAULT_ADDR', 'http://localhost:8200')
        self.vault_token = vault_token or os.getenv('VAULT_TOKEN')
        
        if not self.vault_token:
            raise ValueError("Vault token is required. Set VAULT_TOKEN environment variable or pass vault_token parameter.")
        
        # Initialize the hvac client
        self.client = hvac.Client(url=self.vault_url, token=self.vault_token)
        
        # Verify authentication
        if not self.client.is_authenticated():
            raise VaultError(f"Failed to authenticate with Vault at {self.vault_url}")
        
        self.logger = logging.getLogger(__name__)
        self.logger.info(f"Successfully connected to Vault at {self.vault_url}")
    
    def get_secret(self, path: str) -> Dict[str, Any]:
        """
        Retrieve a secret from the specified path.
        
        Args:
            path: The path to the secret in Vault (e.g., 'secret/data/tennis-bot/lta')
            
        Returns:
            Dictionary containing the secret data
            
        Raises:
            VaultError: If the secret cannot be retrieved
            InvalidPath: If the path does not exist
        """
        try:
            self.logger.debug(f"Retrieving secret from path: {path}")
            
            # Read the secret from Vault
            response = self.client.secrets.kv.v2.read_secret_version(
                path=path.replace('secret/data/', '').replace('secret/', ''),
                mount_point='secret'
            )
            
            if not response or 'data' not in response:
                raise InvalidPath(f"Secret not found at path: {path}")
            
            secret_data = response['data']['data']
            self.logger.debug(f"Successfully retrieved secret from {path}")
            
            return secret_data
            
        except hvac.exceptions.InvalidPath:
            self.logger.error(f"Secret not found at path: {path}")
            raise InvalidPath(f"Secret not found at path: {path}")
        except Exception as e:
            self.logger.error(f"Failed to retrieve secret from {path}: {str(e)}")
            raise VaultError(f"Failed to retrieve secret from {path}: {str(e)}")
    
    def get_secret_field(self, path: str, field: str) -> str:
        """
        Retrieve a specific field from a secret.
        
        Args:
            path: The path to the secret in Vault
            field: The specific field to retrieve from the secret
            
        Returns:
            The value of the specified field as a string
            
        Raises:
            VaultError: If the secret or field cannot be retrieved
            KeyError: If the field does not exist in the secret
        """
        secret_data = self.get_secret(path)
        
        if field not in secret_data:
            raise KeyError(f"Field '{field}' not found in secret at path '{path}'")
        
        value = secret_data[field]
        if not isinstance(value, str):
            self.logger.warning(f"Field '{field}' at path '{path}' is not a string, converting to string")
            value = str(value)
        
        return value
    
    def health_check(self) -> bool:
        """
        Perform a health check on the Vault connection.
        
        Returns:
            True if Vault is healthy and accessible, False otherwise
        """
        try:
            health_status = self.client.sys.read_health_status()
            # health_status is a Response object, we need to check the status code
            is_healthy = health_status.status_code == 200
            
            if is_healthy:
                self.logger.debug("Vault health check passed")
            else:
                self.logger.warning(f"Vault health check failed - Status code: {health_status.status_code}")
            
            return is_healthy
            
        except Exception as e:
            self.logger.error(f"Vault health check failed: {str(e)}")
            return False
    
    def close(self):
        """
        Close the Vault client connection.
        
        Currently a no-op but included for consistency with the Go client
        and potential future cleanup needs.
        """
        self.logger.debug("Vault client closed")


def get_platform_credentials(platform: str) -> Dict[str, str]:
    """
    Convenience function to get credentials for a specific platform.
    
    Args:
        platform: Platform name ('lta' or 'courtsides')
        
    Returns:
        Dictionary containing platform credentials
        
    Raises:
        ValueError: If platform is not supported
        VaultError: If credentials cannot be retrieved
    """
    supported_platforms = ['lta', 'courtsides']
    
    if platform not in supported_platforms:
        raise ValueError(f"Unsupported platform: {platform}. Supported platforms: {supported_platforms}")
    
    client = VaultClient()
    path = f"secret/data/tennis-bot/{platform}"
    
    try:
        credentials = client.get_secret(path)
        client.close()
        return credentials
    except Exception as e:
        client.close()
        raise VaultError(f"Failed to retrieve {platform} credentials: {str(e)}")


if __name__ == "__main__":
    # Simple test when run directly
    import sys
    
    logging.basicConfig(level=logging.INFO)
    
    try:
        print("Testing Vault Python client...")
        
        # Test client creation
        client = VaultClient()
        print("✓ Vault client created successfully")
        
        # Test health check
        if client.health_check():
            print("✓ Vault health check passed")
        else:
            print("✗ Vault health check failed")
            sys.exit(1)
        
        # Test LTA credentials
        lta_creds = client.get_secret("secret/data/tennis-bot/lta")
        print(f"✓ LTA credentials retrieved: {list(lta_creds.keys())}")
        
        # Test specific field
        username = client.get_secret_field("secret/data/tennis-bot/lta", "username")
        print(f"✓ LTA username: {username}")
        
        # Test courtsides credentials
        courtsides_creds = client.get_secret("secret/data/tennis-bot/courtsides")
        print(f"✓ Courtsides credentials retrieved: {list(courtsides_creds.keys())}")
        
        # Test convenience function
        lta_creds_conv = get_platform_credentials("lta")
        print(f"✓ LTA credentials via convenience function: {list(lta_creds_conv.keys())}")
        
        client.close()
        print("\n🎉 All Python Vault tests passed!")
        
    except Exception as e:
        print(f"✗ Test failed: {str(e)}")
        sys.exit(1) 