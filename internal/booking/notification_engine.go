package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"tennis-booking-bot/internal/database"
	"tennis-booking-bot/internal/email"
	"tennis-booking-bot/internal/models"
)

// NotificationEngine handles court availability notifications
type NotificationEngine struct {
	db                   *mongo.Database
	redisClient          *database.RedisClient
	preferenceService    *models.PreferenceService
	alertService         *models.AlertHistoryService
	deduplicationService *models.DeduplicationService
	logger               *log.Logger
	ctx                  context.Context
	cancel               context.CancelFunc
	emailService         EmailService // Interface for sending emails
}

// EmailService interface for sending notification emails
type EmailService interface {
	SendCourtAvailabilityAlert(ctx context.Context, toEmail string, data *email.CourtAvailabilityData) error
	IsEnabled() bool
}

// NotificationConfig holds configuration for the notification engine
type NotificationConfig struct {
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	SlotChannelName string        // Redis channel to subscribe to for slot updates
	PollInterval    time.Duration // How often to poll scraping logs as backup
}

// NewNotificationEngine creates a new notification engine instance
func NewNotificationEngine(db *mongo.Database, config *NotificationConfig, emailService EmailService, logger *log.Logger) (*NotificationEngine, error) {
	// Initialize Redis client
	redisClient, err := database.NewRedisClient(config.RedisAddr, config.RedisPassword, config.RedisDB)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis client: %w", err)
	}

	// Initialize services
	preferenceService := models.NewPreferenceService(db)
	alertService := models.NewAlertHistoryService(db)
	deduplicationService := models.NewDeduplicationService(db)

	// Create context for the engine
	ctx, cancel := context.WithCancel(context.Background())

	return &NotificationEngine{
		db:                   db,
		redisClient:          redisClient,
		preferenceService:    preferenceService,
		alertService:         alertService,
		deduplicationService: deduplicationService,
		emailService:         emailService,
		logger:               logger,
		ctx:                  ctx,
		cancel:               cancel,
	}, nil
}

// Start begins the notification engine operations
func (e *NotificationEngine) Start() error {
	e.logger.Println("Starting notification engine...")

	// Start Redis listener in a goroutine
	go e.listenForCourtAvailability()

	// Start periodic scraping log poller as backup in a goroutine
	go e.pollScrapingLogs()

	// Start cleanup routine for old alerts
	go e.runPeriodicCleanup()

	e.logger.Println("Notification engine started successfully")
	return nil
}

// Stop gracefully shuts down the notification engine
func (e *NotificationEngine) Stop() error {
	e.logger.Println("Stopping notification engine...")
	e.cancel()
	
	if e.redisClient != nil {
		if err := e.redisClient.Close(); err != nil {
			e.logger.Printf("Error closing Redis client: %v", err)
		}
	}
	
	e.logger.Println("Notification engine stopped")
	return nil
}

// listenForCourtAvailability listens for court availability messages from Redis
func (e *NotificationEngine) listenForCourtAvailability() {
	e.logger.Println("Starting Redis court availability listener...")
	
	// Subscribe to the court availability channel
	pubsub := e.redisClient.Client.Subscribe(e.ctx, "court:availability")
	defer pubsub.Close()

	// Listen for messages
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			var event models.CourtAvailabilityEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				e.logger.Printf("Error unmarshaling court availability event: %v", err)
				continue
			}
			
			e.logger.Printf("Received court availability: %s at %s %s-%s for £%.2f", 
				event.CourtName, event.Date, event.StartTime, event.EndTime, event.Price)
			
			// Process the availability event
			go e.processAvailabilityEvent(event)
			
		case <-e.ctx.Done():
			e.logger.Println("Redis listener stopping...")
			return
		}
	}
}

// pollScrapingLogs periodically checks the ScrapingLogs collection for new available slots as backup
func (e *NotificationEngine) pollScrapingLogs() {
	e.logger.Println("Starting scraping logs poller...")
	
	ticker := time.NewTicker(5 * time.Minute) // Poll every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.checkScrapingLogsForNewSlots()
		case <-e.ctx.Done():
			e.logger.Println("Scraping logs poller stopping...")
			return
		}
	}
}

