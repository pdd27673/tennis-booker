package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("APP_ENV")
	defer os.Setenv("APP_ENV", originalEnv)

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
				assert.Equal(t, "Tennis Booker", config.App.Name)
				assert.Equal(t, "development", config.App.Environment)
				assert.Equal(t, 8080, config.API.Port)
				assert.Equal(t, "debug", config.Logging.Level) // development override
				assert.True(t, config.Features["realtimeupdates"]) // development override
			},
		},
		{
			name: "production configuration",
			env:  "production",
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "production", config.App.Environment)
				assert.Equal(t, "info", config.Logging.Level)
				assert.Equal(t, 200, config.API.RateLimit.RequestsPerMinute) // production override
				assert.False(t, config.Logging.EnableConsole) // production override
			},
		},
		{
			name: "test configuration",
			env:  "test",
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "test", config.App.Environment)
				assert.Equal(t, 0, config.API.Port) // test override
				assert.Equal(t, "error", config.Logging.Level) // test override
				assert.Equal(t, 1, config.Scraper.Interval) // test override
			},
		},
		{
			name: "environment variable overrides",
			env:  "development",
			envVars: map[string]string{
				"API_PORT":                    "9000",
				"LOG_LEVEL":                   "warn",
				"SCRAPER_INTERVAL":            "600",
				"FEATURE_ADVANCED_FILTERING":  "false",
				"NOTIFICATION_EMAIL_RATE_LIMIT": "25",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 9000, config.API.Port)
				assert.Equal(t, "warn", config.Logging.Level)
				assert.Equal(t, 600, config.Scraper.Interval)
				assert.False(t, config.Features["advancedfiltering"])
				assert.Equal(t, 25, config.Notification.EmailRateLimit)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			os.Setenv("APP_ENV", tt.env)

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
		errorMsg    string
	}{
		{
			name: "valid configuration",
			config: &Config{
				App: struct {
					Name        string `mapstructure:"name"`
					Version     string `mapstructure:"version"`
					Environment string `mapstructure:"environment"`
				}{
					Name:        "Test App",
					Version:     "1.0.0",
					Environment: "test",
				},
				API: struct {
					Port      int    `mapstructure:"port"`
					Timeout   string `mapstructure:"timeout"`
					RateLimit struct {
						Enabled           bool `mapstructure:"enabled"`
						RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
					} `mapstructure:"rateLimit"`
				}{
					Port:    8080,
					Timeout: "30s",
				},
				Database: struct {
					PoolSize      int    `mapstructure:"poolSize"`
					Timeout       string `mapstructure:"timeout"`
					RetryAttempts int    `mapstructure:"retryAttempts"`
				}{
					PoolSize: 10,
					Timeout:  "30s",
				},
				Scraper: struct {
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
				}{
					Interval: 300,
					Timeout:  30,
				},
				Notification: struct {
					Port           int `mapstructure:"port"`
					EmailRateLimit int `mapstructure:"emailRateLimit"`
					BatchSize      int `mapstructure:"batchSize"`
					RetryAttempts  int `mapstructure:"retryAttempts"`
				}{
					Port: 8081,
				},
				Logging: struct {
					Level         string `mapstructure:"level"`
					Format        string `mapstructure:"format"`
					EnableConsole bool   `mapstructure:"enableConsole"`
					EnableFile    bool   `mapstructure:"enableFile"`
				}{
					Level: "info",
				},
			},
			expectError: false,
		},
		{
			name: "missing app name",
			config: &Config{
				App: struct {
					Name        string `mapstructure:"name"`
					Version     string `mapstructure:"version"`
					Environment string `mapstructure:"environment"`
				}{
					Name: "", // Missing name
				},
			},
			expectError: true,
			errorMsg:    "app.name is required",
		},
		{
			name: "invalid port",
			config: &Config{
				App: struct {
					Name        string `mapstructure:"name"`
					Version     string `mapstructure:"version"`
					Environment string `mapstructure:"environment"`
				}{
					Name: "Test App",
				},
				API: struct {
					Port      int    `mapstructure:"port"`
					Timeout   string `mapstructure:"timeout"`
					RateLimit struct {
						Enabled           bool `mapstructure:"enabled"`
						RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
					} `mapstructure:"rateLimit"`
				}{
					Port:    70000, // Invalid port
					Timeout: "30s",
				},
			},
			expectError: true,
			errorMsg:    "api.port must be between 0 and 65535",
		},
		{
			name: "invalid timeout",
			config: &Config{
				App: struct {
					Name        string `mapstructure:"name"`
					Version     string `mapstructure:"version"`
					Environment string `mapstructure:"environment"`
				}{
					Name: "Test App",
				},
				API: struct {
					Port      int    `mapstructure:"port"`
					Timeout   string `mapstructure:"timeout"`
					RateLimit struct {
						Enabled           bool `mapstructure:"enabled"`
						RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
					} `mapstructure:"rateLimit"`
				}{
					Port:    8080,
					Timeout: "invalid", // Invalid timeout
				},
				Database: struct {
					PoolSize      int    `mapstructure:"poolSize"`
					Timeout       string `mapstructure:"timeout"`
					RetryAttempts int    `mapstructure:"retryAttempts"`
				}{
					PoolSize: 10,
					Timeout:  "30s",
				},
				Scraper: struct {
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
				}{
					Interval: 300,
					Timeout:  30,
				},
				Logging: struct {
					Level         string `mapstructure:"level"`
					Format        string `mapstructure:"format"`
					EnableConsole bool   `mapstructure:"enableConsole"`
					EnableFile    bool   `mapstructure:"enableFile"`
				}{
					Level: "info",
				},
			},
			expectError: true,
			errorMsg:    "api.timeout must be a valid duration",
		},
		{
			name: "invalid log level",
			config: &Config{
				App: struct {
					Name        string `mapstructure:"name"`
					Version     string `mapstructure:"version"`
					Environment string `mapstructure:"environment"`
				}{
					Name: "Test App",
				},
				API: struct {
					Port      int    `mapstructure:"port"`
					Timeout   string `mapstructure:"timeout"`
					RateLimit struct {
						Enabled           bool `mapstructure:"enabled"`
						RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
					} `mapstructure:"rateLimit"`
				}{
					Port:    8080,
					Timeout: "30s",
				},
				Database: struct {
					PoolSize      int    `mapstructure:"poolSize"`
					Timeout       string `mapstructure:"timeout"`
					RetryAttempts int    `mapstructure:"retryAttempts"`
				}{
					PoolSize: 10,
					Timeout:  "30s",
				},
				Scraper: struct {
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
				}{
					Interval: 300,
					Timeout:  30,
				},
				Logging: struct {
					Level         string `mapstructure:"level"`
					Format        string `mapstructure:"format"`
					EnableConsole bool   `mapstructure:"enableConsole"`
					EnableFile    bool   `mapstructure:"enableFile"`
				}{
					Level: "invalid", // Invalid log level
				},
			},
			expectError: true,
			errorMsg:    "logging.level must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigHelperMethods(t *testing.T) {
	config := &Config{
		App: struct {
			Name        string `mapstructure:"name"`
			Version     string `mapstructure:"version"`
			Environment string `mapstructure:"environment"`
		}{
			Environment: "development",
		},
		API: struct {
			Port      int    `mapstructure:"port"`
			Timeout   string `mapstructure:"timeout"`
			RateLimit struct {
				Enabled           bool `mapstructure:"enabled"`
				RequestsPerMinute int  `mapstructure:"requestsPerMinute"`
			} `mapstructure:"rateLimit"`
		}{
			Timeout: "30s",
		},
		Database: struct {
			PoolSize      int    `mapstructure:"poolSize"`
			Timeout       string `mapstructure:"timeout"`
			RetryAttempts int    `mapstructure:"retryAttempts"`
		}{
			Timeout: "60s",
		},
		Scraper: struct {
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
		}{
			Interval: 300,
			Timeout:  30,
		},
		Features: map[string]bool{
			"advancedfiltering": true,
			"smsnotifications":  false,
		},
	}

	// Test timeout methods
	assert.Equal(t, 30*time.Second, config.GetAPITimeout())
	assert.Equal(t, 60*time.Second, config.GetDatabaseTimeout())
	assert.Equal(t, 300*time.Second, config.GetScraperIntervalDuration())
	assert.Equal(t, 30*time.Second, config.GetScraperTimeoutDuration())

	// Test environment methods
	assert.True(t, config.IsDevelopment())
	assert.False(t, config.IsProduction())
	assert.False(t, config.IsTest())

	// Test feature flag methods
	assert.True(t, config.IsFeatureEnabled("advancedfiltering"))
	assert.False(t, config.IsFeatureEnabled("smsnotifications"))
	assert.False(t, config.IsFeatureEnabled("nonExistentFeature"))
} 