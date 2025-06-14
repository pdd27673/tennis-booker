package config

import (
	"fmt"
	"log"
	"sync"

	"tennis-booker/internal/auth"
)

// PlatformCredentials holds credentials for a booking platform
type PlatformCredentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	APIKey     string `json:"api_key,omitempty"`
	BaseURL    string `json:"base_url,omitempty"`
	LoginURL   string `json:"login_url,omitempty"`
	BookingURL string `json:"booking_url,omitempty"`
}

// CredentialsManager manages platform credentials from Vault
type CredentialsManager struct {
	vaultClient *auth.VaultClient
	cache       map[string]*PlatformCredentials
	mutex       sync.RWMutex
}

// NewCredentialsManager creates a new credentials manager
func NewCredentialsManager() (*CredentialsManager, error) {
	vaultClient, err := auth.NewVaultClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &CredentialsManager{
		vaultClient: vaultClient,
		cache:       make(map[string]*PlatformCredentials),
	}, nil
}

// GetPlatformCredentials retrieves credentials for a specific platform
func (cm *CredentialsManager) GetPlatformCredentials(platform string) (*PlatformCredentials, error) {
	// Check cache first
	cm.mutex.RLock()
	if creds, exists := cm.cache[platform]; exists {
		cm.mutex.RUnlock()
		return creds, nil
	}
	cm.mutex.RUnlock()

	// Load from Vault
	path := fmt.Sprintf("secret/data/tennis-bot/%s", platform)
	secretData, err := cm.vaultClient.GetSecret(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s credentials: %w", platform, err)
	}

	// Convert to PlatformCredentials struct
	creds := &PlatformCredentials{}

	if username, ok := secretData["username"].(string); ok {
		creds.Username = username
	}

	if password, ok := secretData["password"].(string); ok {
		creds.Password = password
	}

	if apiKey, ok := secretData["api_key"].(string); ok {
		creds.APIKey = apiKey
	}

	if baseURL, ok := secretData["base_url"].(string); ok {
		creds.BaseURL = baseURL
	}

	if loginURL, ok := secretData["login_url"].(string); ok {
		creds.LoginURL = loginURL
	}

	if bookingURL, ok := secretData["booking_url"].(string); ok {
		creds.BookingURL = bookingURL
	}

	// Cache the credentials
	cm.mutex.Lock()
	cm.cache[platform] = creds
	cm.mutex.Unlock()

	return creds, nil
}

// GetLTACredentials is a convenience method for LTA platform credentials
func (cm *CredentialsManager) GetLTACredentials() (*PlatformCredentials, error) {
	return cm.GetPlatformCredentials("lta")
}

// GetCourtsidesCredentials is a convenience method for Courtsides platform credentials
func (cm *CredentialsManager) GetCourtsidesCredentials() (*PlatformCredentials, error) {
	return cm.GetPlatformCredentials("courtsides")
}

// RefreshCredentials clears the cache and forces a reload from Vault
func (cm *CredentialsManager) RefreshCredentials(platform string) error {
	cm.mutex.Lock()
	delete(cm.cache, platform)
	cm.mutex.Unlock()

	_, err := cm.GetPlatformCredentials(platform)
	return err
}

// RefreshAllCredentials clears the entire cache
func (cm *CredentialsManager) RefreshAllCredentials() {
	cm.mutex.Lock()
	cm.cache = make(map[string]*PlatformCredentials)
	cm.mutex.Unlock()
}

// HealthCheck verifies the Vault connection
func (cm *CredentialsManager) HealthCheck() error {
	return cm.vaultClient.HealthCheck()
}

// Close closes the Vault client connection
func (cm *CredentialsManager) Close() error {
	return cm.vaultClient.Close()
}

// Global credentials manager instance
var (
	globalCredentialsManager *CredentialsManager
	globalCredentialsOnce    sync.Once
)

// GetGlobalCredentialsManager returns a singleton credentials manager
func GetGlobalCredentialsManager() (*CredentialsManager, error) {
	var err error
	globalCredentialsOnce.Do(func() {
		globalCredentialsManager, err = NewCredentialsManager()
		if err != nil {
			log.Printf("Failed to initialize global credentials manager: %v", err)
		}
	})
	return globalCredentialsManager, err
}

// InitializeCredentials initializes the global credentials manager and optionally preloads credentials
func InitializeCredentials(preloadPlatforms ...string) error {
	cm, err := GetGlobalCredentialsManager()
	if err != nil {
		return fmt.Errorf("failed to get credentials manager: %w", err)
	}

	// Verify Vault connection
	if err := cm.HealthCheck(); err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}

	// Preload specified platforms
	for _, platform := range preloadPlatforms {
		_, err := cm.GetPlatformCredentials(platform)
		if err != nil {
			log.Printf("Warning: failed to preload %s credentials: %v", platform, err)
		} else {
			log.Printf("Successfully preloaded %s credentials", platform)
		}
	}

	return nil
}
