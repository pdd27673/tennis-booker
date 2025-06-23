package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of a log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a structured logger
type Logger struct {
	serviceName string
	minLevel    LogLevel
	logger      *log.Logger
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	ServiceName string                 `json:"service"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// New creates a new structured logger
func New(serviceName string) *Logger {
	minLevel := INFO
	if level := strings.ToUpper(os.Getenv("LOG_LEVEL")); level != "" {
		switch level {
		case "DEBUG":
			minLevel = DEBUG
		case "INFO":
			minLevel = INFO
		case "WARN":
			minLevel = WARN
		case "ERROR":
			minLevel = ERROR
		case "FATAL":
			minLevel = FATAL
		}
	}

	return &Logger{
		serviceName: serviceName,
		minLevel:    minLevel,
		logger:      log.New(os.Stdout, "", 0),
	}
}

// shouldLog checks if a message should be logged based on the minimum level
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.minLevel
}

// log writes a structured log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp:   time.Now().UTC(),
		Level:       level.String(),
		ServiceName: l.serviceName,
		Message:     message,
		Fields:      fields,
	}

	// Check if we should use JSON formatting (production) or human-readable (development)
	if os.Getenv("LOG_FORMAT") == "json" {
		jsonData, err := json.Marshal(entry)
		if err != nil {
			l.logger.Printf("ERROR: Failed to marshal log entry: %v", err)
			return
		}
		l.logger.Println(string(jsonData))
	} else {
		// Human-readable format for development
		var fieldsStr string
		if len(fields) > 0 {
			fieldsStr = fmt.Sprintf(" %+v", fields)
		}
		l.logger.Printf("[%s] %s: %s%s", level.String(), l.serviceName, message, fieldsStr)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.log(DEBUG, message, fieldMap)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.log(INFO, message, fieldMap)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.log(WARN, message, fieldMap)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.log(ERROR, message, fieldMap)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields ...map[string]interface{}) {
	var fieldMap map[string]interface{}
	if len(fields) > 0 {
		fieldMap = fields[0]
	}
	l.log(FATAL, message, fieldMap)
	os.Exit(1)
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *FieldLogger {
	return &FieldLogger{
		logger: l,
		fields: fields,
	}
}

// FieldLogger represents a logger with predefined fields
type FieldLogger struct {
	logger *Logger
	fields map[string]interface{}
}

// Debug logs a debug message with predefined fields
func (fl *FieldLogger) Debug(message string) {
	fl.logger.log(DEBUG, message, fl.fields)
}

// Info logs an info message with predefined fields
func (fl *FieldLogger) Info(message string) {
	fl.logger.log(INFO, message, fl.fields)
}

// Warn logs a warning message with predefined fields
func (fl *FieldLogger) Warn(message string) {
	fl.logger.log(WARN, message, fl.fields)
}

// Error logs an error message with predefined fields
func (fl *FieldLogger) Error(message string) {
	fl.logger.log(ERROR, message, fl.fields)
}

// Fatal logs a fatal message with predefined fields and exits
func (fl *FieldLogger) Fatal(message string) {
	fl.logger.log(FATAL, message, fl.fields)
	os.Exit(1)
}

// Convenience functions for common use cases

// StartupInfo logs application startup information
func (l *Logger) StartupInfo(message string, port string, env string) {
	l.Info(message, map[string]interface{}{
		"port":        port,
		"environment": env,
		"event":       "startup",
	})
}

// ShutdownInfo logs application shutdown information
func (l *Logger) ShutdownInfo(message string, reason string) {
	l.Info(message, map[string]interface{}{
		"reason": reason,
		"event":  "shutdown",
	})
}

// ConnectionInfo logs connection establishment
func (l *Logger) ConnectionInfo(message string, connectionType string, host string) {
	l.Info(message, map[string]interface{}{
		"connection_type": connectionType,
		"host":            host,
		"event":           "connection",
	})
}

// ErrorWithCode logs an error with additional context
func (l *Logger) ErrorWithCode(message string, errorCode string, err error) {
	fields := map[string]interface{}{
		"error_code": errorCode,
		"event":      "error",
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.Error(message, fields)
}
