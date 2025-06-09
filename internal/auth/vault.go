package auth

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

// VaultClient wraps the HashiCorp Vault API client
type VaultClient struct {
	client *api.Client
}

// VaultConfig holds configuration for Vault connection
type VaultConfig struct {
	Address string
	Token   string
}

// NewVaultClient creates a new Vault client with the provided configuration
func NewVaultClient(config *VaultConfig) (*VaultClient, error) {
	if config == nil {
		config = &VaultConfig{
			Address: getEnvOrDefault("VAULT_ADDR", "http://localhost:8200"),
			Token:   getEnvOrDefault("VAULT_TOKEN", ""),
		}
	}

	if config.Token == "" {
		return nil, fmt.Errorf("vault token is required")
	}

	// Create Vault client configuration
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = config.Address

	// Create the client
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Set the token
	client.SetToken(config.Token)

	return &VaultClient{client: client}, nil
}

// NewVaultClientFromEnv creates a new Vault client using environment variables
func NewVaultClientFromEnv() (*VaultClient, error) {
	return NewVaultClient(nil)
}

// GetSecret retrieves a secret from the specified path
func (vc *VaultClient) GetSecret(path string) (map[string]interface{}, error) {
	if vc.client == nil {
		return nil, fmt.Errorf("vault client not initialized")
	}

	// Read the secret from Vault
	secret, err := vc.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from path %s: %w", path, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	// For KV v2, the actual data is nested under "data"
	if data, ok := secret.Data["data"]; ok {
		if secretData, ok := data.(map[string]interface{}); ok {
			return secretData, nil
		}
	}

	// Fallback to raw data (KV v1 or other engines)
	return secret.Data, nil
}

// GetSecretField retrieves a specific field from a secret
func (vc *VaultClient) GetSecretField(path, field string) (string, error) {
	secretData, err := vc.GetSecret(path)
	if err != nil {
		return "", err
	}

	value, exists := secretData[field]
	if !exists {
		return "", fmt.Errorf("field %s not found in secret at path %s", field, path)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("field %s at path %s is not a string", field, path)
	}

	return str, nil
}

// HealthCheck verifies the Vault connection
func (vc *VaultClient) HealthCheck() error {
	if vc.client == nil {
		return fmt.Errorf("vault client not initialized")
	}

	resp, err := vc.client.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}

	if resp.Sealed {
		return fmt.Errorf("vault is sealed")
	}

	return nil
}

// Close performs any necessary cleanup (currently no-op but included for future use)
func (vc *VaultClient) Close() error {
	// Nothing to close for the current Vault client
	return nil
}

// getEnvOrDefault returns the environment variable value or a default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 