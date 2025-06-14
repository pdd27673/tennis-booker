package models

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DeduplicationService provides advanced duplicate prevention for notifications
type DeduplicationService struct {
	collection *mongo.Collection
}

// NewDeduplicationService creates a new deduplication service
func NewDeduplicationService(db *mongo.Database) *DeduplicationService {
	return &DeduplicationService{
		collection: db.Collection("notification_deduplication"),
	}
}

// DeduplicationRecord tracks sent notifications to prevent duplicates
type DeduplicationRecord struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	SlotKey       string             `bson:"slot_key" json:"slot_key"`
	ContentHash   string             `bson:"content_hash" json:"content_hash"`
	VenueID       string             `bson:"venue_id" json:"venue_id"`
	CourtID       string             `bson:"court_id" json:"court_id"`
	SlotDate      string             `bson:"slot_date" json:"slot_date"`
	SlotStartTime string             `bson:"slot_start_time" json:"slot_start_time"`
	Price         float64            `bson:"price" json:"price"`
	FirstSentAt   time.Time          `bson:"first_sent_at" json:"first_sent_at"`
	LastSentAt    time.Time          `bson:"last_sent_at" json:"last_sent_at"`
	SendCount     int                `bson:"send_count" json:"send_count"`
	ExpiresAt     time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// DuplicateCheckResult contains the result of a duplicate check
type DuplicateCheckResult struct {
	IsDuplicate       bool                 `json:"is_duplicate"`
	ExistingRecord    *DeduplicationRecord `json:"existing_record,omitempty"`
	ReasonCode        string               `json:"reason_code"`
	ReasonDescription string               `json:"reason_description"`
	TimeSinceLastSent time.Duration        `json:"time_since_last_sent"`
}

// CheckForDuplicate checks if a notification would be a duplicate
func (s *DeduplicationService) CheckForDuplicate(ctx context.Context, userID primitive.ObjectID, event CourtAvailabilityEvent) (*DuplicateCheckResult, error) {
	slotKey := event.GenerateSlotKey()
	contentHash := s.generateContentHash(event)

	// Check for exact slot match (same slot, same user)
	exactMatch, err := s.findExactMatch(ctx, userID, slotKey)
	if err != nil {
		return nil, err
	}

	if exactMatch != nil {
		timeSince := time.Since(exactMatch.LastSentAt)

		// Allow resending after 24 hours for the same slot
		if timeSince < 24*time.Hour {
			return &DuplicateCheckResult{
				IsDuplicate:       true,
				ExistingRecord:    exactMatch,
				ReasonCode:        "EXACT_SLOT_RECENT",
				ReasonDescription: "Same slot notification sent recently",
				TimeSinceLastSent: timeSince,
			}, nil
		}
	}

	// Check for similar content (same venue, court, time, different date)
	similarMatch, err := s.findSimilarMatch(ctx, userID, event, contentHash)
	if err != nil {
		return nil, err
	}

	if similarMatch != nil {
		timeSince := time.Since(similarMatch.LastSentAt)

		// Prevent spam of very similar notifications within 1 hour
		if timeSince < 1*time.Hour {
			return &DuplicateCheckResult{
				IsDuplicate:       true,
				ExistingRecord:    similarMatch,
				ReasonCode:        "SIMILAR_CONTENT_RECENT",
				ReasonDescription: "Very similar notification sent recently",
				TimeSinceLastSent: timeSince,
			}, nil
		}
	}

	// Check for venue flooding (too many notifications from same venue)
	venueCount, err := s.getRecentVenueNotificationCount(ctx, userID, event.VenueID, 1*time.Hour)
	if err != nil {
		return nil, err
	}

	if venueCount >= 5 { // Max 5 notifications per venue per hour
		return &DuplicateCheckResult{
			IsDuplicate:       true,
			ReasonCode:        "VENUE_FLOODING",
			ReasonDescription: "Too many notifications from this venue recently",
		}, nil
	}

	// Not a duplicate
	return &DuplicateCheckResult{
		IsDuplicate:       false,
		ReasonCode:        "NOT_DUPLICATE",
		ReasonDescription: "Notification is unique and can be sent",
	}, nil
}

// RecordNotification records that a notification was sent
func (s *DeduplicationService) RecordNotification(ctx context.Context, userID primitive.ObjectID, event CourtAvailabilityEvent) error {
	slotKey := event.GenerateSlotKey()
	contentHash := s.generateContentHash(event)
	now := time.Now()

	// Check if record already exists
	existing, err := s.findExactMatch(ctx, userID, slotKey)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing record
		update := bson.M{
			"$set": bson.M{
				"last_sent_at": now,
				"expires_at":   now.Add(48 * time.Hour), // Extend expiry
			},
			"$inc": bson.M{
				"send_count": 1,
			},
		}

		_, err = s.collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, update)
		return err
	}

	// Create new record
	record := &DeduplicationRecord{
		UserID:        userID,
		SlotKey:       slotKey,
		ContentHash:   contentHash,
		VenueID:       event.VenueID,
		CourtID:       event.CourtID,
		SlotDate:      event.Date,
		SlotStartTime: event.StartTime,
		Price:         event.Price,
		FirstSentAt:   now,
		LastSentAt:    now,
		SendCount:     1,
		ExpiresAt:     now.Add(48 * time.Hour), // Records expire after 48 hours
		CreatedAt:     now,
	}

	_, err = s.collection.InsertOne(ctx, record)
	return err
}

