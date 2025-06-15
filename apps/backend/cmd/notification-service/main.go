package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"tennis-booker/internal/database"
	"tennis-booker/internal/secrets"
)

// User represents user preferences for notifications
type User struct {
	ID                  primitive.ObjectID `bson:"_id"`
	Email               string             `bson:"email"`
	Name                string             `bson:"name"`
	PreferredVenues     []string           `bson:"preferredVenues"`
	TimePreferences     TimePreferences    `bson:"timePreferences"`
	MaxPrice            float64            `bson:"maxPrice"`
	NotificationEnabled bool               `bson:"notificationEnabled"`
	CreatedAt           time.Time          `bson:"createdAt"`
	UpdatedAt           time.Time          `bson:"updatedAt"`
}

type TimePreferences struct {
	WeekdaySlots []TimeSlot `bson:"weekdaySlots"`
	WeekendSlots []TimeSlot `bson:"weekendSlots"`
}

type TimeSlot struct {
	Start string `bson:"start"`
	End   string `bson:"end"`
}

// SlotData represents a tennis court slot
type SlotData struct {
	VenueID     string    `json:"venueId"`
	VenueName   string    `json:"venueName"`
	Platform    string    `json:"platform"`
	CourtID     string    `json:"courtId"`
	CourtName   string    `json:"courtName"`
	Date        string    `json:"date"`
	StartTime   string    `json:"startTime"`
	EndTime     string    `json:"endTime"`
	Price       float64   `json:"price"`
	IsAvailable bool      `json:"isAvailable"`
	BookingURL  string    `json:"bookingUrl"`
	ScrapedAt   time.Time `json:"scrapedAt"`
}

// NotificationService handles the notification processing
type NotificationService struct {
	db          *mongo.Database
	redisClient *redis.Client
	logger      *log.Logger
	users       []User
	slotBatch   map[string][]SlotData // User email -> list of slots
	batchMutex  sync.RWMutex
	batchTimer  *time.Timer
}

// GmailService handles Gmail SMTP email notifications
type GmailService struct {
	smtpHost     string
	smtpPort     string
	fromEmail    string
	fromPassword string
	fromName     string
	logger       *log.Logger
}

// NewGmailService creates a new Gmail SMTP service
func NewGmailService(email, password, fromName string, logger *log.Logger) *GmailService {
	return &GmailService{
		smtpHost:     "smtp.gmail.com",
		smtpPort:     "587",
		fromEmail:    email,
		fromPassword: password,
		fromName:     fromName,
		logger:       logger,
	}
}

// NewGmailServiceFromVault creates a Gmail service using credentials from Vault
func NewGmailServiceFromVault(secretsManager *secrets.SecretsManager, logger *log.Logger) (*GmailService, error) {
	email, password, smtpHost, smtpPort, err := secretsManager.GetEmailCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get email credentials from vault: %w", err)
	}

	// Use defaults if not provided
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}

	return &GmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		fromEmail:    email,
		fromPassword: password,
		fromName:     "Tennis Court Alerts",
		logger:       logger,
	}, nil
}

// SendCourtAvailabilityAlert sends email notification via Gmail SMTP
func (g *GmailService) SendCourtAvailabilityAlert(toEmail, courtDetails, bookingLink string) error {
	// Detect if this is a batched notification (multiple courts)
	var subject string
	if strings.Contains(courtDetails, " courts just became available") {
		subject = "🎾 Multiple Tennis Courts Available!"
	} else {
		subject = "🎾 Tennis Court Available!"
	}

	body := fmt.Sprintf(`%s

🔗 Primary booking link: %s

---
Tennis Court Booking Alert System
`, courtDetails, bookingLink)

	// Send email via Gmail SMTP
	return g.sendEmail(toEmail, subject, body)
}

