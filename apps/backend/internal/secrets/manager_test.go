package secrets

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecretsManager(t *testing.T) {
	sm := NewSecretsManager()
	assert.NotNil(t, sm)
	assert.NotNil(t, sm.cache)
}

func TestNewSecretsManagerFromEnv(t *testing.T) {
	sm, err := NewSecretsManagerFromEnv()
	assert.NoError(t, err)
	assert.NotNil(t, sm)
}

func TestSecretsManager_GetSecret(t *testing.T) {
	sm := NewSecretsManager()

	// Test environment variable that doesn't exist
	_, err := sm.GetSecret("NON_EXISTENT_ENV_VAR")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable NON_EXISTENT_ENV_VAR not found")

	// Test environment variable that exists
	testEnvVar := "TEST_SECRET"
	testValue := "test_value_123"
	
	// Set environment variable
	os.Setenv(testEnvVar, testValue)
	defer os.Unsetenv(testEnvVar)

	value, err := sm.GetSecret(testEnvVar)
	assert.NoError(t, err)
	assert.Equal(t, testValue, value)

	// Test that it uses cache on second call
	value2, err := sm.GetSecret(testEnvVar)
	assert.NoError(t, err)
	assert.Equal(t, testValue, value2)
}

func TestSecretsManager_RefreshSecret(t *testing.T) {
	sm := NewSecretsManager()

	testEnvVar := "TEST_SECRET_REFRESH"
	testValue := "initial_value"
	
	// Set initial environment variable
	os.Setenv(testEnvVar, testValue)
	defer os.Unsetenv(testEnvVar)

	// Get secret to populate cache
	value, err := sm.GetSecret(testEnvVar)
	assert.NoError(t, err)
	assert.Equal(t, testValue, value)

	// Verify cache contains the value
	assert.Contains(t, sm.cache, testEnvVar)

	// Refresh secret (clears cache)
	sm.RefreshSecret(testEnvVar)

	// Verify cache no longer contains the value
	assert.NotContains(t, sm.cache, testEnvVar)
}

func TestSecretsManager_RefreshAllSecrets(t *testing.T) {
	sm := NewSecretsManager()

	// Set up multiple environment variables and populate cache
	testVars := map[string]string{
		"TEST_VAR_1": "value1",
		"TEST_VAR_2": "value2",
	}

	for envVar, value := range testVars {
		os.Setenv(envVar, value)
		defer os.Unsetenv(envVar)
		
		// Populate cache
		_, err := sm.GetSecret(envVar)
		assert.NoError(t, err)
	}

	assert.Len(t, sm.cache, 2)

	sm.RefreshAllSecrets()

	assert.Len(t, sm.cache, 0)
}

func TestSecretsManager_HealthCheck(t *testing.T) {
	sm := NewSecretsManager()
	err := sm.HealthCheck()
	assert.NoError(t, err) // Environment variables don't need health checks
}

func TestSecretsManager_Close(t *testing.T) {
	sm := NewSecretsManager()
	err := sm.Close()
	assert.NoError(t, err) // Environment variables don't need closing
}

func TestSecretsManager_GetDBCredentials(t *testing.T) {
	sm := NewSecretsManager()

	t.Run("Individual credentials", func(t *testing.T) {
		// Set up individual MongoDB credentials
		os.Setenv(MongoUsernameEnv, "test_user")
		os.Setenv(MongoPasswordEnv, "test_pass")
		os.Setenv(MongoHostEnv, "localhost")
		os.Setenv(MongoDatabaseEnv, "test_db")
		
		defer func() {
			os.Unsetenv(MongoUsernameEnv)
			os.Unsetenv(MongoPasswordEnv)
			os.Unsetenv(MongoHostEnv)
			os.Unsetenv(MongoDatabaseEnv)
		}()

		username, password, host, database, err := sm.GetDBCredentials()
		assert.NoError(t, err)
		assert.Equal(t, "test_user", username)
		assert.Equal(t, "test_pass", password)
		assert.Equal(t, "localhost", host)
		assert.Equal(t, "test_db", database)
	})

	t.Run("MONGO_URI fallback", func(t *testing.T) {
		// Clear individual credentials
		os.Unsetenv(MongoUsernameEnv)
		os.Unsetenv(MongoPasswordEnv)
		os.Unsetenv(MongoHostEnv)
		os.Unsetenv(MongoDatabaseEnv)

		// Set MONGO_URI
		testURI := "mongodb://admin:password@localhost:27017/testdb"
		os.Setenv(MongoURIEnv, testURI)
		defer os.Unsetenv(MongoURIEnv)

		username, password, host, database, err := sm.GetDBCredentials()
		assert.NoError(t, err)
		assert.Equal(t, "", username)
		assert.Equal(t, "", password)
		assert.Equal(t, testURI, host) // URI returned as host
		assert.Equal(t, "", database)
	})

	t.Run("No credentials", func(t *testing.T) {
		// Clear all credentials
		os.Unsetenv(MongoUsernameEnv)
		os.Unsetenv(MongoPasswordEnv)
		os.Unsetenv(MongoHostEnv)
		os.Unsetenv(MongoDatabaseEnv)
		os.Unsetenv(MongoURIEnv)

		_, _, _, _, err := sm.GetDBCredentials()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "neither individual MongoDB credentials nor MONGO_URI found")
	})
}

