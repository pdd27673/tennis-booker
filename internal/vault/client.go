package vault

import (
	"fmt"
	"log"

	vault "github.com/hashicorp/vault/api"
)

// VaultClient represents the HashiCorp Vault client
type VaultClient struct {
	Client *vault.Client
}

// NewVaultClient creates a new HashiCorp Vault client
func NewVaultClient(address, token string) (*VaultClient, error) {
	// Create a Vault configuration
	config := vault.DefaultConfig()
	config.Address = address

	// Create a new client
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	// Set the token
	client.SetToken(token)

	// Verify the connection by reading the sys/health endpoint
	health, err := client.Sys().Health()
	if err != nil {
		return nil, fmt.Errorf("failed to check Vault health: %w", err)
	}

	if !health.Initialized {
		return nil, fmt.Errorf("vault is not initialized")
	}

	if health.Sealed {
		return nil, fmt.Errorf("vault is sealed")
	}

	log.Println("Successfully connected to Vault")

	return &VaultClient{
		Client: client,
	}, nil
}

// ReadSecret reads a secret from the specified path
func (v *VaultClient) ReadSecret(path string) (map[string]interface{}, error) {
	secret, err := v.Client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from %s: %w", path, err)
	}

	if secret == nil {
		return nil, fmt.Errorf("no secret found at %s", path)
	}

	return secret.Data, nil
}

// WriteSecret writes a secret to the specified path
func (v *VaultClient) WriteSecret(path string, data map[string]interface{}) error {
	_, err := v.Client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("failed to write secret to %s: %w", path, err)
	}

	return nil
}

// DeleteSecret deletes a secret from the specified path
func (v *VaultClient) DeleteSecret(path string) error {
	_, err := v.Client.Logical().Delete(path)
	if err != nil {
		return fmt.Errorf("failed to delete secret at %s: %w", path, err)
	}

	return nil
} 