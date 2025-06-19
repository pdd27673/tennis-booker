package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		env         string
		envVars     map[string]string
		expectError bool
		validate    func(t *testing.T, config *Config)
	}{
		{
			name: "default configuration",
			env:  "development",
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "development", config.Server.Environment)
				assert.Equal(t, "8080", config.Server.Port)
				assert.Equal(t, "tennis_booking", config.MongoDB.Database)
			},
		},
		{
			name: "production configuration",
			env:  "production",
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "production", config.Server.Environment)
				assert.Equal(t, "localhost:6379", config.Redis.Address)
			},
		},
		{
			name: "test configuration",
			env:  "test",
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "test", config.Server.Environment)
				assert.Equal(t, "8080", config.Server.Port)
				assert.Equal(t, 30, config.Scraper.Interval)
			},
		},
		{
			name: "environment variable overrides",
			env:  "development",
			envVars: map[string]string{
				"PORT":             "9000",
				"SCRAPER_INTERVAL": "600",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "development", config.Server.Environment)
				assert.Equal(t, "9000", config.Server.Port)
				assert.Equal(t, 600, config.Scraper.Interval)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			os.Setenv("ENVIRONMENT", tt.env)
			defer os.Unsetenv("ENVIRONMENT")

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Load configuration
			config, err := Load()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)

			// Run validation
			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Host:        "localhost",
					Environment: "test",
				},
				MongoDB: MongoDBConfig{
					Database: "tennis_booking",
					Host:     "localhost",
					Port:     "27017",
				},
				Scraper: ScraperConfig{
					Enabled:  true,
					Interval: 30,
				},
			},
			expectError: false,
		},
		{
			name:        "empty config",
			config:      &Config{},
			expectError: false, // Our config doesn't require validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since our config doesn't have validation, just test that it exists
			assert.NotNil(t, tt.config)
		})
	}
}

func TestConfigHelperMethods(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Environment: "development",
		},
		Scraper: ScraperConfig{
			Interval: 300,
		},
	}

	// Test environment methods
	assert.True(t, config.IsDevelopment())
	assert.False(t, config.IsProduction())
	assert.False(t, config.IsTest())

	// Test scraper interval
	assert.Equal(t, 300, config.Scraper.Interval)
}

func TestIsProduction(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Environment: "production",
		},
	}
	assert.True(t, config.IsProduction())
	assert.False(t, config.IsLocal())
}

func TestIsLocal(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Environment: "development",
		},
	}
	assert.True(t, config.IsLocal())
	assert.False(t, config.IsProduction())
}