func TestSecretsManager_GetJWTSecret(t *testing.T) {
	sm := NewSecretsManager()

	// Test missing JWT secret
	os.Unsetenv(JWTSecretEnv)
	_, err := sm.GetJWTSecret()
	assert.Error(t, err)

	// Test valid JWT secret
	testSecret := "jwt_secret_key_123"
	os.Setenv(JWTSecretEnv, testSecret)
	defer os.Unsetenv(JWTSecretEnv)

	secret, err := sm.GetJWTSecret()
	assert.NoError(t, err)
	assert.Equal(t, testSecret, secret)
}

func TestSecretsManager_GetEmailCredentials(t *testing.T) {
	sm := NewSecretsManager()

	// Set up email credentials
	os.Setenv(EmailAddressEnv, "test@example.com")
	os.Setenv(EmailPasswordEnv, "email_pass")
	os.Setenv(SMTPHostEnv, "smtp.gmail.com")
	os.Setenv(SMTPPortEnv, "587")
	
	defer func() {
		os.Unsetenv(EmailAddressEnv)
		os.Unsetenv(EmailPasswordEnv)
		os.Unsetenv(SMTPHostEnv)
		os.Unsetenv(SMTPPortEnv)
	}()

	email, password, smtpHost, smtpPort, err := sm.GetEmailCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "email_pass", password)
	assert.Equal(t, "smtp.gmail.com", smtpHost)
	assert.Equal(t, "587", smtpPort)
}

func TestSecretsManager_GetRedisCredentials(t *testing.T) {
	sm := NewSecretsManager()

	// Set up Redis credentials
	os.Setenv(RedisAddrEnv, "localhost:6379")
	os.Setenv(RedisPasswordEnv, "redis_pass")
	
	defer func() {
		os.Unsetenv(RedisAddrEnv)
		os.Unsetenv(RedisPasswordEnv)
	}()

	addr, password, err := sm.GetRedisCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:6379", addr)
	assert.Equal(t, "redis_pass", password)

	// Test without password (optional)
	os.Unsetenv(RedisPasswordEnv)
	addr, password, err = sm.GetRedisCredentials()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:6379", addr)
	assert.Equal(t, "", password) // Password should be empty
}

func TestSecretsManager_Constants(t *testing.T) {
	// Test that all the environment variable constants are correctly set
	assert.Equal(t, "MONGO_URI", MongoURIEnv)
	assert.Equal(t, "MONGO_USERNAME", MongoUsernameEnv)
	assert.Equal(t, "MONGO_PASSWORD", MongoPasswordEnv)
	assert.Equal(t, "MONGO_HOST", MongoHostEnv)
	assert.Equal(t, "MONGO_DATABASE", MongoDatabaseEnv)
	assert.Equal(t, "JWT_SECRET", JWTSecretEnv)
	assert.Equal(t, "EMAIL_ADDRESS", EmailAddressEnv)
	assert.Equal(t, "EMAIL_PASSWORD", EmailPasswordEnv)
	assert.Equal(t, "SMTP_HOST", SMTPHostEnv)
	assert.Equal(t, "SMTP_PORT", SMTPPortEnv)
	assert.Equal(t, "REDIS_ADDR", RedisAddrEnv)
	assert.Equal(t, "REDIS_PASSWORD", RedisPasswordEnv)
	assert.Equal(t, "LTA_USERNAME", LTAUsernameEnv)
	assert.Equal(t, "LTA_PASSWORD", LTAPasswordEnv)
	assert.Equal(t, "COURTSIDE_USERNAME", CourtsideUsernameEnv)
	assert.Equal(t, "COURTSIDE_PASSWORD", CourtsidePasswordEnv)
	assert.Equal(t, "TWILIO_SID", TwilioSIDEnv)
	assert.Equal(t, "TWILIO_TOKEN", TwilioTokenEnv)
	assert.Equal(t, "SENDGRID_API_KEY", SendGridAPIKeyEnv)
}