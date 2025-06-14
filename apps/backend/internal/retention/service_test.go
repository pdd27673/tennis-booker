package retention

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultRetentionConfig(t *testing.T) {
	config := DefaultRetentionConfig()

	assert.Equal(t, 7*24*time.Hour, config.RetentionWindow)
	assert.Equal(t, 1000, config.BatchSize)
	assert.False(t, config.DryRun)
	assert.True(t, config.EnableMetrics)
	assert.Equal(t, "info", config.LogLevel)
}

func TestRetentionService_ValidateConfiguration(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)

	tests := []struct {
		name        string
		config      RetentionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid default config",
			config:      DefaultRetentionConfig(),
			expectError: false,
		},
		{
			name: "zero retention window",
			config: RetentionConfig{
				RetentionWindow: 0,
				BatchSize:       1000,
			},
			expectError: true,
			errorMsg:    "retention window must be positive",
		},
		{
			name: "negative retention window",
			config: RetentionConfig{
				RetentionWindow: -time.Hour,
				BatchSize:       1000,
			},
			expectError: true,
			errorMsg:    "retention window must be positive",
		},
		{
			name: "zero batch size",
			config: RetentionConfig{
				RetentionWindow: 7 * 24 * time.Hour,
				BatchSize:       0,
			},
			expectError: true,
			errorMsg:    "batch size must be positive",
		},
		{
			name: "negative batch size",
			config: RetentionConfig{
				RetentionWindow: 7 * 24 * time.Hour,
				BatchSize:       -100,
			},
			expectError: true,
			errorMsg:    "batch size must be positive",
		},
		{
			name: "batch size too large",
			config: RetentionConfig{
				RetentionWindow: 7 * 24 * time.Hour,
				BatchSize:       15000,
			},
			expectError: true,
			errorMsg:    "batch size too large",
		},
		{
			name: "valid custom config",
			config: RetentionConfig{
				RetentionWindow: 14 * 24 * time.Hour,
				BatchSize:       500,
				DryRun:          true,
				EnableMetrics:   false,
				LogLevel:        "debug",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &RetentionService{
				config: tt.config,
				logger: logger,
			}

			err := service.ValidateConfiguration()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRetentionService_GetConfiguration(t *testing.T) {
	config := RetentionConfig{
		RetentionWindow: 14 * 24 * time.Hour,
		BatchSize:       500,
		DryRun:          true,
		EnableMetrics:   false,
		LogLevel:        "debug",
	}

	service := &RetentionService{config: config}

	retrievedConfig := service.GetConfiguration()
	assert.Equal(t, config, retrievedConfig)
}

func TestRetentionService_UpdateConfiguration(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)

	// Start with default config
	service := &RetentionService{
		config: DefaultRetentionConfig(),
		logger: logger,
	}

	// Test valid update
	newConfig := RetentionConfig{
		RetentionWindow: 14 * 24 * time.Hour,
		BatchSize:       500,
		DryRun:          true,
		EnableMetrics:   false,
		LogLevel:        "debug",
	}

	err := service.UpdateConfiguration(newConfig)
	assert.NoError(t, err)
	assert.Equal(t, newConfig, service.GetConfiguration())

	// Test invalid update
	invalidConfig := RetentionConfig{
		RetentionWindow: -time.Hour, // Invalid
		BatchSize:       500,
	}

	err = service.UpdateConfiguration(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid configuration")

	// Config should remain unchanged after invalid update
	assert.Equal(t, newConfig, service.GetConfiguration())
}

func TestRetentionMetrics_Structure(t *testing.T) {
	// Test that RetentionMetrics struct has all expected fields
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)

	metrics := RetentionMetrics{
		StartTime:                  startTime,
		EndTime:                    endTime,
		Duration:                   endTime.Sub(startTime),
		CandidateSlotsFound:        100,
		SlotsCheckedAgainstPrefs:   100,
		SlotsIdentifiedForDeletion: 25,
		SlotsActuallyDeleted:       25,
		ActivePreferencesCount:     10,
		ErrorsEncountered:          0,
		DryRunMode:                 false,
	}

	// Verify all fields are accessible
	assert.Equal(t, startTime, metrics.StartTime)
	assert.Equal(t, endTime, metrics.EndTime)
	assert.Equal(t, 5*time.Minute, metrics.Duration)
	assert.Equal(t, 100, metrics.CandidateSlotsFound)
	assert.Equal(t, 100, metrics.SlotsCheckedAgainstPrefs)
	assert.Equal(t, 25, metrics.SlotsIdentifiedForDeletion)
	assert.Equal(t, 25, metrics.SlotsActuallyDeleted)
	assert.Equal(t, 10, metrics.ActivePreferencesCount)
	assert.Equal(t, 0, metrics.ErrorsEncountered)
	assert.False(t, metrics.DryRunMode)
}

func TestRetentionConfig_Structure(t *testing.T) {
	// Test that RetentionConfig struct has all expected fields
	config := RetentionConfig{
		RetentionWindow: 7 * 24 * time.Hour,
		BatchSize:       1000,
		DryRun:          true,
		EnableMetrics:   false,
		LogLevel:        "debug",
	}

	// Verify all fields are accessible
	assert.Equal(t, 7*24*time.Hour, config.RetentionWindow)
	assert.Equal(t, 1000, config.BatchSize)
	assert.True(t, config.DryRun)
	assert.False(t, config.EnableMetrics)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestRetentionService_LoggingMethods(t *testing.T) {
	// Create a logger that writes to a buffer for testing
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)

	service := &RetentionService{
		config: RetentionConfig{LogLevel: "debug"},
		logger: logger,
	}

	// Test that logging methods don't panic
	assert.NotPanics(t, func() {
		service.logInfo("test info message", map[string]interface{}{
			"key": "value",
		})
	})

	assert.NotPanics(t, func() {
		service.logError("test error message", assert.AnError, map[string]interface{}{
			"key": "value",
		})
	})

	assert.NotPanics(t, func() {
		service.logDebug("test debug message", map[string]interface{}{
			"key": "value",
		})
	})

	assert.NotPanics(t, func() {
		service.logWithLevel("CUSTOM", "test custom message", nil)
	})
}

func TestRetentionService_LoggingLevels(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)

	// Test with info level (debug should not log)
	service := &RetentionService{
		config: RetentionConfig{LogLevel: "info"},
		logger: logger,
	}

	// These should not panic regardless of log level
	assert.NotPanics(t, func() {
		service.logDebug("debug message", nil)
	})

	// Test with debug level (debug should log)
	service.config.LogLevel = "debug"
	assert.NotPanics(t, func() {
		service.logDebug("debug message", nil)
	})
}

// Integration tests would go here for MongoDB operations
// These would require a test MongoDB instance

/*
func TestRetentionService_RunRetentionCycle(t *testing.T) {
	// Setup test MongoDB connection
	// Create test data with various slot and preference configurations
	// Call RunRetentionCycle
	// Verify correct slots are identified for deletion
	// Verify metrics are collected correctly
}

func TestRetentionService_RunRetentionCycle_DryRun(t *testing.T) {
	// Setup test MongoDB connection
	// Create test data
	// Run with DryRun=true
	// Verify no actual deletions occur
	// Verify dry-run logging works correctly
}

func TestRetentionService_DeleteSlotsInBatches(t *testing.T) {
	// Setup test MongoDB connection
	// Create test slots
	// Test batch deletion with various batch sizes
	// Verify all slots are deleted correctly
}
*/
