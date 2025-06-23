package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"tennis-booker/internal/secrets"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Username string
	Password string
	Database string
	Port     string
}

// ConnectionManager manages database connections with Vault integration
type ConnectionManager struct {
	secretsManager *secrets.SecretsManager
	config         *DatabaseConfig
}

// NewConnectionManager creates a new database connection manager
func NewConnectionManager(secretsManager *secrets.SecretsManager) *ConnectionManager {
	return &ConnectionManager{
		secretsManager: secretsManager,
	}
}

// NewConnectionManagerFromEnv creates a connection manager using environment variables for Vault
func NewConnectionManagerFromEnv() (*ConnectionManager, error) {
	secretsManager, err := secrets.NewSecretsManagerFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets manager: %w", err)
	}

	return NewConnectionManager(secretsManager), nil
}

// LoadConfig loads database configuration from Vault
func (cm *ConnectionManager) LoadConfig() error {
	username, password, host, database, err := cm.secretsManager.GetDBCredentials()
	if err != nil {
		return fmt.Errorf("failed to load database credentials from vault: %w", err)
	}

	// Set default port if not specified in host
	port := "27017"
	if host == "" {
		host = "localhost"
	}

	cm.config = &DatabaseConfig{
		Host:     host,
		Username: username,
		Password: password,
		Database: database,
		Port:     port,
	}

	return nil
}

// GetConnectionURI builds the MongoDB connection URI from the loaded configuration
func (cm *ConnectionManager) GetConnectionURI() (string, error) {
	if cm.config == nil {
		if err := cm.LoadConfig(); err != nil {
			return "", err
		}
	}

	// Build MongoDB URI
	// Format: mongodb://username:password@host:port/database
	if cm.config.Username != "" && cm.config.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s",
			cm.config.Username,
			cm.config.Password,
			cm.config.Host,
			cm.config.Port,
		), nil
	}

	// No authentication
	return fmt.Sprintf("mongodb://%s:%s", cm.config.Host, cm.config.Port), nil
}

// GetDatabaseName returns the database name from configuration
func (cm *ConnectionManager) GetDatabaseName() (string, error) {
	if cm.config == nil {
		if err := cm.LoadConfig(); err != nil {
			return "", err
		}
	}

	if cm.config.Database == "" {
		return "tennis_booking", nil // Default database name
	}

	return cm.config.Database, nil
}

// Connect establishes a connection to MongoDB using Vault credentials
func (cm *ConnectionManager) Connect() (*mongo.Database, error) {
	// Load configuration from Vault
	if err := cm.LoadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load database configuration: %w", err)
	}

	// Get connection URI and database name
	uri, err := cm.GetConnectionURI()
	if err != nil {
		return nil, fmt.Errorf("failed to build connection URI: %w", err)
	}

	dbName, err := cm.GetDatabaseName()
	if err != nil {
		return nil, fmt.Errorf("failed to get database name: %w", err)
	}

	log.Printf("Connecting to MongoDB at %s (database: %s)", cm.config.Host, dbName)

	// Use the existing InitDatabase function
	db, err := InitDatabase(uri, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✅ Successfully connected to MongoDB using Vault credentials")
	return db, nil
}

// ConnectWithFallback attempts to connect using Vault credentials, with fallback to environment variables
func (cm *ConnectionManager) ConnectWithFallback() (*mongo.Database, error) {
	// Try Vault first
	db, err := cm.Connect()
	if err != nil {
		log.Printf("⚠️ Failed to connect using Vault credentials: %v", err)
		log.Println("🔄 Attempting fallback to environment variables...")

		// Fallback to environment variables
		return cm.connectFromEnv()
	}

	return db, nil
}

// connectFromEnv connects using environment variables as fallback
func (cm *ConnectionManager) connectFromEnv() (*mongo.Database, error) {
	// Read from environment variables
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		// Build from individual components
		username := os.Getenv("MONGO_ROOT_USERNAME")
		password := os.Getenv("MONGO_ROOT_PASSWORD")
		host := os.Getenv("MONGO_HOST")
		port := os.Getenv("MONGO_PORT")

		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "27017"
		}

		if username != "" && password != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s:%s?authSource=admin", username, password, host, port)
		} else {
			uri = fmt.Sprintf("mongodb://%s:%s", host, port)
		}
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "tennis_booking"
	}

	log.Printf("⚠️ Using fallback connection: %s", uri)

	db, err := InitDatabase(uri, dbName)
	if err != nil {
		return nil, fmt.Errorf("fallback connection failed: %w", err)
	}

	log.Println("✅ Connected using fallback credentials")
	return db, nil
}

// HealthCheck verifies the database connection
func (cm *ConnectionManager) HealthCheck(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.Client().Ping(ctx, readpref.Primary())
}

// Close closes the secrets manager connection
func (cm *ConnectionManager) Close() error {
	if cm.secretsManager != nil {
		return cm.secretsManager.Close()
	}
	return nil
}

// GetSecretsManager returns the underlying secrets manager
func (cm *ConnectionManager) GetSecretsManager() *secrets.SecretsManager {
	return cm.secretsManager
}
