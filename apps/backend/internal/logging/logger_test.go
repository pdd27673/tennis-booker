package logging

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewLogger(t *testing.T) {
	logger := New("test-service")
	if logger.serviceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got %s", logger.serviceName)
	}
	if logger.minLevel != INFO {
		t.Errorf("Expected default log level INFO, got %v", logger.minLevel)
	}
}

func TestLogLevelFromEnv(t *testing.T) {
	// Test DEBUG level
	os.Setenv("LOG_LEVEL", "DEBUG")
	logger := New("test-service")
	if logger.minLevel != DEBUG {
		t.Errorf("Expected DEBUG level, got %v", logger.minLevel)
	}

	// Test ERROR level
	os.Setenv("LOG_LEVEL", "ERROR")
	logger = New("test-service")
	if logger.minLevel != ERROR {
		t.Errorf("Expected ERROR level, got %v", logger.minLevel)
	}

	// Clean up
	os.Unsetenv("LOG_LEVEL")
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

func TestShouldLog(t *testing.T) {
	logger := New("test-service")
	logger.minLevel = WARN

	if logger.shouldLog(DEBUG) {
		t.Error("Should not log DEBUG when min level is WARN")
	}
	if logger.shouldLog(INFO) {
		t.Error("Should not log INFO when min level is WARN")
	}
	if !logger.shouldLog(WARN) {
		t.Error("Should log WARN when min level is WARN")
	}
	if !logger.shouldLog(ERROR) {
		t.Error("Should log ERROR when min level is WARN")
	}
}

func TestJSONLogging(t *testing.T) {
	// Set JSON format
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_LEVEL", "DEBUG")

	// Capture output
	var buf bytes.Buffer
	logger := New("test-service")
	logger.logger.SetOutput(&buf)

	logger.Info("test message", map[string]interface{}{
		"key": "value",
		"num": 42,
	})

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("JSON output should contain level")
	}
	if !strings.Contains(output, `"service":"test-service"`) {
		t.Error("JSON output should contain service name")
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Error("JSON output should contain message")
	}

	// Verify it's valid JSON
	var logEntry LogEntry
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}

	// Clean up
	os.Unsetenv("LOG_FORMAT")
	os.Unsetenv("LOG_LEVEL")
}

func TestHumanReadableLogging(t *testing.T) {
	// Ensure human-readable format (default)
	os.Unsetenv("LOG_FORMAT")
	os.Setenv("LOG_LEVEL", "DEBUG")

	// Capture output
	var buf bytes.Buffer
	logger := New("test-service")
	logger.logger.SetOutput(&buf)

	logger.Info("test message", map[string]interface{}{
		"key": "value",
	})

	output := buf.String()
	if !strings.Contains(output, "[INFO]") {
		t.Error("Human-readable output should contain [INFO]")
	}
	if !strings.Contains(output, "test-service") {
		t.Error("Human-readable output should contain service name")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Human-readable output should contain message")
	}

	// Clean up
	os.Unsetenv("LOG_LEVEL")
}

func TestWithFields(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	logger := New("test-service")
	logger.logger.SetOutput(&buf)

	fieldLogger := logger.WithFields(map[string]interface{}{
		"user_id": "123",
		"action":  "login",
	})

	fieldLogger.Info("User logged in")

	output := buf.String()
	if !strings.Contains(output, "User logged in") {
		t.Error("Output should contain the message")
	}
	// Note: For human-readable format, fields are included in the output
	if !strings.Contains(output, "user_id") {
		t.Error("Output should contain field key")
	}
}

func TestConvenienceMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := New("test-service")
	logger.logger.SetOutput(&buf)

	// Test StartupInfo
	logger.StartupInfo("Server starting", "8080", "production")
	output := buf.String()
	if !strings.Contains(output, "Server starting") {
		t.Error("StartupInfo should log the message")
	}

	// Reset buffer
	buf.Reset()

	// Test ConnectionInfo
	logger.ConnectionInfo("Database connected", "mongodb", "localhost:27017")
	output = buf.String()
	if !strings.Contains(output, "Database connected") {
		t.Error("ConnectionInfo should log the message")
	}

	// Reset buffer
	buf.Reset()

	// Test ErrorWithCode
	logger.ErrorWithCode("Database error", "DB001", nil)
	output = buf.String()
	if !strings.Contains(output, "Database error") {
		t.Error("ErrorWithCode should log the message")
	}
}

func TestLogEntryTimestamp(t *testing.T) {
	os.Setenv("LOG_FORMAT", "json")

	var buf bytes.Buffer
	logger := New("test-service")
	logger.logger.SetOutput(&buf)

	before := time.Now()
	logger.Info("test message")
	after := time.Now()

	var logEntry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if logEntry.Timestamp.Before(before) || logEntry.Timestamp.After(after) {
		t.Error("Log entry timestamp should be within expected range")
	}

	// Clean up
	os.Unsetenv("LOG_FORMAT")
}

func TestLogLevels(t *testing.T) {
	os.Setenv("LOG_LEVEL", "DEBUG")

	var buf bytes.Buffer
	logger := New("test-service")
	logger.logger.SetOutput(&buf)

	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	expectedMessages := []string{
		"debug message",
		"info message",
		"warn message",
		"error message",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Output should contain '%s'", msg)
		}
	}

	// Clean up
	os.Unsetenv("LOG_LEVEL")
}
