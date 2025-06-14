package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	App struct {
		Name        string `mapstructure:"name"`
		Version     string `mapstructure:"version"`
		Environment string `mapstructure:"environment"`
	} `mapstructure:"app"`

	API struct {
		Port      int    `mapstructure:"port"`
		Timeout   string `mapstructure:"timeout"`
		RateLimit struct {
			Enabled           bool `mapstructure:"enabled"`
			RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
		} `mapstructure:"rateLimit"`
	} `mapstructure:"api"`

	Database struct {
		PoolSize      int    `mapstructure:"poolSize"`
		Timeout       string `mapstructure:"timeout"`
		RetryAttempts int    `mapstructure:"retryAttempts"`
	} `mapstructure:"database"`

	Scraper struct {
		Interval   int `mapstructure:"interval"`
		Timeout    int `mapstructure:"timeout"`
		MaxRetries int `mapstructure:"maxRetries"`
		DaysAhead  int `mapstructure:"daysAhead"`
		Platforms  struct {
			Clubspark struct {
				Enabled bool   `mapstructure:"enabled"`
				BaseURL string `mapstructure:"baseUrl"`
			} `mapstructure:"clubspark"`
			Courtsides struct {
				Enabled bool   `mapstructure:"enabled"`
				BaseURL string `mapstructure:"baseUrl"`
			} `mapstructure:"courtsides"`
		} `mapstructure:"platforms"`
	} `mapstructure:"scraper"`

	Notification struct {
		Port           int `mapstructure:"port"`
		EmailRateLimit int `mapstructure:"emailRateLimit"`
		BatchSize      int `mapstructure:"batchSize"`
		RetryAttempts  int `mapstructure:"retryAttempts"`
	} `mapstructure:"notification"`

	Logging struct {
		Level         string `mapstructure:"level"`
		Format        string `mapstructure:"format"`
		EnableConsole bool   `mapstructure:"enableConsole"`
		EnableFile    bool   `mapstructure:"enableFile"`
	} `mapstructure:"logging"`

	Features map[string]bool `mapstructure:"features"`
}

// Global configuration instance
var AppConfig *Config

// Load loads the configuration from files and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("default")
	viper.SetConfigType("json")

	// Add config paths (search in multiple locations)
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")
	viper.AddConfigPath("../../../config")    // For deeply nested services
	viper.AddConfigPath("../../../../config") // For tests

	// Read default config
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read default config: %w", err)
	}

	// Get environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	// Merge environment-specific config
	viper.SetConfigName(env)
	if err := viper.MergeInConfig(); err != nil {
		// Environment-specific config is optional
		fmt.Printf("No %s config found, using defaults\n", env)
	}

	// Enable environment variable overrides
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables to config keys
	bindEnvVars()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set global config
	AppConfig = &config

	return &config, nil
}

// bindEnvVars binds environment variables to configuration keys
func bindEnvVars() {
	// App settings
	viper.BindEnv("app.environment", "APP_ENV")

	// API settings
	viper.BindEnv("api.port", "API_PORT", "BACKEND_API_PORT")
	viper.BindEnv("api.timeout", "API_TIMEOUT", "BACKEND_API_TIMEOUT")
	viper.BindEnv("api.rateLimit.enabled", "API_RATE_LIMIT_ENABLED")
	viper.BindEnv("api.rateLimit.requestsPerMinute", "API_RATE_LIMIT_RPM")

	// Database settings
	viper.BindEnv("database.poolSize", "DB_POOL_SIZE", "BACKEND_DB_POOL_SIZE")
	viper.BindEnv("database.timeout", "DB_TIMEOUT", "BACKEND_DB_TIMEOUT")
	viper.BindEnv("database.retryAttempts", "DB_RETRY_ATTEMPTS")

	// Scraper settings
	viper.BindEnv("scraper.interval", "SCRAPER_INTERVAL")
	viper.BindEnv("scraper.timeout", "SCRAPER_TIMEOUT")
	viper.BindEnv("scraper.maxRetries", "SCRAPER_MAX_RETRIES")
	viper.BindEnv("scraper.daysAhead", "SCRAPER_DAYS_AHEAD")

	// Notification settings
	viper.BindEnv("notification.port", "NOTIFICATION_PORT")
	viper.BindEnv("notification.emailRateLimit", "NOTIFICATION_EMAIL_RATE_LIMIT")
	viper.BindEnv("notification.batchSize", "NOTIFICATION_BATCH_SIZE")
	viper.BindEnv("notification.retryAttempts", "NOTIFICATION_RETRY_ATTEMPTS")

	// Logging settings
	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")
	viper.BindEnv("logging.enableConsole", "LOG_ENABLE_CONSOLE")
	viper.BindEnv("logging.enableFile", "LOG_ENABLE_FILE")

	// Feature flags
	viper.BindEnv("features.advancedFiltering", "FEATURE_ADVANCED_FILTERING")
	viper.BindEnv("features.smsNotifications", "FEATURE_SMS_NOTIFICATIONS")
	viper.BindEnv("features.analytics", "FEATURE_ANALYTICS")
	viper.BindEnv("features.realTimeUpdates", "FEATURE_REAL_TIME_UPDATES")
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	// Validate required fields
	if config.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}

	// Validate port ranges
	if config.API.Port < 0 || config.API.Port > 65535 {
		return fmt.Errorf("api.port must be between 0 and 65535")
	}

	if config.Notification.Port < 0 || config.Notification.Port > 65535 {
		return fmt.Errorf("notification.port must be between 0 and 65535")
	}

	// Validate positive integers
	if config.Database.PoolSize <= 0 {
		return fmt.Errorf("database.poolSize must be positive")
	}

	if config.Scraper.Interval <= 0 {
		return fmt.Errorf("scraper.interval must be positive")
	}

	if config.Scraper.Timeout <= 0 {
		return fmt.Errorf("scraper.timeout must be positive")
	}

	// Validate timeout formats
	if _, err := time.ParseDuration(config.API.Timeout); err != nil {
		return fmt.Errorf("api.timeout must be a valid duration: %w", err)
	}

	if _, err := time.ParseDuration(config.Database.Timeout); err != nil {
		return fmt.Errorf("database.timeout must be a valid duration: %w", err)
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal"}
	validLevel := false
	for _, level := range validLogLevels {
		if config.Logging.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("logging.level must be one of: %v", validLogLevels)
	}

	return nil
}

// GetAPITimeout returns the API timeout as a time.Duration
func (c *Config) GetAPITimeout() time.Duration {
	duration, _ := time.ParseDuration(c.API.Timeout)
	return duration
}

// GetDatabaseTimeout returns the database timeout as a time.Duration
func (c *Config) GetDatabaseTimeout() time.Duration {
	duration, _ := time.ParseDuration(c.Database.Timeout)
	return duration
}

// IsFeatureEnabled checks if a feature flag is enabled
func (c *Config) IsFeatureEnabled(feature string) bool {
	enabled, exists := c.Features[feature]
	return exists && enabled
}

// GetScraperIntervalDuration returns the scraper interval as a time.Duration
func (c *Config) GetScraperIntervalDuration() time.Duration {
	return time.Duration(c.Scraper.Interval) * time.Second
}

// GetScraperTimeoutDuration returns the scraper timeout as a time.Duration
func (c *Config) GetScraperTimeoutDuration() time.Duration {
	return time.Duration(c.Scraper.Timeout) * time.Second
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsTest returns true if running in test environment
func (c *Config) IsTest() bool {
	return c.App.Environment == "test"
}
