package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// SendCourtAvailabilityAlert sends email notification via Gmail SMTP
func (g *GmailService) SendCourtAvailabilityAlert(toEmail, courtDetails, bookingLink string) error {
	// Compose email
	subject := "🎾 Tennis Court Available!"
	
	body := fmt.Sprintf(`🎾 A tennis court just became available!

%s

🔗 Book now: %s

This slot just became available - book quickly!

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
	// Load environment variables
	godotenv.Load()

	// Configure logging
	logger := log.New(os.Stdout, "[NOTIFICATION-SERVICE] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 Starting Tennis Court Notification Service...")

	// Check for test mode
	if len(os.Args) > 1 && os.Args[1] == "test" {
		logger.Println("📧 Running in test mode - sending test email...")
		gmailService := NewGmailService(
			getEnvWithDefault("GMAIL_EMAIL", "demo@example.com"),
			getEnvWithDefault("GMAIL_PASSWORD", "eswk jgaw zbet wgxo"),
			"Tennis Court Alerts",
			logger,
		)

		if err := gmailService.SendTestEmail("demo@example.com"); err != nil {
			logger.Printf("❌ Test email failed: %v", err)
			os.Exit(1)
		} else {
			logger.Println("✅ Test email sent successfully!")
			os.Exit(0)
		}
	}

	// Get configuration from environment
	mongoURI := getEnvWithDefault("MONGO_URI", "mongodb://admin:YOUR_PASSWORD@localhost:27017")
	dbName := getEnvWithDefault("DB_NAME", "tennis_booking")
	redisAddr := getEnvWithDefault("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnvWithDefault("REDIS_PASSWORD", "password")

	// Connect to MongoDB
	db, err := connectMongoDB(mongoURI, dbName)
	if err != nil {
		logger.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	logger.Println("✅ Connected to MongoDB")

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

	// Initialize Gmail service
	gmailService := NewGmailService(
		getEnvWithDefault("GMAIL_EMAIL", "demo@example.com"),
		getEnvWithDefault("GMAIL_PASSWORD", "eswk jgaw zbet wgxo"),
		"Tennis Court Alerts",
		logger,
	)

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

// connectMongoDB establishes connection to MongoDB
func connectMongoDB(uri, dbName string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client.Database(dbName), nil
}

// loadUsers loads user preferences from MongoDB
func (s *NotificationService) loadUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := s.db.Collection("users").Find(ctx, bson.M{"notificationEnabled": true})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	s.users = []User{}
	if err := cursor.All(ctx, &s.users); err != nil {
		return err
	}

	s.logger.Printf("✅ Loaded %d users with notifications enabled", len(s.users))
	return nil
}

// startNotificationEngine starts listening for Redis notifications
func (s *NotificationService) startNotificationEngine(gmailService *GmailService) {
	s.logger.Println("🔔 Starting notification engine - listening for court slots...")

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
			s.processSlotNotification(result[1], gmailService)
		}
	}
}

// processSlotNotification processes a slot notification
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