// runPeriodicCleanup removes old alert history records
func (e *NotificationEngine) runPeriodicCleanup() {
	e.logger.Println("Starting periodic cleanup routine...")
	
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Clean up old alert history
			deleted, err := e.alertService.CleanupOldAlerts(e.ctx, 30) // Remove alerts older than 30 days
			if err != nil {
				e.logger.Printf("Error during alert cleanup: %v", err)
			} else {
				e.logger.Printf("Cleaned up %d old alert records", deleted)
			}

			// Clean up expired deduplication records
			dedupDeleted, err := e.deduplicationService.CleanupExpiredRecords(e.ctx)
			if err != nil {
				e.logger.Printf("Error during deduplication cleanup: %v", err)
			} else {
				e.logger.Printf("Cleaned up %d expired deduplication records", dedupDeleted)
			}
		case <-e.ctx.Done():
			e.logger.Println("Cleanup routine stopping...")
			return
		}
	}
}

// processAvailabilityEvent processes a court availability event and sends notifications
func (e *NotificationEngine) processAvailabilityEvent(event models.CourtAvailabilityEvent) {
	e.logger.Printf("Processing availability event for %s - %s", event.VenueName, event.CourtName)

	// Get all users with notification preferences
	users, err := e.getUsersWithNotificationPreferences()
	if err != nil {
		e.logger.Printf("Error getting users with preferences: %v", err)
		return
	}

	e.logger.Printf("Checking %d users for matches", len(users))

	// Check each user's preferences against this slot
	for _, user := range users {
		if e.shouldNotifyUser(event, user) {
			e.logger.Printf("Sending notification to user %s for %s", user.NotificationSettings.EmailAddress, event.CourtName)
			go e.sendNotificationToUser(event, user)
		}
	}
}

