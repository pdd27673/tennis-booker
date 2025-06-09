package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"tennis-booker/internal/models"
)

// EventPublisher publishes court availability events to Redis
type EventPublisher struct {
	redisClient *redis.Client
	db          *mongo.Database
	logger      *log.Logger
	channel     string
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(redisClient *redis.Client, db *mongo.Database, logger *log.Logger) *EventPublisher {
	return &EventPublisher{
		redisClient: redisClient,
		db:          db,
		logger:      logger,
		channel:     "court:availability",
	}
}

// CourtSlot represents a single court availability slot from scraping
type CourtSlot struct {
	Date         string  `json:"date"`          // YYYY-MM-DD
	StartTime    string  `json:"start_time"`    // HH:MM
	EndTime      string  `json:"end_time"`      // HH:MM  
	CourtName    string  `json:"court_name"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	BookingURL   string  `json:"booking_url"`
	Available    bool    `json:"available"`
}

// ScrapingLogData represents the structure we expect from scraping logs
type ScrapingLogData struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	VenueID      string             `bson:"venue_id"`
	VenueName    string             `bson:"venue_name"`
	ProviderType string             `bson:"provider_type"`
	ScrapedAt    time.Time          `bson:"scraped_at"`
	SlotsFound   int                `bson:"slots_found"`
	Slots        []CourtSlot        `bson:"slots"`
	Status       string             `bson:"status"`
	Errors       []string           `bson:"errors,omitempty"`
}

// StartScrapingLogListener starts listening for new scraping logs and publishes availability events
func (p *EventPublisher) StartScrapingLogListener(ctx context.Context) error {
	p.logger.Println("Starting scraping log listener for court availability events...")

	// Use MongoDB Change Streams to watch for new scraping log inserts
	collection := p.db.Collection("scraping_logs")
	
	// Create pipeline to watch only inserts of successful scrapes
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"operationType": "insert",
				"fullDocument.status": "success",
				"fullDocument.slots_found": bson.M{"$gt": 0},
			},
		},
	}
	
	changeStream, err := collection.Watch(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to create change stream: %w", err)
	}
	defer changeStream.Close(ctx)

	p.logger.Println("Successfully started MongoDB change stream listener")

	// Process change events
	for changeStream.Next(ctx) {
		var event bson.M
		if err := changeStream.Decode(&event); err != nil {
			p.logger.Printf("Error decoding change stream event: %v", err)
			continue
		}

		// Extract the full document (new scraping log)
		fullDoc, ok := event["fullDocument"].(bson.M)
		if !ok {
			p.logger.Printf("Invalid fullDocument in change stream event")
			continue
		}

		// Convert to ScrapingLogData
		var scrapingLog ScrapingLogData
		docBytes, err := bson.Marshal(fullDoc)
		if err != nil {
			p.logger.Printf("Error marshaling document: %v", err)
			continue
		}

		if err := bson.Unmarshal(docBytes, &scrapingLog); err != nil {
			p.logger.Printf("Error unmarshaling scraping log: %v", err)
			continue
		}

		p.logger.Printf("Processing new scraping log: %s - %d slots found", scrapingLog.VenueName, scrapingLog.SlotsFound)

		// Process each slot and publish availability events
		p.processScrapingLogSlots(ctx, scrapingLog)
	}

	if err := changeStream.Err(); err != nil {
		return fmt.Errorf("change stream error: %w", err)
	}

	return nil
}

// PublishManualAvailabilityEvent publishes a manually created availability event
func (p *EventPublisher) PublishManualAvailabilityEvent(ctx context.Context, event *models.CourtAvailabilityEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.redisClient.Publish(ctx, p.channel, eventJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to Redis: %w", err)
	}

	p.logger.Printf("Published manual court availability event: %s at %s %s-%s", 
		event.CourtName, event.Date, event.StartTime, event.EndTime)
	
	return nil
}

// processScrapingLogSlots processes slots from a scraping log and publishes events
func (p *EventPublisher) processScrapingLogSlots(ctx context.Context, scrapingLog ScrapingLogData) {
	for _, slot := range scrapingLog.Slots {
		if !slot.Available {
			continue // Skip unavailable slots
		}

		// Create court availability event
		event := &models.CourtAvailabilityEvent{
			VenueID:      scrapingLog.VenueID,
			VenueName:    scrapingLog.VenueName,
			CourtID:      p.generateCourtID(scrapingLog.VenueID, slot.CourtName),
			CourtName:    slot.CourtName,
			Date:         slot.Date,
			StartTime:    slot.StartTime,
			EndTime:      slot.EndTime,
			Price:        slot.Price,
			Currency:     slot.Currency,
			BookingURL:   slot.BookingURL,
			DiscoveredAt: scrapingLog.ScrapedAt,
			ScrapeLogID:  scrapingLog.ID.Hex(),
		}

		// Check if this is a new availability (not seen in last 30 minutes)
		if p.isNewAvailability(ctx, event) {
			err := p.PublishManualAvailabilityEvent(ctx, event)
			if err != nil {
				p.logger.Printf("Error publishing availability event: %v", err)
			}
		}
	}
}

// generateCourtID creates a consistent court ID from venue and court name
func (p *EventPublisher) generateCourtID(venueID, courtName string) string {
	// Clean court name and create consistent ID
	cleanName := strings.ToLower(strings.ReplaceAll(courtName, " ", "_"))
	return fmt.Sprintf("%s_%s", venueID, cleanName)
}

// isNewAvailability checks if this court slot was recently seen to avoid spam
func (p *EventPublisher) isNewAvailability(ctx context.Context, event *models.CourtAvailabilityEvent) bool {
	// Use Redis to track recently seen slots
	slotKey := fmt.Sprintf("recent_slot:%s", event.GenerateSlotKey())
	
	// Check if we've seen this slot in the last 30 minutes
	exists := p.redisClient.Exists(ctx, slotKey).Val()
	if exists > 0 {
		return false // Not new
	}

	// Mark this slot as seen for 30 minutes
	p.redisClient.SetEx(ctx, slotKey, "1", 30*time.Minute)
	return true
}

// PublishTestEvent publishes a test court availability event for testing
func (p *EventPublisher) PublishTestEvent(ctx context.Context) error {
	testEvent := &models.CourtAvailabilityEvent{
		VenueID:      "test_venue_001",
		VenueName:    "Test Tennis Club",
		CourtID:      "test_venue_001_court_1",
		CourtName:    "Court 1",
		Date:         time.Now().AddDate(0, 0, 1).Format("2006-01-02"), // Tomorrow
		StartTime:    "18:00",
		EndTime:      "19:00",
		Price:        15.00,
		Currency:     "GBP",
		BookingURL:   "https://example.com/book/test",
		DiscoveredAt: time.Now(),
		ScrapeLogID:  "test_scrape_log",
	}

	return p.PublishManualAvailabilityEvent(ctx, testEvent)
}

// GetSubscriberCount returns the number of active subscribers to the court availability channel
func (p *EventPublisher) GetSubscriberCount(ctx context.Context) (int64, error) {
	cmd := p.redisClient.PubSubNumSub(ctx, p.channel)
	if cmd.Err() != nil {
		return 0, cmd.Err()
	}

	result := cmd.Val()
	if count, exists := result[p.channel]; exists {
		return count, nil
	}
	
	return 0, nil
}

// StartPollingFallback starts a fallback polling mechanism for scraping logs
// This is used when change streams are not available or as a backup
func (p *EventPublisher) StartPollingFallback(ctx context.Context, interval time.Duration) error {
	p.logger.Printf("Starting polling fallback every %v for new scraping logs...", interval)
	
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Track last processed timestamp
	lastProcessed := time.Now().Add(-24 * time.Hour) // Start from 24 hours ago

	for {
		select {
		case <-ticker.C:
			newLastProcessed, err := p.pollForNewLogs(ctx, lastProcessed)
			if err != nil {
				p.logger.Printf("Error polling for new logs: %v", err)
			} else {
				lastProcessed = newLastProcessed
			}
		case <-ctx.Done():
			p.logger.Println("Polling fallback stopping...")
			return nil
		}
	}
}

// pollForNewLogs polls the scraping_logs collection for new successful logs
func (p *EventPublisher) pollForNewLogs(ctx context.Context, since time.Time) (time.Time, error) {
	collection := p.db.Collection("scraping_logs")
	
	// Find logs created after the last processed time
	filter := bson.M{
		"scraped_at": bson.M{"$gt": since},
		"status": "success",
		"slots_found": bson.M{"$gt": 0},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return since, err
	}
	defer cursor.Close(ctx)

	newLastProcessed := since
	count := 0

	for cursor.Next(ctx) {
		var scrapingLog ScrapingLogData
		if err := cursor.Decode(&scrapingLog); err != nil {
			p.logger.Printf("Error decoding scraping log: %v", err)
			continue
		}

		// Update the last processed time
		if scrapingLog.ScrapedAt.After(newLastProcessed) {
			newLastProcessed = scrapingLog.ScrapedAt
		}

		// Process the slots
		p.processScrapingLogSlots(ctx, scrapingLog)
		count++
	}

	if count > 0 {
		p.logger.Printf("Processed %d new scraping logs via polling", count)
	}

	return newLastProcessed, nil
} 