// CleanupExpiredRecords removes expired deduplication records
func (s *DeduplicationService) CleanupExpiredRecords(ctx context.Context) (int64, error) {
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now()},
	}

	result, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// GetUserNotificationStats returns notification statistics for a user
func (s *DeduplicationService) GetUserNotificationStats(ctx context.Context, userID primitive.ObjectID, since time.Time) (*NotificationStats, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"user_id":       userID,
				"first_sent_at": bson.M{"$gte": since},
			},
		},
		{
			"$group": bson.M{
				"_id":                 nil,
				"total_notifications": bson.M{"$sum": "$send_count"},
				"unique_slots":        bson.M{"$sum": 1},
				"venues":              bson.M{"$addToSet": "$venue_id"},
				"avg_price":           bson.M{"$avg": "$price"},
				"min_price":           bson.M{"$min": "$price"},
				"max_price":           bson.M{"$max": "$price"},
			},
		},
	}

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &NotificationStats{}, nil
	}

	result := results[0]
	venues, _ := result["venues"].(primitive.A)

	return &NotificationStats{
		TotalNotifications: getInt64(result, "total_notifications"),
		UniqueSlots:        getInt64(result, "unique_slots"),
		UniqueVenues:       int64(len(venues)),
		AveragePrice:       getFloat64(result, "avg_price"),
		MinPrice:           getFloat64(result, "min_price"),
		MaxPrice:           getFloat64(result, "max_price"),
		Period:             time.Since(since),
	}, nil
}

// NotificationStats contains notification statistics
type NotificationStats struct {
	TotalNotifications int64         `json:"total_notifications"`
	UniqueSlots        int64         `json:"unique_slots"`
	UniqueVenues       int64         `json:"unique_venues"`
	AveragePrice       float64       `json:"average_price"`
	MinPrice           float64       `json:"min_price"`
	MaxPrice           float64       `json:"max_price"`
	Period             time.Duration `json:"period"`
}

// findExactMatch finds an exact slot match for a user
func (s *DeduplicationService) findExactMatch(ctx context.Context, userID primitive.ObjectID, slotKey string) (*DeduplicationRecord, error) {
	filter := bson.M{
		"user_id":  userID,
		"slot_key": slotKey,
	}

	var record DeduplicationRecord
	err := s.collection.FindOne(ctx, filter).Decode(&record)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// findSimilarMatch finds a similar notification (same venue, court, time, different date)
func (s *DeduplicationService) findSimilarMatch(ctx context.Context, userID primitive.ObjectID, event CourtAvailabilityEvent, contentHash string) (*DeduplicationRecord, error) {
	filter := bson.M{
		"user_id":         userID,
		"venue_id":        event.VenueID,
		"court_id":        event.CourtID,
		"slot_start_time": event.StartTime,
		"slot_date":       bson.M{"$ne": event.Date},                      // Different date
		"last_sent_at":    bson.M{"$gte": time.Now().Add(-1 * time.Hour)}, // Within last hour
	}

	opts := options.FindOne().SetSort(bson.M{"last_sent_at": -1})

	var record DeduplicationRecord
	err := s.collection.FindOne(ctx, filter, opts).Decode(&record)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// getRecentVenueNotificationCount counts recent notifications from a venue
func (s *DeduplicationService) getRecentVenueNotificationCount(ctx context.Context, userID primitive.ObjectID, venueID string, since time.Duration) (int64, error) {
	filter := bson.M{
		"user_id":      userID,
		"venue_id":     venueID,
		"last_sent_at": bson.M{"$gte": time.Now().Add(-since)},
	}

	return s.collection.CountDocuments(ctx, filter)
}

// generateContentHash creates a hash of the notification content for similarity detection
func (s *DeduplicationService) generateContentHash(event CourtAvailabilityEvent) string {
	content := fmt.Sprintf("%s:%s:%s:%.2f",
		event.VenueID, event.CourtID, event.StartTime, event.Price)
	return fmt.Sprintf("%x", md5.Sum([]byte(content)))
}

// Helper functions for type conversion
func getInt64(m bson.M, key string) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int32:
			return int64(v)
		case int:
			return int64(v)
		}
	}
	return 0
}

func getFloat64(m bson.M, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0
}

// CreateIndexes creates necessary indexes for the deduplication collection
func (s *DeduplicationService) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "slot_key", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "expires_at", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "venue_id", Value: 1},
				{Key: "last_sent_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "venue_id", Value: 1},
				{Key: "court_id", Value: 1},
				{Key: "slot_start_time", Value: 1},
			},
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}
