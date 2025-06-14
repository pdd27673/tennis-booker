package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"

	"tennis-booker/internal/database"
	"tennis-booker/internal/retention"
)

// RetentionServiceApp manages the retention service application
type RetentionServiceApp struct {
	retentionService *retention.RetentionService
	logger           *log.Logger
	config           AppConfig
	db               *mongo.Database
}

// AppConfig holds application-level configuration
type AppConfig struct {
	// Scheduling
	CronExpression string
	RunOnce        bool

	// Retention configuration
	RetentionConfig retention.RetentionConfig

	// Database
	MongoURI     string
	DatabaseName string

	// Monitoring
	EnableMetrics     bool
	MetricsOutputFile string
	LogLevel          string
	LogFormat         string // "text" or "json"

	// Operational
	GracefulShutdownTimeout time.Duration
}

// DefaultAppConfig returns sensible defaults for the application
func DefaultAppConfig() AppConfig {
	return AppConfig{
		CronExpression:          "0 3 * * *", // Daily at 3 AM UTC
		RunOnce:                 false,
		RetentionConfig:         retention.DefaultRetentionConfig(),
		DatabaseName:            "tennis_booker",
		EnableMetrics:           true,
		MetricsOutputFile:       "/var/log/retention-metrics.json",
		LogLevel:                "info",
		LogFormat:               "json",
		GracefulShutdownTimeout: 30 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() AppConfig {
	config := DefaultAppConfig()

	// Scheduling
	if cronExpr := os.Getenv("RETENTION_CRON_EXPRESSION"); cronExpr != "" {
		config.CronExpression = cronExpr
	}

	if runOnce := os.Getenv("RETENTION_RUN_ONCE"); runOnce == "true" {
		config.RunOnce = true
	}

	// Retention configuration
	if retentionWindow := os.Getenv("RETENTION_WINDOW_HOURS"); retentionWindow != "" {
		if hours, err := strconv.Atoi(retentionWindow); err == nil {
			config.RetentionConfig.RetentionWindow = time.Duration(hours) * time.Hour
		}
	}

	if batchSize := os.Getenv("RETENTION_BATCH_SIZE"); batchSize != "" {
		if size, err := strconv.Atoi(batchSize); err == nil {
			config.RetentionConfig.BatchSize = size
		}
	}

	if dryRun := os.Getenv("RETENTION_DRY_RUN"); dryRun == "true" {
		config.RetentionConfig.DryRun = true
	}

	if logLevel := os.Getenv("RETENTION_LOG_LEVEL"); logLevel != "" {
		config.RetentionConfig.LogLevel = logLevel
		config.LogLevel = logLevel
	}

	// Database
	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		config.MongoURI = mongoURI
	}

	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		config.DatabaseName = dbName
	}

	// Monitoring
	if enableMetrics := os.Getenv("RETENTION_ENABLE_METRICS"); enableMetrics == "false" {
		config.EnableMetrics = false
		config.RetentionConfig.EnableMetrics = false
	}

	if metricsFile := os.Getenv("RETENTION_METRICS_FILE"); metricsFile != "" {
		config.MetricsOutputFile = metricsFile
	}

	if logFormat := os.Getenv("RETENTION_LOG_FORMAT"); logFormat != "" {
		config.LogFormat = logFormat
	}

	// Operational
	if timeout := os.Getenv("RETENTION_SHUTDOWN_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.GracefulShutdownTimeout = duration
		}
	}

	return config
}

