package auth

import (
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

// VaultClient wraps the Vault API client with additional functionality
type VaultClient struct {
	client *api.Client
}

// NewVaultClient creates a new Vault client with the provided configuration
func NewVaultClient(config *api.Config) (*VaultClient, error) {
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &VaultClient{
		client: client,
	}, nil
}

// NewVaultClientFromEnv creates a new Vault client using environment variables
func NewVaultClientFromEnv() (*VaultClient, error) {
	// Read VAULT_ADDR from environment
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		return nil, fmt.Errorf("VAULT_ADDR environment variable is required")
	}

	// Create default config
	config := api.DefaultConfig()
	config.Address = vaultAddr

	// Create client
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	// Handle authentication
	vaultClient := &VaultClient{client: client}
	if err := vaultClient.authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate with vault: %w", err)
	}

	return vaultClient, nil
}

// authenticate handles Vault authentication using either AppRole or Token
func (vc *VaultClient) authenticate() error {
	// Try AppRole authentication first
	roleID := os.Getenv("VAULT_ROLE_ID")
	secretID := os.Getenv("VAULT_SECRET_ID")

	if roleID != "" && secretID != "" {
		return vc.authenticateWithAppRole(roleID, secretID)
	}

	// Fallback to token authentication
	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return fmt.Errorf("vault token is required: set VAULT_TOKEN or VAULT_ROLE_ID/VAULT_SECRET_ID")
	}

	vc.client.SetToken(token)
	return nil
}

// authenticateWithAppRole performs AppRole authentication
func (vc *VaultClient) authenticateWithAppRole(roleID, secretID string) error {
	// Prepare AppRole login data
	data := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	// Perform AppRole login
	resp, err := vc.client.Logical().Write("auth/approle/login", data)
	if err != nil {
		return fmt.Errorf("failed to login with approle: %w", err)
	}

	if resp == nil || resp.Auth == nil {
		return fmt.Errorf("no auth info returned from approle login")
	}

	// Set the client token
	vc.client.SetToken(resp.Auth.ClientToken)
	return nil
}

// GetSecret retrieves a secret from Vault at the specified path
func (vc *VaultClient) GetSecret(path string) (map[string]interface{}, error) {
	secret, err := vc.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret at path %s: %w", path, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("no secret found at path %s", path)
	}

	// For KV v2, the actual data is nested under "data"
	if data, ok := secret.Data["data"].(map[string]interface{}); ok {
		return data, nil
	}

	// For KV v1 or other secret engines, return data directly
	return secret.Data, nil
}

// HealthCheck verifies the Vault connection and authentication
func (vc *VaultClient) HealthCheck() error {
	// Try to read the token's own information
	_, err := vc.client.Auth().Token().LookupSelf()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	return nil
}

// Close closes the Vault client connection
func (vc *VaultClient) Close() error {
	// The Vault client doesn't require explicit closing
	// This method is provided for interface compatibility
	return nil
}

// GetClient returns the underlying Vault API client
func (vc *VaultClient) GetClient() *api.Client {
	return vc.client
}
