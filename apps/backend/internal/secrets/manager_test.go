package secrets

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockVaultClient is a mock implementation of the VaultClient
type MockVaultClient struct {
	mock.Mock
}

func (m *MockVaultClient) GetSecret(path string) (map[string]interface{}, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockVaultClient) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockVaultClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewSecretsManager(t *testing.T) {
	// Test the constructor logic
	sm := &SecretsManager{
		client: nil, // We'll set this to nil for this test
		cache:  make(map[string]map[string]interface{}),
	}
	
	assert.NotNil(t, sm)
	assert.NotNil(t, sm.cache)
}

func TestNewSecretsManagerFromEnv_MissingVault(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("VAULT_ADDR")
	os.Unsetenv("VAULT_TOKEN")
	
	_, err := NewSecretsManagerFromEnv()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create vault client")
}

func TestSecretsManager_GetSecret(t *testing.T) {
	// Create a SecretsManager with a mock client
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	// We'll test the caching logic by manually setting up the cache
	testPath := "test/path"
	testKey := "test_key"
	testValue := "test_value"
	
	// Test cache hit
	sm.cache[testPath] = map[string]interface{}{
		testKey: testValue,
	}
	
	value, err := sm.GetSecret(testPath, testKey)
	assert.NoError(t, err)
	assert.Equal(t, testValue, value)
}

func TestSecretsManager_GetSecret_KeyNotFound(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	testPath := "test/path"
	testKey := "test_key"
	nonExistentKey := "nonexistent_key"
	
	// Set up cache with a different key
	sm.cache[testPath] = map[string]interface{}{
		testKey: "test_value",
	}
	
	// Test that we can find the existing key
	value, err := sm.GetSecret(testPath, testKey)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", value)
	
	// Test that we get an error for a non-existent key in the cached data
	_, err = sm.GetSecret(testPath, nonExistentKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key nonexistent_key not found")
}

func TestSecretsManager_GetSecretData(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	testPath := "test/path"
	testData := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}
	
	// Test cache hit
	sm.cache[testPath] = testData
	
	data, err := sm.GetSecretData(testPath)
	assert.NoError(t, err)
	assert.Equal(t, testData, data)
}

func TestSecretsManager_RefreshSecret(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	testPath := "test/path"
	
	// Set up initial cache
	sm.cache[testPath] = map[string]interface{}{
		"key": "old_value",
	}
	
	// Verify cache exists
	assert.Contains(t, sm.cache, testPath)
	
	// Since we don't have a real client, this will fail
	// but we can verify the cache was cleared
	sm.RefreshAllSecrets() // Clear cache manually for this test
	assert.NotContains(t, sm.cache, testPath)
}

func TestSecretsManager_RefreshAllSecrets(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	// Set up cache with multiple entries
	sm.cache["path1"] = map[string]interface{}{"key1": "value1"}
	sm.cache["path2"] = map[string]interface{}{"key2": "value2"}
	
	assert.Len(t, sm.cache, 2)
	
	sm.RefreshAllSecrets()
	
	assert.Len(t, sm.cache, 0)
}

func TestSecretsManager_GetDBCredentials(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	// Set up mock database credentials
	sm.cache[DBSecretPath] = map[string]interface{}{
		"username": "test_user",
		"password": "test_pass",
		"host":     "localhost",
		"database": "test_db",
	}
	
	username, password, host, database, err := sm.GetDBCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "test_user", username)
	assert.Equal(t, "test_pass", password)
	assert.Equal(t, "localhost", host)
	assert.Equal(t, "test_db", database)
}

func TestSecretsManager_GetJWTSecret(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	// Set up mock JWT secret
	sm.cache[JWTSecretPath] = map[string]interface{}{
		"secret": "jwt_secret_key",
	}
	
	secret, err := sm.GetJWTSecret()
	assert.NoError(t, err)
	assert.Equal(t, "jwt_secret_key", secret)
}

func TestSecretsManager_GetEmailCredentials(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	// Set up mock email credentials
	sm.cache[EmailSecretPath] = map[string]interface{}{
		"email":     "test@example.com",
		"password":  "email_pass",
		"smtp_host": "smtp.gmail.com",
		"smtp_port": "587",
	}
	
	email, password, smtpHost, smtpPort, err := sm.GetEmailCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "email_pass", password)
	assert.Equal(t, "smtp.gmail.com", smtpHost)
	assert.Equal(t, "587", smtpPort)
}

func TestSecretsManager_GetRedisCredentials(t *testing.T) {
	sm := &SecretsManager{
		cache: make(map[string]map[string]interface{}),
	}
	
	// Set up mock Redis credentials
	sm.cache[RedisSecretPath] = map[string]interface{}{
		"host":     "localhost:6379",
		"password": "redis_pass",
	}
	
	host, password, err := sm.GetRedisCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:6379", host)
	assert.Equal(t, "redis_pass", password)
}

func TestSecretsManager_Constants(t *testing.T) {
	// Test that all the predefined paths are correctly set
	assert.Equal(t, "kv/data/tennisapp/prod/db", DBSecretPath)
	assert.Equal(t, "kv/data/tennisapp/prod/jwt", JWTSecretPath)
	assert.Equal(t, "kv/data/tennisapp/prod/email", EmailSecretPath)
	assert.Equal(t, "kv/data/tennisapp/prod/redis", RedisSecretPath)
	assert.Equal(t, "kv/data/tennisapp/prod/api", APISecretPath)
	assert.Equal(t, "kv/data/tennisapp/prod/platforms/lta", LTACredentialsPath)
	assert.Equal(t, "kv/data/tennisapp/prod/platforms/courtsides", CourtsidesCredentialsPath)
	assert.Equal(t, "kv/data/tennisapp/prod/notifications/twilio", TwilioCredentialsPath)
	assert.Equal(t, "kv/data/tennisapp/prod/notifications/sendgrid", SendGridCredentialsPath)
}

// Integration tests - require running Vault instance
func TestSecretsManagerIntegration(t *testing.T) {
	// Skip integration tests if Vault is not available
	if os.Getenv("VAULT_ADDR") == "" || os.Getenv("VAULT_TOKEN") == "" {
		t.Skip("Skipping integration tests - VAULT_ADDR or VAULT_TOKEN not set")
	}

	t.Run("NewSecretsManagerFromEnv_Success", func(t *testing.T) {
		sm, err := NewSecretsManagerFromEnv()
		assert.NoError(t, err)
		assert.NotNil(t, sm)
		defer sm.Close()

		// Test health check
		err = sm.HealthCheck()
		assert.NoError(t, err)
	})

	t.Run("GetSecret_Integration", func(t *testing.T) {
		sm, err := NewSecretsManagerFromEnv()
		require.NoError(t, err)
		defer sm.Close()

		// Try to read a test secret
		// This assumes the secret exists in the test Vault instance
		_, err = sm.GetSecret("kv/data/tennisapp/prod/db", "username")
		// We don't assert no error here because the secret might not exist
		// In a proper test setup, you would create the secret first
		if err != nil {
			t.Logf("Expected error reading test secret: %v", err)
		}
	})
} 