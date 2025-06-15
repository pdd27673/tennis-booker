package secrets

import (
	"fmt"
	"sync"

	"tennis-booker/internal/auth"
)

// SecretsManager encapsulates the Vault client and provides methods to fetch secrets
type SecretsManager struct {
	client *auth.VaultClient
	cache  map[string]map[string]interface{}
	mutex  sync.RWMutex
}

// NewSecretsManager creates a new SecretsManager with the provided Vault client
func NewSecretsManager(client *auth.VaultClient) *SecretsManager {
	return &SecretsManager{
		client: client,
		cache:  make(map[string]map[string]interface{}),
	}
}

// NewSecretsManagerFromEnv creates a new SecretsManager using environment variables
func NewSecretsManagerFromEnv() (*SecretsManager, error) {
	client, err := auth.NewVaultClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return NewSecretsManager(client), nil
}

// GetSecret retrieves a specific secret value by path and key
func (sm *SecretsManager) GetSecret(path string, key string) (string, error) {
	// Check cache first
	sm.mutex.RLock()
	if secretData, exists := sm.cache[path]; exists {
		if value, keyExists := secretData[key]; keyExists {
			sm.mutex.RUnlock()
			if strValue, ok := value.(string); ok {
				return strValue, nil
			}
			return fmt.Sprintf("%v", value), nil
		}
		// Key not found in cached data
		sm.mutex.RUnlock()
		return "", fmt.Errorf("key %s not found in secret at path %s", key, path)
	}
	sm.mutex.RUnlock()

	// Check if client is available
	if sm.client == nil {
		return "", fmt.Errorf("vault client not available")
	}

	// Fetch from Vault
	secretData, err := sm.client.GetSecret(path)
	if err != nil {
		return "", fmt.Errorf("failed to fetch secret from path %s: %w", path, err)
	}

	// Cache the entire secret data
	sm.mutex.Lock()
	sm.cache[path] = secretData
	sm.mutex.Unlock()

	// Extract the specific key
	if value, exists := secretData[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue, nil
		}
		return fmt.Sprintf("%v", value), nil
	}

	return "", fmt.Errorf("key %s not found in secret at path %s", key, path)
}

// GetSecretData retrieves all secret data from a path
func (sm *SecretsManager) GetSecretData(path string) (map[string]interface{}, error) {
	// Check cache first
	sm.mutex.RLock()
	if secretData, exists := sm.cache[path]; exists {
		sm.mutex.RUnlock()
		return secretData, nil
	}
	sm.mutex.RUnlock()

	// Check if client is available
	if sm.client == nil {
		return nil, fmt.Errorf("vault client not available")
	}

	// Fetch from Vault
	secretData, err := sm.client.GetSecret(path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret from path %s: %w", path, err)
	}

	// Cache the secret data
	sm.mutex.Lock()
	sm.cache[path] = secretData
	sm.mutex.Unlock()

	return secretData, nil
}

// RefreshSecret clears the cache for a specific path and forces a reload
func (sm *SecretsManager) RefreshSecret(path string) error {
	sm.mutex.Lock()
	delete(sm.cache, path)
	sm.mutex.Unlock()

	// Force reload by fetching the secret
	_, err := sm.GetSecretData(path)
	return err
}

// RefreshAllSecrets clears the entire cache
func (sm *SecretsManager) RefreshAllSecrets() {
	sm.mutex.Lock()
	sm.cache = make(map[string]map[string]interface{})
	sm.mutex.Unlock()
}

// HealthCheck verifies the Vault connection
func (sm *SecretsManager) HealthCheck() error {
	return sm.client.HealthCheck()
}

// Close closes the underlying Vault client
func (sm *SecretsManager) Close() error {
	return sm.client.Close()
}

// GetClient returns the underlying Vault client
func (sm *SecretsManager) GetClient() *auth.VaultClient {
	return sm.client
}

// Predefined secret paths for the tennis app
const (
	// Database secrets
	DBSecretPath = "secret/data/tennisapp/prod/db"

	// JWT secrets
	JWTSecretPath = "secret/data/tennisapp/prod/jwt"

	// Email secrets
	EmailSecretPath = "secret/data/tennisapp/prod/email"

	// Redis secrets
	RedisSecretPath = "secret/data/tennisapp/prod/redis"

	// API secrets
	APISecretPath = "secret/data/tennisapp/prod/api"

	// Platform credentials
	LTACredentialsPath        = "secret/data/tennisapp/prod/platforms/lta"
	CourtsidesCredentialsPath = "secret/data/tennisapp/prod/platforms/courtsides"

	// Notification services
	TwilioCredentialsPath   = "secret/data/tennisapp/prod/notifications/twilio"
	SendGridCredentialsPath = "secret/data/tennisapp/prod/notifications/sendgrid"
)

// Convenience methods for common secrets

// GetDBCredentials retrieves database connection credentials
func (sm *SecretsManager) GetDBCredentials() (username, password, host, database string, err error) {
	secretData, err := sm.GetSecretData(DBSecretPath)
	if err != nil {
		return "", "", "", "", err
	}

	username, _ = secretData["username"].(string)
	password, _ = secretData["password"].(string)
	host, _ = secretData["host"].(string)
	database, _ = secretData["database"].(string)

	return username, password, host, database, nil
}

// GetJWTSecret retrieves the JWT signing secret
func (sm *SecretsManager) GetJWTSecret() (string, error) {
	return sm.GetSecret(JWTSecretPath, "secret")
}

// GetEmailCredentials retrieves email service credentials
func (sm *SecretsManager) GetEmailCredentials() (email, password, smtpHost, smtpPort string, err error) {
	secretData, err := sm.GetSecretData(EmailSecretPath)
	if err != nil {
		return "", "", "", "", err
	}

	email, _ = secretData["email"].(string)
	password, _ = secretData["password"].(string)
	smtpHost, _ = secretData["smtp_host"].(string)
	smtpPort, _ = secretData["smtp_port"].(string)

	return email, password, smtpHost, smtpPort, nil
}

// GetRedisCredentials retrieves Redis connection credentials
func (sm *SecretsManager) GetRedisCredentials() (host, password string, err error) {
	secretData, err := sm.GetSecretData(RedisSecretPath)
	if err != nil {
		return "", "", err
	}

	host, _ = secretData["host"].(string)
	password, _ = secretData["password"].(string)

	return host, password, nil
}