// getUsersWithNotificationPreferences retrieves all users who have notification preferences set up
func (e *NotificationEngine) getUsersWithNotificationPreferences() ([]models.UserPreferences, error) {
	collection := e.db.Collection("user_preferences")
	
	// Find users who have email notifications enabled and are not unsubscribed
	filter := bson.M{
		"notification_settings.email": true,
		"notification_settings.unsubscribed": bson.M{"$ne": true},
		"notification_settings.email_address": bson.M{"$exists": true, "$ne": ""},
	}
	
	cursor, err := collection.Find(e.ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(e.ctx)
	
	var users []models.UserPreferences
	if err = cursor.All(e.ctx, &users); err != nil {
		return nil, err
	}
	
	return users, nil
}

// shouldNotifyUser determines if a user should be notified about this court availability
func (e *NotificationEngine) shouldNotifyUser(event models.CourtAvailabilityEvent, user models.UserPreferences) bool {
	// Check if user has instant alerts enabled
	if !user.NotificationSettings.InstantAlerts {
		return false
	}

	// Check if we're within the user's alert time window
	if !e.isWithinAlertTimeWindow(user.NotificationSettings) {
		return false
	}

	// Check venue preferences
	if !e.venueMatches(event.VenueID, user) {
		return false
	}

	// Check time preferences
	if !e.timeMatches(event.StartTime, event.EndTime, user) {
		return false
	}

	// Check day preferences
	if !e.dayMatches(event.Date, user) {
		return false
	}

	// Check price preferences
	if user.MaxPrice > 0 && event.Price > user.MaxPrice {
		return false
	}

	// Check rate limiting
	if e.isRateLimited(user) {
		return false
	}

	// Enhanced duplicate prevention check
	duplicateResult, err := e.deduplicationService.CheckForDuplicate(e.ctx, user.UserID, event)
	if err != nil {
		e.logger.Printf("Error checking for duplicates: %v", err)
		return false // Err on the side of caution
	}
	
	if duplicateResult.IsDuplicate {
		e.logger.Printf("Skipping duplicate notification for user %s: %s (reason: %s)", 
			user.NotificationSettings.EmailAddress, duplicateResult.ReasonDescription, duplicateResult.ReasonCode)
		return false
	}

	return true
}

// isWithinAlertTimeWindow checks if current time is within the user's alert time window
func (e *NotificationEngine) isWithinAlertTimeWindow(settings models.NotificationSettings) bool {
	if settings.AlertTimeWindowStart == "" || settings.AlertTimeWindowEnd == "" {
		return true // No time window restriction
	}

	now := time.Now()
	currentTime := now.Format("15:04")

	return currentTime >= settings.AlertTimeWindowStart && currentTime <= settings.AlertTimeWindowEnd
}

// venueMatches checks if the venue matches user preferences
func (e *NotificationEngine) venueMatches(venueID string, user models.UserPreferences) bool {
	// If user has excluded venues, check if this venue is excluded
	for _, excluded := range user.ExcludedVenues {
		if excluded == venueID {
			return false
		}
	}

	// If user has preferred venues, check if this venue is in the list
	if len(user.PreferredVenues) > 0 {
		for _, preferred := range user.PreferredVenues {
			if preferred == venueID {
				return true
			}
		}
		return false // Venue not in preferred list
	}

	return true // No venue restrictions or venue not excluded
}

// timeMatches checks if the slot time matches user preferences
func (e *NotificationEngine) timeMatches(startTime, endTime string, user models.UserPreferences) bool {
	if len(user.Times) == 0 {
		return true // No time restrictions
	}

	for _, timeRange := range user.Times {
		if e.timeInRange(startTime, timeRange.Start, timeRange.End) {
			return true
		}
	}

	return false
}

// dayMatches checks if the slot day matches user preferences
func (e *NotificationEngine) dayMatches(dateStr string, user models.UserPreferences) bool {
	if len(user.PreferredDays) == 0 {
		return true // No day restrictions
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}

	dayName := strings.ToLower(date.Weekday().String())
	for _, preferredDay := range user.PreferredDays {
		if strings.ToLower(preferredDay) == dayName {
			return true
		}
	}

	return false
}

// isRateLimited checks if user has exceeded rate limits
func (e *NotificationEngine) isRateLimited(user models.UserPreferences) bool {
	// Check hourly limit
	if user.NotificationSettings.MaxAlertsPerHour > 0 {
		hourlyCount, err := e.alertService.GetUserAlertCount(e.ctx, user.UserID, 1)
		if err != nil {
			e.logger.Printf("Error checking hourly rate limit: %v", err)
			return true // Err on the side of caution
		}
		if hourlyCount >= int64(user.NotificationSettings.MaxAlertsPerHour) {
			return true
		}
	}

	// Check daily limit
	if user.NotificationSettings.MaxAlertsPerDay > 0 {
		dailyCount, err := e.alertService.GetUserAlertCount(e.ctx, user.UserID, 24)
		if err != nil {
			e.logger.Printf("Error checking daily rate limit: %v", err)
			return true // Err on the side of caution
		}
		if dailyCount >= int64(user.NotificationSettings.MaxAlertsPerDay) {
			return true
		}
	}

	return false
}

// sendNotificationToUser sends an email notification to a user about court availability
func (e *NotificationEngine) sendNotificationToUser(event models.CourtAvailabilityEvent, user models.UserPreferences) {
	// Check if email service is enabled
	if !e.emailService.IsEnabled() {
		e.logger.Printf("Email service not enabled, skipping notification for %s", user.NotificationSettings.EmailAddress)
		return
	}

	// Generate unsubscribe URL (placeholder for now)
	unsubscribeURL := fmt.Sprintf("https://tennisbooker.com/unsubscribe?user=%s", user.UserID.Hex())

	// Prepare email data
	emailData := &email.CourtAvailabilityData{
		UserName:          e.extractNameFromEmail(user.NotificationSettings.EmailAddress),
		VenueName:         event.VenueName,
		VenueLocation:     "", // Could be enhanced with venue location data
		CourtType:         event.CourtName,
		AvailableDate:     e.formatDate(event.Date),
		AvailableTime:     fmt.Sprintf("%s - %s", event.StartTime, event.EndTime),
		Duration:          e.calculateDuration(event.StartTime, event.EndTime),
		Price:             e.formatPrice(event.Price, event.Currency),
		BookingURL:        event.BookingURL,
		UnsubscribeURL:    unsubscribeURL,
		AlertType:         "new_slot", // Could be enhanced to detect cancellations vs new slots
		NotificationTime:  time.Now().Format("15:04 on January 2, 2006"),
	}

	// Send email
	err := e.emailService.SendCourtAvailabilityAlert(e.ctx, user.NotificationSettings.EmailAddress, emailData)

	// Create alert history record
	alert := &models.AlertHistory{
		UserID:        user.UserID,
		VenueID:       event.VenueID,
		VenueName:     event.VenueName,
		CourtID:       event.CourtID,
		CourtName:     event.CourtName,
		SlotDate:      event.Date,
		SlotStartTime: event.StartTime,
		SlotEndTime:   event.EndTime,
		Price:         event.Price,
		Currency:      event.Currency,
		BookingURL:    event.BookingURL,
		EmailAddress:  user.NotificationSettings.EmailAddress,
		SlotKey:       event.GenerateSlotKey(),
		EmailStatus:   "sent",
	}

	if err != nil {
		e.logger.Printf("Error sending email to %s: %v", user.NotificationSettings.EmailAddress, err)
		alert.EmailStatus = "failed"
	} else {
		e.logger.Printf("Successfully sent notification to %s", user.NotificationSettings.EmailAddress)
	}

	// Save alert history
	if err := e.alertService.CreateAlert(e.ctx, alert); err != nil {
		e.logger.Printf("Error saving alert history: %v", err)
	}

	// Record notification in deduplication service (only if email was sent successfully)
	if err == nil {
		if dedupErr := e.deduplicationService.RecordNotification(e.ctx, user.UserID, event); dedupErr != nil {
			e.logger.Printf("Error recording notification for deduplication: %v", dedupErr)
		}
	}
}

// formatDate formats a date string for display
func (e *NotificationEngine) formatDate(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return date.Format("Monday, January 2, 2006")
}

// calculateDuration calculates the duration between start and end times
func (e *NotificationEngine) calculateDuration(startTime, endTime string) string {
	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return "Unknown duration"
	}
	
	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return "Unknown duration"
	}
	
	duration := end.Sub(start)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	
	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%d hour %d minutes", hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d hour", hours)
	} else {
		return fmt.Sprintf("%d minutes", minutes)
	}
}

