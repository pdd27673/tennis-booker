package secrets

import (
	"fmt"
	"os"
	"sync"
)

// SecretsManager provides methods to fetch secrets from environment variables
type SecretsManager struct {
	cache map[string]string
	mutex sync.RWMutex
}

// NewSecretsManager creates a new SecretsManager
func NewSecretsManager() *SecretsManager {
	return &SecretsManager{
		cache: make(map[string]string),
	}
}

// NewSecretsManagerFromEnv creates a new SecretsManager using environment variables
func NewSecretsManagerFromEnv() (*SecretsManager, error) {
	return NewSecretsManager(), nil
}

// GetSecret retrieves a secret value from environment variables
func (sm *SecretsManager) GetSecret(envKey string) (string, error) {
	// Check cache first
	sm.mutex.RLock()
	if value, exists := sm.cache[envKey]; exists {
		sm.mutex.RUnlock()
		return value, nil
	}
	sm.mutex.RUnlock()

	// Get from environment
	value := os.Getenv(envKey)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not found or empty", envKey)
	}

	// Cache the value
	sm.mutex.Lock()
	sm.cache[envKey] = value
	sm.mutex.Unlock()

	return value, nil
}

// RefreshSecret clears the cache for a specific environment variable
func (sm *SecretsManager) RefreshSecret(envKey string) {
	sm.mutex.Lock()
	delete(sm.cache, envKey)
	sm.mutex.Unlock()
}

// RefreshAllSecrets clears the entire cache
func (sm *SecretsManager) RefreshAllSecrets() {
	sm.mutex.Lock()
	sm.cache = make(map[string]string)
	sm.mutex.Unlock()
}

// HealthCheck always returns nil since environment variables don't need health checks
func (sm *SecretsManager) HealthCheck() error {
	return nil
}

// Close is a no-op for environment variables
func (sm *SecretsManager) Close() error {
	return nil
}

// Environment variable names for common secrets
const (
	// Database environment variables
	MongoURIEnv      = "MONGO_URI"
	MongoUsernameEnv = "MONGO_USERNAME"
	MongoPasswordEnv = "MONGO_PASSWORD"
	MongoHostEnv     = "MONGO_HOST"
	MongoDatabaseEnv = "MONGO_DATABASE"

	// JWT environment variables
	JWTSecretEnv = "JWT_SECRET"

	// Email environment variables
	EmailAddressEnv  = "EMAIL_ADDRESS"
	EmailPasswordEnv = "EMAIL_PASSWORD"
	SMTPHostEnv      = "SMTP_HOST"
	SMTPPortEnv      = "SMTP_PORT"

	// Redis environment variables
	RedisAddrEnv     = "REDIS_ADDR"
	RedisPasswordEnv = "REDIS_PASSWORD"

	// Platform credentials
	LTAUsernameEnv        = "LTA_USERNAME"
	LTAPasswordEnv        = "LTA_PASSWORD"
	CourtsideUsernameEnv  = "COURTSIDE_USERNAME"
	CourtsidePasswordEnv  = "COURTSIDE_PASSWORD"

	// Notification services
	TwilioSIDEnv        = "TWILIO_SID"
	TwilioTokenEnv      = "TWILIO_TOKEN"
	SendGridAPIKeyEnv   = "SENDGRID_API_KEY"
)

// Convenience methods for common secrets

// GetDBCredentials retrieves database connection credentials from environment variables
func (sm *SecretsManager) GetDBCredentials() (username, password, host, database string, err error) {
	// Try to get individual components first
	username, _ = sm.GetSecret(MongoUsernameEnv)
	password, _ = sm.GetSecret(MongoPasswordEnv)
	host, _ = sm.GetSecret(MongoHostEnv)
	database, _ = sm.GetSecret(MongoDatabaseEnv)

	// If we have all components, return them
	if username != "" && password != "" && host != "" && database != "" {
		return username, password, host, database, nil
	}

	// Otherwise, we expect MONGO_URI to be set
	mongoURI, err := sm.GetSecret(MongoURIEnv)
	if err != nil {
		return "", "", "", "", fmt.Errorf("neither individual MongoDB credentials nor MONGO_URI found: %w", err)
	}

	// For MONGO_URI, we'll return it as the host and let the database layer handle parsing
	return "", "", mongoURI, "", nil
}

// GetJWTSecret retrieves the JWT signing secret
func (sm *SecretsManager) GetJWTSecret() (string, error) {
	return sm.GetSecret(JWTSecretEnv)
}

// GetEmailCredentials retrieves email service credentials
func (sm *SecretsManager) GetEmailCredentials() (email, password, smtpHost, smtpPort string, err error) {
	email, err = sm.GetSecret(EmailAddressEnv)
	if err != nil {
		return "", "", "", "", err
	}

	password, err = sm.GetSecret(EmailPasswordEnv)
	if err != nil {
		return "", "", "", "", err
	}

	smtpHost, err = sm.GetSecret(SMTPHostEnv)
	if err != nil {
		return "", "", "", "", err
	}

	smtpPort, err = sm.GetSecret(SMTPPortEnv)
	if err != nil {
		return "", "", "", "", err
	}

	return email, password, smtpHost, smtpPort, nil
}

// GetRedisCredentials retrieves Redis connection credentials
func (sm *SecretsManager) GetRedisCredentials() (addr, password string, err error) {
	addr, err = sm.GetSecret(RedisAddrEnv)
	if err != nil {
		return "", "", err
	}

	password, err = sm.GetSecret(RedisPasswordEnv)
	if err != nil {
		// Redis password is optional
		password = ""
	}

	return addr, password, nil
}