func (g *GmailService) sendEmail(toEmail, subject, body string) error {
	// Gmail SMTP configuration
	auth := smtp.PlainAuth("", g.fromEmail, g.fromPassword, g.smtpHost)

	// Compose message
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", toEmail, subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%s", g.smtpHost, g.smtpPort)
	err := smtp.SendMail(addr, auth, g.fromEmail, []string{toEmail}, []byte(msg))

	if err != nil {
		g.logger.Printf("❌ Failed to send email to %s: %v", toEmail, err)
		return err
	}

	g.logger.Printf("✅ Email sent successfully to %s", toEmail)
	return nil
}

// SendTestEmail sends a test email
func (g *GmailService) SendTestEmail(toEmail string) error {
	testDetails := fmt.Sprintf(`🎾 TEST NOTIFICATION

Venue: Test Tennis Club
Court: Test Court 1
Date: %s
Time: 19:00-20:00
Price: £15.00`, time.Now().Format("2006-01-02"))

	g.logger.Printf("📧 [TEST EMAIL] Sending test notification to %s", toEmail)
	return g.SendCourtAvailabilityAlert(toEmail, testDetails, "https://example.com/book")
}

func main() {
	// Load environment variables from multiple possible locations
	godotenv.Load()
	godotenv.Load(".env")
	godotenv.Load("../.env")
	godotenv.Load("../../.env")

	// Configure logging
	logger := log.New(os.Stdout, "[NOTIFICATION-SERVICE] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 Starting Tennis Court Notification Service...")

	// Check for test mode
	if len(os.Args) > 1 && os.Args[1] == "test" {
		logger.Println("📧 Running in test mode - sending test email...")

		// Try to use Vault for test email
		secretsManager, err := secrets.NewSecretsManagerFromEnv()
		if err != nil {
			logger.Printf("⚠️ Failed to connect to Vault for test: %v", err)
			logger.Println("🔄 Using fallback credentials for test...")

			// Fallback to environment variables (no hardcoded credentials)
			email := os.Getenv("GMAIL_EMAIL")
			password := os.Getenv("GMAIL_PASSWORD")

			if email == "" || password == "" {
				logger.Printf("❌ Test mode requires either Vault or GMAIL_EMAIL/GMAIL_PASSWORD environment variables")
				os.Exit(1)
			}

			gmailService := NewGmailService(email, password, "Tennis Court Alerts", logger)

			// Use the configured email for testing
			if err := gmailService.SendTestEmail(email); err != nil {
				logger.Printf("❌ Test email failed: %v", err)
				os.Exit(1)
			} else {
				logger.Println("✅ Test email sent successfully!")
				os.Exit(0)
			}
		} else {
			defer secretsManager.Close()

			gmailService, err := NewGmailServiceFromVault(secretsManager, logger)
			if err != nil {
				logger.Printf("❌ Failed to create Gmail service from Vault: %v", err)
				os.Exit(1)
			}

			// Use the same email that's configured in Vault for testing
			if err := gmailService.SendTestEmail(gmailService.fromEmail); err != nil {
				logger.Printf("❌ Test email failed: %v", err)
				os.Exit(1)
			} else {
				logger.Println("✅ Test email sent successfully using Vault credentials!")
				os.Exit(0)
			}
		}
	}

	// Initialize database connection using Vault
	connectionManager, err := database.NewConnectionManagerFromEnv()
	if err != nil {
		logger.Printf("⚠️ Failed to create database connection manager: %v", err)
		logger.Println("🔄 Attempting fallback connection...")

		// Fallback to environment variables
		mongoURI := getEnvWithDefault("MONGO_URI", "")
		if mongoURI == "" {
			// Build from individual components
			username := getEnvWithDefault("MONGO_ROOT_USERNAME", "")
			password := getEnvWithDefault("MONGO_ROOT_PASSWORD", "")
			host := getEnvWithDefault("MONGO_HOST", "localhost")
			port := getEnvWithDefault("MONGO_PORT", "27017")
			
			if username != "" && password != "" {
				mongoURI = fmt.Sprintf("mongodb://%s:%s@%s:%s?authSource=admin", username, password, host, port)
			} else {
				mongoURI = fmt.Sprintf("mongodb://%s:%s", host, port)
			}
		}
		dbName := getEnvWithDefault("DB_NAME", "tennis_booking")

		db, err := database.InitDatabase(mongoURI, dbName)
		if err != nil {
			logger.Fatalf("Failed to connect to MongoDB with fallback: %v", err)
		}
		logger.Println("✅ Connected to MongoDB using fallback credentials")

		// Continue with the rest of the initialization using fallback
		initializeServiceWithFallback(db, logger)
		return
	}
	defer connectionManager.Close()

	// Connect to database using Vault credentials
	db, err := connectionManager.ConnectWithFallback()
	if err != nil {
		logger.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	logger.Println("✅ Connected to MongoDB")

	// Get secrets manager for other credentials
	secretsManager := connectionManager.GetSecretsManager()

	// Initialize Redis connection using Vault
	redisHost, redisPassword, err := secretsManager.GetRedisCredentials()
	if err != nil {
		logger.Printf("⚠️ Failed to get Redis credentials from Vault: %v", err)
		logger.Println("🔄 Using fallback Redis credentials...")
		redisHost = getEnvWithDefault("REDIS_ADDR", "localhost:6379")
		redisPassword = getEnvWithDefault("REDIS_PASSWORD", "password")
	}

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.Println("✅ Connected to Redis")

	// Initialize Gmail service using Vault
	logger.Println("🔐 Attempting to get email credentials from Vault...")
	gmailService, err := NewGmailServiceFromVault(secretsManager, logger)
	if err != nil {
		logger.Printf("⚠️ Failed to create Gmail service from Vault: %v", err)
		logger.Println("🔄 Attempting to use environment variables for email credentials...")

		email := os.Getenv("GMAIL_EMAIL")
		password := os.Getenv("GMAIL_PASSWORD")

		if email == "" || password == "" {
			logger.Fatalf("❌ Failed to get email credentials from Vault and no GMAIL_EMAIL/GMAIL_PASSWORD environment variables set")
		}

		gmailService = NewGmailService(email, password, "Tennis Court Alerts", logger)
		logger.Println("✅ Using email credentials from environment variables")
	} else {
		logger.Println("✅ Successfully retrieved email credentials from Vault")
	}

	// Create notification service
	service := &NotificationService{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}

	// Load users
	if err := service.loadUsers(); err != nil {
		logger.Fatalf("Failed to load users: %v", err)
	}

	// Log service status
	service.logServiceStatus()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start notification engine in a goroutine
	go func() {
		service.startNotificationEngine(gmailService)
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Println("🛑 Shutdown signal received, stopping notification service...")

	// Cleanup
	redisClient.Close()
	logger.Println("✅ Notification service stopped gracefully")
}

// initializeServiceWithFallback initializes the service using fallback credentials
func initializeServiceWithFallback(db *mongo.Database, logger *log.Logger) {
	// Load environment variables again to ensure they're available
	godotenv.Load("../../.env")
	
	// Get configuration from environment
	redisAddr := getEnvWithDefault("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnvWithDefault("REDIS_PASSWORD", "password")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	logger.Println("✅ Connected to Redis")

	// Initialize Gmail service from environment variables
	email := os.Getenv("GMAIL_EMAIL")
	password := os.Getenv("GMAIL_PASSWORD")



	if email == "" || password == "" {
		logger.Fatalf("❌ GMAIL_EMAIL and GMAIL_PASSWORD environment variables are required for fallback mode")
	}

	gmailService := NewGmailService(email, password, "Tennis Court Alerts", logger)
	logger.Println("✅ Using email credentials from environment variables")

	// Create notification service
	service := &NotificationService{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}

	// Load users
	if err := service.loadUsers(); err != nil {
		logger.Fatalf("Failed to load users: %v", err)
	}

	// Log service status
	service.logServiceStatus()

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start notification engine in a goroutine
	go func() {
		service.startNotificationEngine(gmailService)
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Println("🛑 Shutdown signal received, stopping notification service...")

	// Cleanup
	redisClient.Close()
	logger.Println("✅ Notification service stopped gracefully")
}

// loadUsers loads user preferences from MongoDB
func (s *NotificationService) loadUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query user_preferences collection for users with notifications enabled
	filter := bson.M{
		"notification_settings.email": true,
		"notification_settings.unsubscribed": bson.M{"$ne": true},
	}

	cursor, err := s.db.Collection("user_preferences").Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	// Load user preferences and convert to User struct
	var userPrefs []struct {
		ID                   primitive.ObjectID `bson:"_id"`
		UserID               primitive.ObjectID `bson:"user_id"`
		Times                []struct {
			Start string `bson:"start"`
			End   string `bson:"end"`
		} `bson:"times"`
		WeekdayTimes         []struct {
			Start string `bson:"start"`
			End   string `bson:"end"`
		} `bson:"weekday_times"`
		WeekendTimes         []struct {
			Start string `bson:"start"`
			End   string `bson:"end"`
		} `bson:"weekend_times"`
		MaxPrice             float64  `bson:"max_price"`
		PreferredVenues      []string `bson:"preferred_venues"`
		NotificationSettings struct {
			Email        bool   `bson:"email"`
			EmailAddress string `bson:"email_address"`
		} `bson:"notification_settings"`
	}

	if err := cursor.All(ctx, &userPrefs); err != nil {
		return err
	}

	// Convert to User structs and get user details
	s.users = []User{}
	for _, pref := range userPrefs {
		// Get user details from users collection
		var userDoc struct {
			Email string `bson:"email"`
			Name  string `bson:"name"`
		}
		
		userFilter := bson.M{"_id": pref.UserID}
		err := s.db.Collection("users").FindOne(ctx, userFilter).Decode(&userDoc)
		if err != nil {
			s.logger.Printf("⚠️ Failed to load user details for user_id %s: %v", pref.UserID.Hex(), err)
			continue
		}

		// Convert time preferences to the expected format
		var weekdaySlots, weekendSlots []TimeSlot
		
		// Use new weekday/weekend specific times if available
		if len(pref.WeekdayTimes) > 0 || len(pref.WeekendTimes) > 0 {
			// Use the new separate weekday/weekend times
			for _, timeRange := range pref.WeekdayTimes {
				weekdaySlots = append(weekdaySlots, TimeSlot{
					Start: timeRange.Start,
					End:   timeRange.End,
				})
			}
			for _, timeRange := range pref.WeekendTimes {
				weekendSlots = append(weekendSlots, TimeSlot{
					Start: timeRange.Start,
					End:   timeRange.End,
				})
			}
		} else if len(pref.Times) > 0 {
			// Fallback to legacy times field (treat as both weekday and weekend)
			for _, timeRange := range pref.Times {
				slot := TimeSlot{
					Start: timeRange.Start,
					End:   timeRange.End,
				}
				weekdaySlots = append(weekdaySlots, slot)
				weekendSlots = append(weekendSlots, slot)
			}
		}

		user := User{
			ID:    pref.UserID,
			Email: userDoc.Email,
			Name:  userDoc.Name,
			PreferredVenues: pref.PreferredVenues,
			TimePreferences: TimePreferences{
				WeekdaySlots: weekdaySlots,
				WeekendSlots: weekendSlots,
			},
			MaxPrice:            pref.MaxPrice,
			NotificationEnabled: true, // We already filtered for this
		}

		// Use email from notification settings if available, otherwise from user doc
		if pref.NotificationSettings.EmailAddress != "" {
			user.Email = pref.NotificationSettings.EmailAddress
		}

		s.users = append(s.users, user)
	}

	s.logger.Printf("✅ Loaded %d users with notifications enabled", len(s.users))
	return nil
}

// startNotificationEngine starts listening for Redis notifications with batching
func (s *NotificationService) startNotificationEngine(gmailService *GmailService) {
	s.logger.Println("🔔 Starting notification engine - listening for court slots...")
	s.slotBatch = make(map[string][]SlotData)

	for {
		// Block and wait for messages from Redis queue
		result, err := s.redisClient.BRPop(context.Background(), 0, "court_slots").Result()
		if err != nil {
			s.logger.Printf("Error reading from Redis queue: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// result[0] is the queue name, result[1] is the data
		if len(result) > 1 {
			s.addSlotToBatch(result[1], gmailService)
		}
	}
}

// addSlotToBatch adds a slot to the batching system
func (s *NotificationService) addSlotToBatch(slotJSON string, gmailService *GmailService) {
	var slot SlotData
	if err := json.Unmarshal([]byte(slotJSON), &slot); err != nil {
		s.logger.Printf("Error parsing slot JSON: %v", err)
		return
	}

	s.logger.Printf("Processing slot: %s at %s, %s %s--%s, £%.2f",
		slot.VenueName, slot.CourtName, slot.Date, slot.StartTime, slot.EndTime, slot.Price)

	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()

	// Check each user's preferences and add to their batch if matches
	for _, user := range s.users {
		if s.matchesUserPreferences(user, slot) {
			// Check for duplicates
			if s.isDuplicateNotification(user, slot) {
				s.logger.Printf("Skipping duplicate notification for user: %s", user.Email)
				continue
			}

			s.logger.Printf("Slot matches preferences for user: %s", user.Email)

			// Add to batch
			if s.slotBatch[user.Email] == nil {
				s.slotBatch[user.Email] = make([]SlotData, 0)
			}
			s.slotBatch[user.Email] = append(s.slotBatch[user.Email], slot)

			// Reset/start the batch timer (10 seconds)
			if s.batchTimer != nil {
				s.batchTimer.Stop()
			}
			s.batchTimer = time.AfterFunc(10*time.Second, func() {
				s.sendBatchedNotifications(gmailService)
			})
		}
	}
}

// sendBatchedNotifications sends all batched slots as consolidated emails
func (s *NotificationService) sendBatchedNotifications(gmailService *GmailService) {
	s.batchMutex.Lock()
	currentBatch := s.slotBatch
	s.slotBatch = make(map[string][]SlotData) // Reset batch
	s.batchMutex.Unlock()

	for email, slots := range currentBatch {
		if len(slots) > 0 {
			// Find user by email
			var user User
			for _, u := range s.users {
				if u.Email == email {
					user = u
					break
				}
			}

			// Send consolidated notification
			if err := s.sendBatchedNotification(user, slots, gmailService); err != nil {
				s.logger.Printf("Error sending batched notification to %s: %v", email, err)
			} else {
				s.logger.Printf("✅ Batched notification sent to %s (%d slots)", email, len(slots))

				// Record all notifications
				for _, slot := range slots {
					s.recordNotification(user, slot)
				}
			}
		}
	}
}

// processSlotNotification processes a slot notification (kept for compatibility)
func (s *NotificationService) processSlotNotification(slotJSON string, gmailService *GmailService) {
	var slot SlotData
	if err := json.Unmarshal([]byte(slotJSON), &slot); err != nil {
		s.logger.Printf("Error parsing slot JSON: %v", err)
		return
	}

	s.logger.Printf("Processing slot: %s at %s, %s %s--%s, £%.2f",
		slot.VenueName, slot.CourtName, slot.Date, slot.StartTime, slot.EndTime, slot.Price)

	// Check each user's preferences
	for _, user := range s.users {
		if s.matchesUserPreferences(user, slot) {
			s.logger.Printf("Slot matches preferences for user: %s", user.Email)

			// Check for duplicates
			if s.isDuplicateNotification(user, slot) {
				s.logger.Printf("Skipping duplicate notification for user: %s", user.Email)
				continue
			}

			// Send notification
			if err := s.sendNotification(user, slot, gmailService); err != nil {
				s.logger.Printf("Error sending notification to %s: %v", user.Email, err)
			} else {
				s.logger.Printf("✅ Notification sent to %s", user.Email)
				s.recordNotification(user, slot)
			}
		}
	}
}

// matchesUserPreferences checks if a slot matches user preferences
func (s *NotificationService) matchesUserPreferences(user User, slot SlotData) bool {
	// Check venue preference
	venueMatch := false
	for _, venue := range user.PreferredVenues {
		if venue == slot.VenueName {
			venueMatch = true
			break
		}
	}
	if !venueMatch {
		return false
	}

	// Check price
	if slot.Price > user.MaxPrice {
		return false
	}

	// Check time preferences
	return s.matchesTimePreferences(user.TimePreferences, slot)
}

// matchesTimePreferences checks if slot time matches user preferences
func (s *NotificationService) matchesTimePreferences(prefs TimePreferences, slot SlotData) bool {
	// Parse slot date to determine if it's a weekend
	slotTime, err := time.Parse("2006-01-02", slot.Date)
	if err != nil {
		s.logger.Printf("Error parsing slot date: %v", err)
		return false
	}

	var relevantSlots []TimeSlot
	if slotTime.Weekday() == time.Saturday || slotTime.Weekday() == time.Sunday {
		relevantSlots = prefs.WeekendSlots
	} else {
		relevantSlots = prefs.WeekdaySlots
	}

	// Check if slot time falls within any preferred time slot
	for _, timeSlot := range relevantSlots {
		if s.timeInRange(slot.StartTime, timeSlot.Start, timeSlot.End) {
			return true
		}
	}

	return false
}

// timeInRange checks if a time falls within a range
func (s *NotificationService) timeInRange(timeStr, start, end string) bool {
	slotTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return false
	}

	startTime, err := time.Parse("15:04", start)
	if err != nil {
		return false
	}

	endTime, err := time.Parse("15:04", end)
	if err != nil {
		return false
	}

	return (slotTime.After(startTime) || slotTime.Equal(startTime)) && slotTime.Before(endTime)
}

// isDuplicateNotification checks if this notification was already sent
func (s *NotificationService) isDuplicateNotification(user User, slot SlotData) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create unique identifier for this slot
	slotID := fmt.Sprintf("%s_%s_%s_%s", slot.VenueID, slot.CourtID, slot.Date, slot.StartTime)

	// Check if notification was sent in the last 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	filter := bson.M{
		"userId": user.ID,
		"slotId": slotID,
		"sentAt": bson.M{"$gte": cutoff},
	}

	count, err := s.db.Collection("notification_history").CountDocuments(ctx, filter)
	if err != nil {
		s.logger.Printf("Error checking duplicate notification: %v", err)
		return false
	}

	return count > 0
}

// recordNotification records that a notification was sent
func (s *NotificationService) recordNotification(user User, slot SlotData) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slotID := fmt.Sprintf("%s_%s_%s_%s", slot.VenueID, slot.CourtID, slot.Date, slot.StartTime)

	record := bson.M{
		"userId":    user.ID,
		"userEmail": user.Email,
		"slotId":    slotID,
		"venueId":   slot.VenueID,
		"venueName": slot.VenueName,
		"courtName": slot.CourtName,
		"date":      slot.Date,
		"startTime": slot.StartTime,
		"price":     slot.Price,
		"sentAt":    time.Now(),
	}

	_, err := s.db.Collection("notification_history").InsertOne(ctx, record)
	if err != nil {
		s.logger.Printf("Error recording notification: %v", err)
	}
}

// sendNotification sends an email notification
func (s *NotificationService) sendNotification(user User, slot SlotData, gmailService *GmailService) error {
	courtDetails := fmt.Sprintf(`Venue: %s
Court: %s
Date: %s
Time: %s--%s
Price: £%.2f`,
		slot.VenueName,
		slot.CourtName,
		slot.Date,
		slot.StartTime,
		slot.EndTime,
		slot.Price)

	return gmailService.SendCourtAvailabilityAlert(user.Email, courtDetails, slot.BookingURL)
}

// sendBatchedNotification sends a consolidated email for multiple slots
func (s *NotificationService) sendBatchedNotification(user User, slots []SlotData, gmailService *GmailService) error {
	if len(slots) == 0 {
		return nil
	}

	// Group slots by venue and date for better organization
	venueGroups := make(map[string]map[string][]SlotData)
	for _, slot := range slots {
		if venueGroups[slot.VenueName] == nil {
			venueGroups[slot.VenueName] = make(map[string][]SlotData)
		}
		if venueGroups[slot.VenueName][slot.Date] == nil {
			venueGroups[slot.VenueName][slot.Date] = make([]SlotData, 0)
		}
		venueGroups[slot.VenueName][slot.Date] = append(venueGroups[slot.VenueName][slot.Date], slot)
	}

	// Build consolidated details
	var courtDetails strings.Builder
	slotCount := len(slots)

	if slotCount == 1 {
		courtDetails.WriteString("🎾 A tennis court just became available!\n\n")
	} else {
		courtDetails.WriteString(fmt.Sprintf("🎾 %d tennis courts just became available!\n\n", slotCount))
	}

	// Add booking links section at the top for quick access
	courtDetails.WriteString("🔗 QUICK BOOKING LINKS:\n")
	for i, slot := range slots {
		courtDetails.WriteString(fmt.Sprintf("  %d. %s %s %s-%s: %s\n",
			i+1, slot.VenueName, slot.CourtName, slot.StartTime, slot.EndTime, slot.BookingURL))
	}
	courtDetails.WriteString("\n📋 COURT DETAILS:\n")

	// Organize by venue and date
	for venueName, dates := range venueGroups {
		courtDetails.WriteString(fmt.Sprintf("\n🏟️ %s:\n", venueName))

		for date, venueSlots := range dates {
			courtDetails.WriteString(fmt.Sprintf("  📅 %s:\n", date))

			for _, slot := range venueSlots {
				courtDetails.WriteString(fmt.Sprintf("    • %s: %s-%s (£%.2f)\n",
					slot.CourtName, slot.StartTime, slot.EndTime, slot.Price))
			}
		}
	}

	courtDetails.WriteString("\n⚡ These slots just became available - book quickly!")

	// Use the first slot's booking URL as the primary link (they should all be for the same venue group anyway)
	primaryBookingURL := slots[0].BookingURL

	return gmailService.SendCourtAvailabilityAlert(user.Email, courtDetails.String(), primaryBookingURL)
}

// SendTestNotification sends a test notification
func (s *NotificationService) SendTestNotification(email string, gmailService *GmailService) error {
	return gmailService.SendTestEmail(email)
}

// logServiceStatus logs the current status of services
func (s *NotificationService) logServiceStatus() {
	s.logger.Println("📊 Service Status:")
	s.logger.Println("  ✅ Email Service: ENABLED (Gmail SMTP Real)")
	s.logger.Println("  ✅ Redis Listener: ENABLED")
	s.logger.Println("  ✅ MongoDB Connection: ENABLED")
	s.logger.Println("  ✅ Duplicate Prevention: ENABLED")
	s.logger.Printf("  ✅ Users Loaded: %d", len(s.users))

	if len(s.users) > 0 {
		user := s.users[0]
		s.logger.Printf("  📧 Monitoring for: %s", user.Email)
		s.logger.Printf("  🏟️ Preferred venues: %v", user.PreferredVenues)
		s.logger.Printf("  ⏰ Weekday slots: %v", user.TimePreferences.WeekdaySlots)
		s.logger.Printf("  🌅 Weekend slots: %v", user.TimePreferences.WeekendSlots)
		s.logger.Printf("  💰 Max price: £%.2f", user.MaxPrice)
	}
}

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
