package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVaultClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *VaultConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &VaultConfig{
				Address: "http://localhost:8200",
				Token:   "test-token",
			},
			expectError: false,
		},
		{
			name: "missing token",
			config: &VaultConfig{
				Address: "http://localhost:8200",
				Token:   "",
			},
			expectError: true,
			errorMsg:    "vault token is required",
		},
		{
			name:        "nil config without env vars",
			config:      nil,
			expectError: true,
			errorMsg:    "vault token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables for consistent testing
			os.Unsetenv("VAULT_ADDR")
			os.Unsetenv("VAULT_TOKEN")

			client, err := NewVaultClient(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.client)
			}
		})
	}
}

func TestNewVaultClientFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		vaultAddr   string
		vaultToken  string
		expectError bool
	}{
		{
			name:        "valid env vars",
			vaultAddr:   "http://localhost:8200",
			vaultToken:  "test-token",
			expectError: false,
		},
		{
			name:        "missing token env var",
			vaultAddr:   "http://localhost:8200",
			vaultToken:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("VAULT_ADDR", tt.vaultAddr)
			os.Setenv("VAULT_TOKEN", tt.vaultToken)
			defer func() {
				os.Unsetenv("VAULT_ADDR")
				os.Unsetenv("VAULT_TOKEN")
			}()

			client, err := NewVaultClientFromEnv()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "env_value",
			expected:     "env_value",
		},
		{
			name:         "env var does not exist",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVaultClient_Methods_WithoutConnection(t *testing.T) {
	// Test methods without actually connecting to Vault
	// These tests verify error handling for uninitialized clients

	t.Run("GetSecret with nil client", func(t *testing.T) {
		client := &VaultClient{client: nil}
		_, err := client.GetSecret("test/path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vault client not initialized")
	})

	t.Run("GetSecretField with nil client", func(t *testing.T) {
		client := &VaultClient{client: nil}
		_, err := client.GetSecretField("test/path", "field")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vault client not initialized")
	})

	t.Run("HealthCheck with nil client", func(t *testing.T) {
		client := &VaultClient{client: nil}
		err := client.HealthCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vault client not initialized")
	})

	t.Run("Close always succeeds", func(t *testing.T) {
		client := &VaultClient{client: nil}
		err := client.Close()
		assert.NoError(t, err)
	})
}

// Integration tests - these require a running Vault instance
func TestVaultClientIntegration(t *testing.T) {
	// Skip integration tests if Vault is not available
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("Skipping integration tests - VAULT_ADDR or VAULT_TOKEN not set")
	}

	client, err := NewVaultClientFromEnv()
	require.NoError(t, err)
	require.NotNil(t, client)

	t.Run("HealthCheck", func(t *testing.T) {
		err := client.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("GetSecret - existing secret", func(t *testing.T) {
		// This test assumes the secret exists (created in task 3.1)
		secret, err := client.GetSecret("secret/data/tennis-bot/lta")
		assert.NoError(t, err)
		assert.NotNil(t, secret)
		
		// Verify expected fields exist
		assert.Contains(t, secret, "username")
		assert.Contains(t, secret, "password")
		assert.Contains(t, secret, "api_key")
		assert.Contains(t, secret, "base_url")
	})

	t.Run("GetSecret - non-existent secret", func(t *testing.T) {
		_, err := client.GetSecret("secret/data/tennis-bot/nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found")
	})

	t.Run("GetSecretField - existing field", func(t *testing.T) {
		username, err := client.GetSecretField("secret/data/tennis-bot/lta", "username")
		assert.NoError(t, err)
		assert.Equal(t, "lta_test_user", username)
	})

	t.Run("GetSecretField - non-existent field", func(t *testing.T) {
		_, err := client.GetSecretField("secret/data/tennis-bot/lta", "nonexistent_field")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "field nonexistent_field not found")
	})
} 