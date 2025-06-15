package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	MongoDB  MongoDBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Email    EmailConfig
	CORS     CORSConfig
	Scraper  ScraperConfig
	Vault    VaultConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	Environment  string
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string
	Database string
	Username string
	Password string
	Host     string
	Port     string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Issuer           string
	AccessTokenTTL   int // in hours
	RefreshTokenTTL  int // in hours
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// ScraperConfig holds scraper configuration
type ScraperConfig struct {
	Enabled  bool
	Interval int // in minutes
}

// VaultConfig holds Vault configuration
type VaultConfig struct {
	Address string
	Token   string
}

// Global configuration instance
var AppConfig *Config

// Load loads configuration from environment variables
func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 30),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 30),
			IdleTimeout:  getEnvAsInt("IDLE_TIMEOUT", 120),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin"),
			Database: getEnv("DB_NAME", "tennis_booking"),
			Username: getEnv("MONGO_ROOT_USERNAME", "admin"),
			Password: getEnv("MONGO_ROOT_PASSWORD", "password"),
			Host:     getEnv("MONGO_HOST", "localhost"),
			Port:     getEnv("MONGO_PORT", "27017"),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", "password"),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Issuer:          getEnv("JWT_ISSUER", "tennis-booker"),
			AccessTokenTTL:  getEnvAsInt("JWT_ACCESS_TTL", 24),   // 24 hours
			RefreshTokenTTL: getEnvAsInt("JWT_REFRESH_TTL", 168), // 7 days
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("GMAIL_EMAIL", "demo@example.com"),
			SMTPPassword: getEnv("GMAIL_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "demo@example.com"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{
				"http://localhost:3000",
				"http://localhost:5173",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:5173",
			}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{
				"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
			}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{
				"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin",
			}),
		},
		Scraper: ScraperConfig{
			Enabled:  getEnvAsBool("SCRAPER_ENABLED", true),
			Interval: getEnvAsInt("SCRAPER_INTERVAL", 30), // 30 minutes
		},
		Vault: VaultConfig{
			Address: getEnv("VAULT_ADDR", "http://localhost:8200"),
			Token:   getEnv("VAULT_TOKEN", "dev-token"),
		},
	}, nil
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// IsLocal returns true if running in local development environment
func (c *Config) IsLocal() bool {
	return c.Server.Environment == "development" || c.Server.Environment == "local"
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
}
	return defaultValue
}

// Validate configuration
func validateConfig(config *Config) error {
	// Validate required fields
	if config.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}

	if config.Server.Host == "" {
		return fmt.Errorf("server.host is required")
	}

	if config.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server.readTimeout must be positive")
	}

	if config.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server.writeTimeout must be positive")
	}

	if config.Server.IdleTimeout <= 0 {
		return fmt.Errorf("server.idleTimeout must be positive")
	}

	if config.Server.Environment == "" {
		return fmt.Errorf("server.environment is required")
	}

	if config.MongoDB.URI == "" {
		return fmt.Errorf("mongoDB.uri is required")
	}

	if config.MongoDB.Database == "" {
		return fmt.Errorf("mongoDB.database is required")
	}

	if config.MongoDB.Username == "" {
		return fmt.Errorf("mongoDB.username is required")
	}

	if config.MongoDB.Password == "" {
		return fmt.Errorf("mongoDB.password is required")
	}

	if config.MongoDB.Host == "" {
		return fmt.Errorf("mongoDB.host is required")
	}

	if config.MongoDB.Port == "" {
		return fmt.Errorf("mongoDB.port is required")
	}

	if config.Redis.Address == "" {
		return fmt.Errorf("redis.address is required")
	}

	if config.Redis.DB < 0 {
		return fmt.Errorf("redis.db must be non-negative")
	}

	if config.JWT.Issuer == "" {
		return fmt.Errorf("jwt.issuer is required")
	}

	if config.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("jwt.accessTokenTTL must be positive")
	}

	if config.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("jwt.refreshTokenTTL must be positive")
	}

	if config.Email.SMTPHost == "" {
		return fmt.Errorf("email.smtpHost is required")
	}

	if config.Email.SMTPPort == "" {
		return fmt.Errorf("email.smtpPort is required")
	}

	if config.Email.SMTPUsername == "" {
		return fmt.Errorf("email.smtpUsername is required")
	}

	if config.Email.SMTPPassword == "" {
		return fmt.Errorf("email.smtpPassword is required")
	}

	if config.Email.FromEmail == "" {
		return fmt.Errorf("email.fromEmail is required")
	}

	if len(config.CORS.AllowedOrigins) == 0 {
		return fmt.Errorf("cors.allowedOrigins is required")
	}

	if len(config.CORS.AllowedMethods) == 0 {
		return fmt.Errorf("cors.allowedMethods is required")
	}

	if len(config.CORS.AllowedHeaders) == 0 {
		return fmt.Errorf("cors.allowedHeaders is required")
	}

	if !config.Scraper.Enabled {
		return fmt.Errorf("scraper.enabled must be true")
	}

	if config.Scraper.Interval <= 0 {
		return fmt.Errorf("scraper.interval must be positive")
	}

	if config.Vault.Address == "" {
		return fmt.Errorf("vault.address is required")
	}

	if config.Vault.Token == "" {
		return fmt.Errorf("vault.token is required")
	}

	return nil
}

// Set global config
func (c *Config) SetGlobalConfig() {
	AppConfig = c
}

// GetAPITimeout returns the API timeout as a time.Duration
func (c *Config) GetAPITimeout() time.Duration {
	return time.Duration(c.Server.ReadTimeout) * time.Second
}

// GetDatabaseTimeout returns the database timeout as a time.Duration
func (c *Config) GetDatabaseTimeout() time.Duration {
	return time.Duration(c.Server.IdleTimeout) * time.Second
}

// IsFeatureEnabled checks if a feature flag is enabled
func (c *Config) IsFeatureEnabled(feature string) bool {
	// Implementation of IsFeatureEnabled method
	return false // Placeholder, actual implementation needed
}

// GetScraperIntervalDuration returns the scraper interval as a time.Duration
func (c *Config) GetScraperIntervalDuration() time.Duration {
	return time.Duration(c.Scraper.Interval) * time.Minute
}

// GetScraperTimeoutDuration returns the scraper timeout as a time.Duration
func (c *Config) GetScraperTimeoutDuration() time.Duration {
	return 0 // Placeholder, actual implementation needed
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsTest returns true if running in test environment
func (c *Config) IsTest() bool {
	return c.Server.Environment == "test"
}
