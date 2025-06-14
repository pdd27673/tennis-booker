package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformCredentials(t *testing.T) {
	creds := &PlatformCredentials{
		Username:   "test_user",
		Password:   "test_pass",
		APIKey:     "test_key",
		BaseURL:    "https://test.com",
		LoginURL:   "https://test.com/login",
		BookingURL: "https://test.com/book",
	}

	assert.Equal(t, "test_user", creds.Username)
	assert.Equal(t, "test_pass", creds.Password)
	assert.Equal(t, "test_key", creds.APIKey)
	assert.Equal(t, "https://test.com", creds.BaseURL)
	assert.Equal(t, "https://test.com/login", creds.LoginURL)
	assert.Equal(t, "https://test.com/book", creds.BookingURL)
}

func TestNewCredentialsManager(t *testing.T) {
	// Save original environment variables
	originalVaultAddr := os.Getenv("VAULT_ADDR")
	originalVaultToken := os.Getenv("VAULT_TOKEN")
	originalRoleID := os.Getenv("VAULT_ROLE_ID")
	originalSecretID := os.Getenv("VAULT_SECRET_ID")

	// Clean up after test
	defer func() {
		if originalVaultAddr != "" {
			os.Setenv("VAULT_ADDR", originalVaultAddr)
		} else {
			os.Unsetenv("VAULT_ADDR")
		}
		if originalVaultToken != "" {
			os.Setenv("VAULT_TOKEN", originalVaultToken)
		} else {
			os.Unsetenv("VAULT_TOKEN")
		}
		if originalRoleID != "" {
			os.Setenv("VAULT_ROLE_ID", originalRoleID)
		} else {
			os.Unsetenv("VAULT_ROLE_ID")
		}
		if originalSecretID != "" {
			os.Setenv("VAULT_SECRET_ID", originalSecretID)
		} else {
			os.Unsetenv("VAULT_SECRET_ID")
		}
	}()

	t.Run("MissingVaultAddr", func(t *testing.T) {
		// Test without VAULT_ADDR
		os.Unsetenv("VAULT_ADDR")
		os.Unsetenv("VAULT_TOKEN")
		os.Unsetenv("VAULT_ROLE_ID")
		os.Unsetenv("VAULT_SECRET_ID")

		_, err := NewCredentialsManager()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "VAULT_ADDR environment variable is required")
	})

	t.Run("MissingAuthCredentials", func(t *testing.T) {
		// Test with VAULT_ADDR but no auth credentials
		os.Setenv("VAULT_ADDR", "http://localhost:8200")
		os.Unsetenv("VAULT_TOKEN")
		os.Unsetenv("VAULT_ROLE_ID")
		os.Unsetenv("VAULT_SECRET_ID")

		// Use a channel to implement timeout
		done := make(chan bool, 1)
		var err error

		go func() {
			_, err = NewCredentialsManager()
			done <- true
		}()

		select {
		case <-done:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "vault token is required")
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out - this indicates the vault client is trying to connect")
		}
	})
}

// Integration tests - require running Vault instance
func TestCredentialsManagerIntegration(t *testing.T) {
	// Skip integration tests if Vault is not available
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("Skipping integration tests - VAULT_ADDR or VAULT_TOKEN not set")
	}

	cm, err := NewCredentialsManager()
	require.NoError(t, err)
	require.NotNil(t, cm)
	defer cm.Close()

	t.Run("HealthCheck", func(t *testing.T) {
		err := cm.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("GetLTACredentials", func(t *testing.T) {
		creds, err := cm.GetLTACredentials()
		assert.NoError(t, err)
		assert.NotNil(t, creds)
		assert.Equal(t, "lta_test_user", creds.Username)
		assert.Equal(t, "lta_test_password", creds.Password)
		assert.Equal(t, "lta_test_api_key", creds.APIKey)
		assert.Equal(t, "https://lta.booking.com", creds.BaseURL)
	})

	t.Run("GetCourtsidesCredentials", func(t *testing.T) {
		creds, err := cm.GetCourtsidesCredentials()
		assert.NoError(t, err)
		assert.NotNil(t, creds)
		assert.Equal(t, "courtsides_test_user", creds.Username)
		assert.Equal(t, "courtsides_test_password", creds.Password)
		assert.Equal(t, "https://courtsides.com/login", creds.LoginURL)
		assert.Equal(t, "https://courtsides.com/book", creds.BookingURL)
	})

	t.Run("GetPlatformCredentials", func(t *testing.T) {
		creds, err := cm.GetPlatformCredentials("lta")
		assert.NoError(t, err)
		assert.NotNil(t, creds)
		assert.Equal(t, "lta_test_user", creds.Username)
	})

	t.Run("CachingBehavior", func(t *testing.T) {
		// First call should load from Vault
		creds1, err := cm.GetPlatformCredentials("lta")
		assert.NoError(t, err)

		// Second call should use cache (same instance)
		creds2, err := cm.GetPlatformCredentials("lta")
		assert.NoError(t, err)
		assert.Equal(t, creds1, creds2) // Should be the same instance due to caching
	})

	t.Run("RefreshCredentials", func(t *testing.T) {
		// Load credentials first
		_, err := cm.GetPlatformCredentials("lta")
		assert.NoError(t, err)

		// Refresh should work without error
		err = cm.RefreshCredentials("lta")
		assert.NoError(t, err)
	})

	t.Run("RefreshAllCredentials", func(t *testing.T) {
		// Load some credentials first
		_, err := cm.GetPlatformCredentials("lta")
		assert.NoError(t, err)
		_, err = cm.GetPlatformCredentials("courtsides")
		assert.NoError(t, err)

		// Refresh all should not cause errors
		cm.RefreshAllCredentials()

		// Should be able to load again
		_, err = cm.GetPlatformCredentials("lta")
		assert.NoError(t, err)
	})

	t.Run("NonExistentPlatform", func(t *testing.T) {
		_, err := cm.GetPlatformCredentials("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get nonexistent credentials")
	})
}

func TestGlobalCredentialsManager(t *testing.T) {
	// Skip integration tests if Vault is not available
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("Skipping integration tests - VAULT_ADDR or VAULT_TOKEN not set")
	}

	t.Run("GetGlobalCredentialsManager", func(t *testing.T) {
		cm1, err := GetGlobalCredentialsManager()
		assert.NoError(t, err)
		assert.NotNil(t, cm1)

		// Should return the same instance
		cm2, err := GetGlobalCredentialsManager()
		assert.NoError(t, err)
		assert.Equal(t, cm1, cm2)
	})

	t.Run("InitializeCredentials", func(t *testing.T) {
		err := InitializeCredentials("lta", "courtsides")
		assert.NoError(t, err)
	})

	t.Run("InitializeCredentialsWithInvalidPlatform", func(t *testing.T) {
		// Should not fail even with invalid platform (just logs warning)
		err := InitializeCredentials("lta", "invalid_platform")
		assert.NoError(t, err)
	})
}