// NewRetentionServiceApp creates a new retention service application
func NewRetentionServiceApp(config AppConfig) (*RetentionServiceApp, error) {
	// Configure logger
	logger := configureLogger(config.LogLevel, config.LogFormat)

	// Connect to database
	var db *mongo.Database
	var err error

	if config.MongoURI != "" {
		// Use provided URI
		db, err = database.InitDatabase(config.MongoURI, config.DatabaseName)
	} else {
		// Try to get from Vault, fallback to environment
		connectionManager, connErr := database.NewConnectionManagerFromEnv()
		if connErr == nil {
			db, err = connectionManager.ConnectWithFallback()
		} else {
			logger.Printf("⚠️ Failed to create database connection manager: %v", connErr)
			err = fmt.Errorf("no MongoDB connection available")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create retention service
	retentionService := retention.NewRetentionService(config.RetentionConfig, db, logger)

	// Validate configuration
	if err := retentionService.ValidateConfiguration(); err != nil {
		return nil, fmt.Errorf("invalid retention configuration: %w", err)
	}

	return &RetentionServiceApp{
		retentionService: retentionService,
		logger:           logger,
		config:           config,
		db:               db,
	}, nil
}

// Run starts the retention service application
func (app *RetentionServiceApp) Run(ctx context.Context) error {
	app.logger.Println("🚀 Starting Tennis Court Data Retention Service...")

	// Log configuration
	app.logConfiguration()

	if app.config.RunOnce {
		app.logger.Println("📋 Running in single execution mode...")
		return app.runRetentionCycle(ctx)
	}

	// Set up cron scheduler
	app.logger.Printf("⏰ Setting up scheduled execution with cron expression: %s", app.config.CronExpression)

	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(app.logger)))

	_, err := c.AddFunc(app.config.CronExpression, func() {
		app.logger.Println("⏰ Scheduled retention cycle starting...")

		// Create context with timeout for each run
		runCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := app.runRetentionCycle(runCtx); err != nil {
			app.logger.Printf("❌ Scheduled retention cycle failed: %v", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to schedule retention job: %w", err)
	}

	c.Start()
	app.logger.Println("✅ Retention service scheduled and running...")

	// Wait for shutdown signal
	<-ctx.Done()

	app.logger.Println("🛑 Shutting down retention service...")
	c.Stop()

	return nil
}

// runRetentionCycle executes a single retention cycle
func (app *RetentionServiceApp) runRetentionCycle(ctx context.Context) error {
	app.logger.Println("🔄 Starting retention cycle...")

	startTime := time.Now()

	// Run retention cycle
	metrics, err := app.retentionService.RunRetentionCycle(ctx)
	if err != nil {
		app.logger.Printf("❌ Retention cycle failed: %v", err)
		return err
	}

	// Log results
	app.logRetentionResults(metrics)

	// Save metrics if enabled
	if app.config.EnableMetrics {
		if err := app.saveMetrics(metrics); err != nil {
			app.logger.Printf("⚠️ Failed to save metrics: %v", err)
		}
	}

	duration := time.Since(startTime)
	app.logger.Printf("✅ Retention cycle completed in %v", duration)

	return nil
}

// logConfiguration logs the current configuration
func (app *RetentionServiceApp) logConfiguration() {
	config := app.config

	if config.LogFormat == "json" {
		configJSON, _ := json.Marshal(map[string]interface{}{
			"retention_window": config.RetentionConfig.RetentionWindow.String(),
			"batch_size":       config.RetentionConfig.BatchSize,
			"dry_run":          config.RetentionConfig.DryRun,
			"cron_expression":  config.CronExpression,
			"run_once":         config.RunOnce,
			"enable_metrics":   config.EnableMetrics,
			"log_level":        config.LogLevel,
		})
		app.logger.Printf("📋 Configuration: %s", string(configJSON))
	} else {
		app.logger.Printf("📋 Configuration:")
		app.logger.Printf("  - Retention Window: %v", config.RetentionConfig.RetentionWindow)
		app.logger.Printf("  - Batch Size: %d", config.RetentionConfig.BatchSize)
		app.logger.Printf("  - Dry Run: %v", config.RetentionConfig.DryRun)
		app.logger.Printf("  - Cron Expression: %s", config.CronExpression)
		app.logger.Printf("  - Run Once: %v", config.RunOnce)
		app.logger.Printf("  - Enable Metrics: %v", config.EnableMetrics)
		app.logger.Printf("  - Log Level: %s", config.LogLevel)
	}
}

// logRetentionResults logs the results of a retention cycle
func (app *RetentionServiceApp) logRetentionResults(metrics *retention.RetentionMetrics) {
	if app.config.LogFormat == "json" {
		metricsJSON, _ := json.Marshal(map[string]interface{}{
			"event":                         "retention_cycle_completed",
			"duration":                      metrics.Duration.String(),
			"candidate_slots_found":         metrics.CandidateSlotsFound,
			"slots_checked_against_prefs":   metrics.SlotsCheckedAgainstPrefs,
			"slots_identified_for_deletion": metrics.SlotsIdentifiedForDeletion,
			"slots_actually_deleted":        metrics.SlotsActuallyDeleted,
			"active_preferences_count":      metrics.ActivePreferencesCount,
			"errors_encountered":            metrics.ErrorsEncountered,
			"dry_run_mode":                  metrics.DryRunMode,
			"timestamp":                     time.Now().UTC().Format(time.RFC3339),
		})
		app.logger.Printf("📊 %s", string(metricsJSON))
	} else {
		app.logger.Printf("📊 Retention Cycle Results:")
		app.logger.Printf("  - Duration: %v", metrics.Duration)
		app.logger.Printf("  - Candidate Slots Found: %d", metrics.CandidateSlotsFound)
		app.logger.Printf("  - Slots Checked Against Preferences: %d", metrics.SlotsCheckedAgainstPrefs)
		app.logger.Printf("  - Slots Identified for Deletion: %d", metrics.SlotsIdentifiedForDeletion)
		app.logger.Printf("  - Slots Actually Deleted: %d", metrics.SlotsActuallyDeleted)
		app.logger.Printf("  - Active Preferences Count: %d", metrics.ActivePreferencesCount)
		app.logger.Printf("  - Errors Encountered: %d", metrics.ErrorsEncountered)
		app.logger.Printf("  - Dry Run Mode: %v", metrics.DryRunMode)
	}
}

// saveMetrics saves metrics to a file for monitoring systems
func (app *RetentionServiceApp) saveMetrics(metrics *retention.RetentionMetrics) error {
	metricsData := map[string]interface{}{
		"timestamp":                     time.Now().UTC().Format(time.RFC3339),
		"start_time":                    metrics.StartTime.UTC().Format(time.RFC3339),
		"end_time":                      metrics.EndTime.UTC().Format(time.RFC3339),
		"duration_seconds":              metrics.Duration.Seconds(),
		"candidate_slots_found":         metrics.CandidateSlotsFound,
		"slots_checked_against_prefs":   metrics.SlotsCheckedAgainstPrefs,
		"slots_identified_for_deletion": metrics.SlotsIdentifiedForDeletion,
		"slots_actually_deleted":        metrics.SlotsActuallyDeleted,
		"active_preferences_count":      metrics.ActivePreferencesCount,
		"errors_encountered":            metrics.ErrorsEncountered,
		"dry_run_mode":                  metrics.DryRunMode,
		"retention_window_hours":        app.config.RetentionConfig.RetentionWindow.Hours(),
		"batch_size":                    app.config.RetentionConfig.BatchSize,
	}

	jsonData, err := json.MarshalIndent(metricsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(strings.TrimSuffix(app.config.MetricsOutputFile, "/retention-metrics.json"), 0755); err != nil {
		return fmt.Errorf("failed to create metrics directory: %w", err)
	}

	// Write metrics to file
	if err := os.WriteFile(app.config.MetricsOutputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	app.logger.Printf("📊 Metrics saved to %s", app.config.MetricsOutputFile)
	return nil
}

// configureLogger sets up the logger based on configuration
func configureLogger(logLevel, logFormat string) *log.Logger {
	var prefix string
	if logFormat == "json" {
		prefix = ""
	} else {
		prefix = "[RETENTION-SERVICE] "
	}

	logger := log.New(os.Stdout, prefix, log.LstdFlags)

	// Set log level (basic implementation)
	if logLevel == "debug" {
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	return logger
}

// handleTestMode runs the service in test mode
func handleTestMode(logger *log.Logger) error {
	logger.Println("🧪 Running in test mode...")

	config := LoadConfigFromEnv()
	config.RunOnce = true
	config.RetentionConfig.DryRun = true
	config.LogLevel = "debug"
	config.RetentionConfig.LogLevel = "debug"

	app, err := NewRetentionServiceApp(config)
	if err != nil {
		return fmt.Errorf("failed to create retention service app: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	return app.runRetentionCycle(ctx)
}

// handleDryRunMode runs the service in dry-run mode
func handleDryRunMode(logger *log.Logger) error {
	logger.Println("🔍 Running in dry-run mode...")

	config := LoadConfigFromEnv()
	config.RunOnce = true
	config.RetentionConfig.DryRun = true

	app, err := NewRetentionServiceApp(config)
	if err != nil {
		return fmt.Errorf("failed to create retention service app: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	return app.runRetentionCycle(ctx)
}

func main() {
	// Load environment variables
	godotenv.Load()

	// Configure basic logger for startup
	logger := log.New(os.Stdout, "[RETENTION-SERVICE] ", log.LstdFlags)

	// Handle command line arguments
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "test":
			if err := handleTestMode(logger); err != nil {
				logger.Fatalf("❌ Test mode failed: %v", err)
			}
			return
		case "dry-run":
			if err := handleDryRunMode(logger); err != nil {
				logger.Fatalf("❌ Dry-run mode failed: %v", err)
			}
			return
		case "help", "-h", "--help":
			fmt.Println("Tennis Court Data Retention Service")
			fmt.Println("")
			fmt.Println("Usage:")
			fmt.Println("  retention-service              Run in scheduled mode")
			fmt.Println("  retention-service test         Run in test mode (dry-run with debug logging)")
			fmt.Println("  retention-service dry-run      Run once in dry-run mode")
			fmt.Println("  retention-service help         Show this help message")
			fmt.Println("")
			fmt.Println("Environment Variables:")
			fmt.Println("  RETENTION_CRON_EXPRESSION      Cron expression for scheduling (default: '0 3 * * *')")
			fmt.Println("  RETENTION_RUN_ONCE             Run once and exit (default: false)")
			fmt.Println("  RETENTION_WINDOW_HOURS          Hours before slots are eligible for deletion (default: 168)")
			fmt.Println("  RETENTION_BATCH_SIZE            Batch size for deletions (default: 1000)")
			fmt.Println("  RETENTION_DRY_RUN               Enable dry-run mode (default: false)")
			fmt.Println("  RETENTION_LOG_LEVEL             Log level: info, debug (default: info)")
			fmt.Println("  RETENTION_LOG_FORMAT            Log format: text, json (default: json)")
			fmt.Println("  RETENTION_ENABLE_METRICS        Enable metrics collection (default: true)")
			fmt.Println("  RETENTION_METRICS_FILE          Metrics output file (default: /var/log/retention-metrics.json)")
			fmt.Println("  MONGO_URI                       MongoDB connection URI")
			fmt.Println("  DATABASE_NAME                   Database name (default: tennis_booker)")
			return
		default:
			logger.Fatalf("❌ Unknown command: %s. Use 'help' for usage information.", os.Args[1])
		}
	}

	// Load configuration
	config := LoadConfigFromEnv()

	// Create application
	app, err := NewRetentionServiceApp(config)
	if err != nil {
		logger.Fatalf("❌ Failed to create retention service: %v", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Println("🛑 Received shutdown signal...")
		cancel()
	}()

	// Run the application
	if err := app.Run(ctx); err != nil {
		logger.Fatalf("❌ Retention service failed: %v", err)
	}

	logger.Println("👋 Retention service stopped")
}
