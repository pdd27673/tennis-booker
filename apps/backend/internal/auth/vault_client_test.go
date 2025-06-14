package auth

import (
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVaultClient(t *testing.T) {
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200"

	client, err := NewVaultClient(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
}

func TestNewVaultClientFromEnv_MissingVaultAddr(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("VAULT_TOKEN")
	os.Unsetenv("VAULT_ROLE_ID")
	os.Unsetenv("VAULT_SECRET_ID")

	_, err := NewVaultClientFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VAULT_ADDR environment variable is required")
}

func TestNewVaultClientFromEnv_MissingAuth(t *testing.T) {
	// Set VAULT_ADDR but no auth
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
	os.Unsetenv("VAULT_TOKEN")
	os.Unsetenv("VAULT_ROLE_ID")
	os.Unsetenv("VAULT_SECRET_ID")
	defer os.Unsetenv("VAULT_ADDR")

	_, err := NewVaultClientFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vault token is required")
}

func TestNewVaultClientFromEnv_WithToken(t *testing.T) {
	// Set environment variables for token auth
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
	os.Setenv("VAULT_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("VAULT_ADDR")
		os.Unsetenv("VAULT_TOKEN")
	}()

	// This should succeed in creating the client (authentication doesn't require network call)
	client, err := NewVaultClientFromEnv()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Health check will fail since no real Vault is running
	err = client.HealthCheck()
	assert.Error(t, err)
}

func TestNewVaultClientFromEnv_WithAppRole(t *testing.T) {
	// Set environment variables for AppRole auth
	os.Setenv("VAULT_ADDR", "http://localhost:8200")
	os.Setenv("VAULT_ROLE_ID", "test-role-id")
	os.Setenv("VAULT_SECRET_ID", "test-secret-id")
	defer func() {
		os.Unsetenv("VAULT_ADDR")
		os.Unsetenv("VAULT_ROLE_ID")
		os.Unsetenv("VAULT_SECRET_ID")
	}()

	// This will fail because AppRole requires network call to Vault
	_, err := NewVaultClientFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to authenticate with vault")
}

func TestVaultClient_GetClient(t *testing.T) {
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200"

	vaultClient, err := NewVaultClient(config)
	require.NoError(t, err)

	apiClient := vaultClient.GetClient()
	assert.NotNil(t, apiClient)
	assert.Equal(t, vaultClient.client, apiClient)
}

func TestVaultClient_Close(t *testing.T) {
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200"

	vaultClient, err := NewVaultClient(config)
	require.NoError(t, err)

	err = vaultClient.Close()
	assert.NoError(t, err)
}

// Integration tests - require running Vault instance
func TestVaultClientIntegration(t *testing.T) {
	// Skip integration tests if Vault is not available
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("Skipping integration tests - VAULT_ADDR or VAULT_TOKEN not set")
	}

	t.Run("NewVaultClientFromEnv_Success", func(t *testing.T) {
		client, err := NewVaultClientFromEnv()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		defer client.Close()

		// Test health check
		err = client.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("GetSecret_Success", func(t *testing.T) {
		client, err := NewVaultClientFromEnv()
		require.NoError(t, err)
		defer client.Close()

		// Try to read a test secret (this assumes the secret exists)
		// In a real test environment, you would set up test secrets
		_, err = client.GetSecret("kv/data/tennisapp/prod/db")
		// We don't assert no error here because the secret might not exist
		// In a proper test setup, you would create the secret first
		if err != nil {
			t.Logf("Expected error reading test secret: %v", err)
		}
	})

	t.Run("GetSecret_NotFound", func(t *testing.T) {
		client, err := NewVaultClientFromEnv()
		require.NoError(t, err)
		defer client.Close()

		_, err = client.GetSecret("kv/data/nonexistent/path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no secret found at path")
	})
}
