"""
Unit tests for the Vault client utility.

Run with: python -m pytest test_vault_client.py -v
"""

import os
import pytest
from unittest.mock import Mock, patch, MagicMock
import logging

from vault_client import VaultClient, get_platform_credentials
from hvac.exceptions import VaultError, InvalidPath


class TestVaultClient:
    """Test cases for VaultClient class."""
    
    def test_init_with_explicit_config(self):
        """Test VaultClient initialization with explicit configuration."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            
            assert client.vault_url == "http://test:8200"
            assert client.vault_token == "test-token"
            mock_hvac.assert_called_once_with(url="http://test:8200", token="test-token")
    
    def test_init_with_env_vars(self):
        """Test VaultClient initialization using environment variables."""
        with patch.dict(os.environ, {'VAULT_ADDR': 'http://env:8200', 'VAULT_TOKEN': 'env-token'}):
            with patch('vault_client.hvac.Client') as mock_hvac:
                mock_client = Mock()
                mock_client.is_authenticated.return_value = True
                mock_hvac.return_value = mock_client
                
                client = VaultClient()
                
                assert client.vault_url == "http://env:8200"
                assert client.vault_token == "env-token"
    
    def test_init_missing_token(self):
        """Test VaultClient initialization fails when token is missing."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ValueError, match="Vault token is required"):
                VaultClient()
    
    def test_init_authentication_failure(self):
        """Test VaultClient initialization fails when authentication fails."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = False
            mock_hvac.return_value = mock_client
            
            with pytest.raises(VaultError, match="Failed to authenticate"):
                VaultClient(vault_url="http://test:8200", vault_token="invalid-token")
    
    def test_get_secret_success(self):
        """Test successful secret retrieval."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_client.secrets.kv.v2.read_secret_version.return_value = {
                'data': {
                    'data': {
                        'username': 'test_user',
                        'password': 'test_pass'
                    }
                }
            }
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            result = client.get_secret("secret/data/tennis-bot/lta")
            
            assert result == {'username': 'test_user', 'password': 'test_pass'}
            mock_client.secrets.kv.v2.read_secret_version.assert_called_once_with(
                path='tennis-bot/lta',
                mount_point='secret'
            )
    
    def test_get_secret_not_found(self):
        """Test secret retrieval when secret doesn't exist."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_client.secrets.kv.v2.read_secret_version.side_effect = InvalidPath("Not found")
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            
            with pytest.raises(InvalidPath, match="Secret not found"):
                client.get_secret("secret/data/tennis-bot/nonexistent")
    
    def test_get_secret_field_success(self):
        """Test successful retrieval of a specific field."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_client.secrets.kv.v2.read_secret_version.return_value = {
                'data': {
                    'data': {
                        'username': 'test_user',
                        'password': 'test_pass'
                    }
                }
            }
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            result = client.get_secret_field("secret/data/tennis-bot/lta", "username")
            
            assert result == "test_user"
    
    def test_get_secret_field_not_found(self):
        """Test retrieval of a non-existent field."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_client.secrets.kv.v2.read_secret_version.return_value = {
                'data': {
                    'data': {
                        'username': 'test_user',
                        'password': 'test_pass'
                    }
                }
            }
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            
            with pytest.raises(KeyError, match="Field 'nonexistent' not found"):
                client.get_secret_field("secret/data/tennis-bot/lta", "nonexistent")
    
    def test_get_secret_field_type_conversion(self):
        """Test field retrieval with type conversion."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_client.secrets.kv.v2.read_secret_version.return_value = {
                'data': {
                    'data': {
                        'port': 8080,  # Non-string value
                        'enabled': True
                    }
                }
            }
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            
            port = client.get_secret_field("secret/data/test", "port")
            enabled = client.get_secret_field("secret/data/test", "enabled")
            
            assert port == "8080"
            assert enabled == "True"
    
    def test_health_check_success(self):
        """Test successful health check."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_response = Mock()
            mock_response.status_code = 200
            mock_client.sys.read_health_status.return_value = mock_response
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            result = client.health_check()
            
            assert result is True
    
    def test_health_check_sealed(self):
        """Test health check when Vault is sealed."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_response = Mock()
            mock_response.status_code = 503  # Vault returns 503 when sealed
            mock_client.sys.read_health_status.return_value = mock_response
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            result = client.health_check()
            
            assert result is False
    
    def test_health_check_exception(self):
        """Test health check when an exception occurs."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_client.sys.read_health_status.side_effect = Exception("Connection error")
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            result = client.health_check()
            
            assert result is False
    
    def test_close(self):
        """Test client close method."""
        with patch('vault_client.hvac.Client') as mock_hvac:
            mock_client = Mock()
            mock_client.is_authenticated.return_value = True
            mock_hvac.return_value = mock_client
            
            client = VaultClient(vault_url="http://test:8200", vault_token="test-token")
            client.close()  # Should not raise any exception


class TestGetPlatformCredentials:
    """Test cases for get_platform_credentials convenience function."""
    
    def test_get_lta_credentials(self):
        """Test getting LTA credentials."""
        with patch('vault_client.VaultClient') as mock_vault_client:
            mock_client_instance = Mock()
            mock_client_instance.get_secret.return_value = {
                'username': 'lta_user',
                'password': 'lta_pass'
            }
            mock_vault_client.return_value = mock_client_instance
            
            result = get_platform_credentials('lta')
            
            assert result == {'username': 'lta_user', 'password': 'lta_pass'}
            mock_client_instance.get_secret.assert_called_once_with('secret/data/tennis-bot/lta')
            mock_client_instance.close.assert_called_once()
    
    def test_get_courtsides_credentials(self):
        """Test getting courtsides credentials."""
        with patch('vault_client.VaultClient') as mock_vault_client:
            mock_client_instance = Mock()
            mock_client_instance.get_secret.return_value = {
                'username': 'courtsides_user',
                'password': 'courtsides_pass'
            }
            mock_vault_client.return_value = mock_client_instance
            
            result = get_platform_credentials('courtsides')
            
            assert result == {'username': 'courtsides_user', 'password': 'courtsides_pass'}
            mock_client_instance.get_secret.assert_called_once_with('secret/data/tennis-bot/courtsides')
            mock_client_instance.close.assert_called_once()
    
    def test_unsupported_platform(self):
        """Test error handling for unsupported platform."""
        with pytest.raises(ValueError, match="Unsupported platform: invalid"):
            get_platform_credentials('invalid')
    
    def test_vault_error_handling(self):
        """Test error handling when Vault operations fail."""
        with patch('vault_client.VaultClient') as mock_vault_client:
            mock_client_instance = Mock()
            mock_client_instance.get_secret.side_effect = VaultError("Vault error")
            mock_vault_client.return_value = mock_client_instance
            
            with pytest.raises(VaultError, match="Failed to retrieve lta credentials"):
                get_platform_credentials('lta')
            
            # Ensure close is called even on error
            mock_client_instance.close.assert_called_once()


# Integration tests (require running Vault instance)
class TestVaultClientIntegration:
    """Integration tests that require a running Vault instance."""
    
    @pytest.mark.skipif(
        not os.getenv('VAULT_ADDR') or not os.getenv('VAULT_TOKEN'),
        reason="Integration tests require VAULT_ADDR and VAULT_TOKEN environment variables"
    )
    def test_real_vault_connection(self):
        """Test connection to a real Vault instance."""
        client = VaultClient()
        
        # Test health check
        assert client.health_check() is True
        
        # Test reading existing secret (assumes secrets were created in task 3.1)
        lta_secret = client.get_secret("secret/data/tennis-bot/lta")
        assert 'username' in lta_secret
        assert 'password' in lta_secret
        
        # Test reading specific field
        username = client.get_secret_field("secret/data/tennis-bot/lta", "username")
        assert username == "lta_test_user"
        
        client.close()
    
    @pytest.mark.skipif(
        not os.getenv('VAULT_ADDR') or not os.getenv('VAULT_TOKEN'),
        reason="Integration tests require VAULT_ADDR and VAULT_TOKEN environment variables"
    )
    def test_convenience_function_integration(self):
        """Test the convenience function with real Vault."""
        lta_creds = get_platform_credentials('lta')
        assert 'username' in lta_creds
        assert 'password' in lta_creds
        
        courtsides_creds = get_platform_credentials('courtsides')
        assert 'username' in courtsides_creds
        assert 'password' in courtsides_creds


if __name__ == "__main__":
    # Run tests when executed directly
    pytest.main([__file__, "-v"]) 