// formatPrice formats a price with currency
func (e *NotificationEngine) formatPrice(price float64, currency string) string {
	if currency == "" {
		currency = "GBP"
	}
	
	switch currency {
	case "GBP":
		return fmt.Sprintf("£%.2f", price)
	case "USD":
		return fmt.Sprintf("$%.2f", price)
	case "EUR":
		return fmt.Sprintf("€%.2f", price)
	default:
		return fmt.Sprintf("%.2f %s", price, currency)
	}
}

// extractNameFromEmail extracts a display name from an email address
func (e *NotificationEngine) extractNameFromEmail(emailAddr string) string {
	if emailAddr == "" {
		return "Tennis Player"
	}
	
	// Split on @ and take the part before
	parts := strings.Split(emailAddr, "@")
	if len(parts) == 0 {
		return "Tennis Player"
	}
	
	username := parts[0]
	
	// Replace common separators with spaces and title case
	username = strings.ReplaceAll(username, ".", " ")
	username = strings.ReplaceAll(username, "_", " ")
	username = strings.ReplaceAll(username, "-", " ")
	
	// Title case each word
	words := strings.Fields(username)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	
	result := strings.Join(words, " ")
	if result == "" {
		return "Tennis Player"
	}
	
	return result
}

// timeInRange checks if a time falls within a time range
func (e *NotificationEngine) timeInRange(timeStr, startStr, endStr string) bool {
	// Parse times
	timeTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return false
	}
	
	startTime, err := time.Parse("15:04", startStr)
	if err != nil {
		return false
	}
	
	endTime, err := time.Parse("15:04", endStr)
	if err != nil {
		return false
	}
	
	// Check if time is within range
	return (timeTime.Equal(startTime) || timeTime.After(startTime)) && 
		   (timeTime.Equal(endTime) || timeTime.Before(endTime))
}

// checkScrapingLogsForNewSlots checks for new available slots in the ScrapingLogs collection as backup
func (e *NotificationEngine) checkScrapingLogsForNewSlots() {
	// Get scraping logs from the last hour with available slots
	since := time.Now().Add(-1 * time.Hour)
	
	collection := e.db.Collection("scraping_logs")
	cursor, err := collection.Find(e.ctx, bson.M{
		"scrape_timestamp": bson.M{"$gte": since},
		"slots_found": bson.M{"$gt": 0},
		"success": true,
	})
	
	if err != nil {
		e.logger.Printf("Error querying scraping logs: %v", err)
		return
	}
	defer cursor.Close(e.ctx)

	var logs []models.ScrapingLog
	if err = cursor.All(e.ctx, &logs); err != nil {
		e.logger.Printf("Error decoding scraping logs: %v", err)
		return
	}

	for _, log := range logs {
		// Process each slot in the log
		for _, slot := range log.SlotsFound {
			// Parse time range to get start and end times
			timeRange := slot.Time // Format: "HH:MM-HH:MM"
			timeParts := strings.Split(timeRange, "-")
			startTime := ""
			endTime := ""
			if len(timeParts) == 2 {
				startTime = timeParts[0]
				endTime = timeParts[1]
			}
			
			// Convert slot to availability event
			event := models.CourtAvailabilityEvent{
				VenueID:      log.VenueID.Hex(),
				VenueName:    log.VenueName,
				CourtID:      slot.CourtID,
				CourtName:    slot.Court,
				Date:         slot.Date,
				StartTime:    startTime,
				EndTime:      endTime,
				Price:        slot.Price,
				Currency:     "GBP", // Default currency
				BookingURL:   slot.URL,
				DiscoveredAt: log.ScrapeTimestamp,
				ScrapeLogID:  log.ID.Hex(),
			}
			
			// Process the event (only if it's recent to avoid spam)
			if time.Since(log.ScrapeTimestamp) < 10*time.Minute {
				go e.processAvailabilityEvent(event)
			}
		}
	}
} 