package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tennis-booker/internal/secrets"
)

func TestNewConnectionManager(t *testing.T) {
	// Create a mock secrets manager
	sm := &secrets.SecretsManager{}

	cm := NewConnectionManager(sm)
	assert.NotNil(t, cm)
	assert.Equal(t, sm, cm.secretsManager)
}

func TestConnectionManager_GetConnectionURI(t *testing.T) {
	// Create a connection manager with mock config
	cm := &ConnectionManager{
		config: &DatabaseConfig{
			Host:     "localhost",
			Username: "testuser",
			Password: "testpass",
			Database: "testdb",
			Port:     "27017",
		},
	}

	uri, err := cm.GetConnectionURI()
	assert.NoError(t, err)
	assert.Equal(t, "mongodb://testuser:testpass@localhost:27017", uri)
}

func TestConnectionManager_GetConnectionURI_NoAuth(t *testing.T) {
	// Create a connection manager with no authentication
	cm := &ConnectionManager{
		config: &DatabaseConfig{
			Host:     "localhost",
			Username: "",
			Password: "",
			Database: "testdb",
			Port:     "27017",
		},
	}

	uri, err := cm.GetConnectionURI()
	assert.NoError(t, err)
	assert.Equal(t, "mongodb://localhost:27017", uri)
}

func TestConnectionManager_GetDatabaseName(t *testing.T) {
	// Test with configured database name
	cm := &ConnectionManager{
		config: &DatabaseConfig{
			Database: "custom_db",
		},
	}

	dbName, err := cm.GetDatabaseName()
	assert.NoError(t, err)
	assert.Equal(t, "custom_db", dbName)

	// Test with empty database name (should return default)
	cm.config.Database = ""
	dbName, err = cm.GetDatabaseName()
	assert.NoError(t, err)
	assert.Equal(t, "tennis_booking", dbName)
}

func TestConnectionManager_LoadConfig_WithMockSecrets(t *testing.T) {
	// Create a secrets manager with mock data
	sm := &secrets.SecretsManager{}

	// We can't easily test this without a real secrets manager
	// because the GetDBCredentials method would need to be mocked
	// This test demonstrates the structure
	cm := NewConnectionManager(sm)
	assert.NotNil(t, cm)
}

func TestDatabaseConfig(t *testing.T) {
	config := &DatabaseConfig{
		Host:     "test-host",
		Username: "test-user",
		Password: "test-pass",
		Database: "test-db",
		Port:     "27017",
	}

	assert.Equal(t, "test-host", config.Host)
	assert.Equal(t, "test-user", config.Username)
	assert.Equal(t, "test-pass", config.Password)
	assert.Equal(t, "test-db", config.Database)
	assert.Equal(t, "27017", config.Port)
}

// Integration test - requires Vault to be running
func TestConnectionManager_Integration(t *testing.T) {
	// Skip if no Vault environment
	t.Skip("Integration test requires running Vault instance with test data")

	cm, err := NewConnectionManagerFromEnv()
	require.NoError(t, err)
	defer cm.Close()

	// Test loading config
	err = cm.LoadConfig()
	if err != nil {
		t.Logf("Expected error loading config from Vault: %v", err)
		return
	}

	// Test getting URI
	uri, err := cm.GetConnectionURI()
	assert.NoError(t, err)
	assert.NotEmpty(t, uri)
	t.Logf("Generated URI: %s", uri)

	// Test getting database name
	dbName, err := cm.GetDatabaseName()
	assert.NoError(t, err)
	assert.NotEmpty(t, dbName)
	t.Logf("Database name: %s", dbName)
}
