package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

	"tennis-booking-bot/internal/booking"
	"tennis-booking-bot/internal/database"
	"tennis-booking-bot/internal/email"
	"tennis-booking-bot/internal/models"
)

// MockEmailService is a placeholder email service for development/testing
type MockEmailService struct {
	logger *log.Logger
}

func NewMockEmailService(logger *log.Logger) *MockEmailService {
	return &MockEmailService{logger: logger}
}

func (m *MockEmailService) SendCourtAvailabilityAlert(toEmail, courtDetails, bookingLink string) error {
	m.logger.Printf("📧 MOCK EMAIL SENT TO: %s", toEmail)
	m.logger.Printf("📄 COURT DETAILS:\n%s", courtDetails)
	m.logger.Printf("🔗 BOOKING LINK: %s", bookingLink)
	m.logger.Printf("------")
	return nil
}

func main() {
	logger := log.New(os.Stdout, "[NOTIFICATION-SERVICE] ", log.LstdFlags)
	logger.Println("Starting Tennis Court Notification Service...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Println("Warning: .env file not found, using system environment variables")
	}

	// Get database configuration from environment
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:YOUR_PASSWORD@localhost:27017"
		logger.Println("Using default MongoDB URI")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "tennis_booking"
		logger.Println("Using default database name: tennis_booking")
	}

	// Get Redis configuration from environment
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
		logger.Println("Using default Redis address")
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	// Redis DB defaults to 0

	// Initialize database connection
	logger.Println("Connecting to MongoDB...")
	db, err := database.InitDatabase(mongoURI, dbName)
	if err != nil {
		logger.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	logger.Println("Connected to MongoDB successfully")

	// Initialize notification engine configuration
	notificationConfig := &booking.NotificationConfig{
		RedisAddr:       redisAddr,
		RedisPassword:   redisPassword,
		RedisDB:         0,
		SlotChannelName: "court:availability",
		PollInterval:    5 * time.Minute,
	}

	// Initialize email service with SendGrid
	emailConfig := &email.Config{
		APIKey:     os.Getenv("SENDGRID_API_KEY"),
		FromEmail:  getEnvWithDefault("FROM_EMAIL", "alerts@tennisbooker.com"),
		FromName:   getEnvWithDefault("FROM_NAME", "Tennis Court Alerts"),
		TemplateID: os.Getenv("SENDGRID_TEMPLATE_ID"),
	}

	emailClient := email.NewClient(emailConfig)
	emailService := email.NewServiceAdapter(emailClient)

	// Initialize database indexes
	if err := initializeIndexes(db, logger); err != nil {
		logger.Printf("Warning: Failed to create database indexes: %v", err)
	}

	// Initialize notification engine
	logger.Println("Initializing notification engine...")
	engine, err := booking.NewNotificationEngine(db, notificationConfig, emailService, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize notification engine: %v", err)
	}

	// Log service status
	logServiceStatus(emailService, logger)

	// Start the notification engine
	if err := engine.Start(); err != nil {
		logger.Fatalf("Failed to start notification engine: %v", err)
	}

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	logger.Println("🚀 Tennis Court Availability Alert System is running!")
	logger.Println("📧 Monitoring for court availability and sending instant email alerts...")
	logger.Println("Press Ctrl+C to stop.")
	<-c

	logger.Println("Shutting down notification service...")
	if err := engine.Stop(); err != nil {
		logger.Printf("Error stopping notification engine: %v", err)
	}

	logger.Println("✅ Tennis Court Availability Alert System stopped gracefully.")
}

// initializeIndexes creates necessary database indexes
func initializeIndexes(db *mongo.Database, logger *log.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Println("Creating database indexes...")

	// Initialize deduplication service indexes
	deduplicationService := models.NewDeduplicationService(db)
	if err := deduplicationService.CreateIndexes(ctx); err != nil {
		return err
	}

	// Note: AlertHistoryService doesn't need custom indexes as it uses the default collection setup

	logger.Println("✅ Database indexes created successfully")
	return nil
}

// logServiceStatus logs the current status of services
func logServiceStatus(emailService *email.ServiceAdapter, logger *log.Logger) {
	logger.Println("📊 Service Status:")
	
	if emailService.IsEnabled() {
		logger.Println("  ✅ Email Service: ENABLED (SendGrid)")
	} else {
		logger.Println("  ⚠️  Email Service: DISABLED (No API key configured)")
		logger.Println("     Set SENDGRID_API_KEY environment variable to enable email notifications")
	}

	logger.Println("  ✅ Redis Listener: ENABLED")
	logger.Println("  ✅ MongoDB Connection: ENABLED")
	logger.Println("  ✅ Duplicate Prevention: ENABLED")
	logger.Println("  ✅ Rate Limiting: ENABLED")
	logger.Println("  ✅ Alert History Tracking: ENABLED")
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 