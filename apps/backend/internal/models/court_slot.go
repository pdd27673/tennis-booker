package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CourtSlot represents a tennis court time slot available for booking
type CourtSlot struct {
	ID            string             `json:"id" bson:"_id,omitempty"`                            // Unique identifier for the slot
	VenueID       primitive.ObjectID `json:"venue_id" bson:"venue_id"`                           // Reference to the venue
	VenueName     string             `json:"venue_name" bson:"venue_name"`                       // Venue name for convenience
	CourtID       string             `json:"court_id" bson:"court_id"`                           // Court identifier
	CourtName     string             `json:"court_name" bson:"court_name"`                       // Human-readable court name
	Date          string             `json:"date" bson:"date"`                                   // Format: "YYYY-MM-DD" (kept for backward compatibility)
	SlotDate      time.Time          `json:"slot_date" bson:"slot_date"`                         // Parsed date+time for efficient querying
	StartTime     string             `json:"start_time" bson:"start_time"`                       // Format: "HH:MM"
	EndTime       string             `json:"end_time" bson:"end_time"`                           // Format: "HH:MM"
	Price         float64            `json:"price" bson:"price"`                                 // Price for the slot
	Currency      string             `json:"currency" bson:"currency"`                           // Currency code (e.g., "GBP", "USD")
	Available     bool               `json:"available" bson:"available"`                         // Whether the slot is available
	BookingURL    string             `json:"booking_url" bson:"booking_url"`                     // Direct booking URL if available
	Provider      string             `json:"provider" bson:"provider"`                           // Provider type (e.g., "lta", "courtsides")
	LastScraped   time.Time          `json:"last_scraped" bson:"last_scraped"`                   // When this slot was last found
	NotifiedAt    *time.Time         `json:"notified_at,omitempty" bson:"notified_at,omitempty"` // When notification was sent for this slot (null if never notified)
	ScrapingLogID primitive.ObjectID `json:"scraping_log_id" bson:"scraping_log_id"`             // Reference to the scraping log
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`                       // When this slot record was created
	UpdatedAt     time.Time          `json:"updated_at" bson:"updated_at"`                       // When this slot record was last updated
}

// GenerateSlotID creates a unique identifier for a court slot
func (cs *CourtSlot) GenerateSlotID() string {
	return cs.VenueID.Hex() + "_" + cs.CourtID + "_" + cs.Date + "_" + cs.StartTime
}

// IsOlderThan checks if the slot is older than the specified duration
func (cs *CourtSlot) IsOlderThan(duration time.Duration) bool {
	return time.Since(cs.SlotDate) > duration
}

// HasBeenNotified returns true if a notification has been sent for this slot
func (cs *CourtSlot) HasBeenNotified() bool {
	return cs.NotifiedAt != nil
}

// MarkAsNotified sets the NotifiedAt timestamp to the current time
func (cs *CourtSlot) MarkAsNotified() {
	now := time.Now()
	cs.NotifiedAt = &now
	cs.UpdatedAt = now
}

// CourtSlotFilter represents filtering options for court slots
type CourtSlotFilter struct {
	VenueID     *primitive.ObjectID `json:"venue_id,omitempty" bson:"venue_id,omitempty"`
	Date        *string             `json:"date,omitempty" bson:"date,omitempty"`               // Format: "YYYY-MM-DD"
	SlotDateGTE *time.Time          `json:"slot_date_gte,omitempty" bson:"slot_date,omitempty"` // Greater than or equal to
	SlotDateLTE *time.Time          `json:"slot_date_lte,omitempty" bson:"slot_date,omitempty"` // Less than or equal to
	StartTime   *string             `json:"start_time,omitempty" bson:"start_time,omitempty"`   // Format: "HH:MM"
	EndTime     *string             `json:"end_time,omitempty" bson:"end_time,omitempty"`       // Format: "HH:MM"
	Available   *bool               `json:"available,omitempty" bson:"available,omitempty"`
	Provider    *string             `json:"provider,omitempty" bson:"provider,omitempty"`
	MinPrice    *float64            `json:"min_price,omitempty" bson:"price,omitempty"`
	MaxPrice    *float64            `json:"max_price,omitempty" bson:"price,omitempty"`
	Notified    *bool               `json:"notified,omitempty"` // Filter by notification status
}

// CourtSlotService provides methods for interacting with court slots
type CourtSlotService struct {
	collection *mongo.Collection
}

// NewCourtSlotService creates a new court slot service
func NewCourtSlotService(db *mongo.Database) *CourtSlotService {
	return &CourtSlotService{
		collection: db.Collection("court_slots"),
	}
}

// CreateIndexes creates the necessary indexes for efficient retention queries
func (s *CourtSlotService) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "slot_date", Value: 1},
			},
			Options: options.Index().SetName("slot_date_1"),
		},
		{
			Keys: bson.D{
				{Key: "notified_at", Value: 1},
			},
			Options: options.Index().SetName("notified_at_1").SetSparse(true), // Sparse index since notified_at can be null
		},
		{
			Keys: bson.D{
				{Key: "slot_date", Value: 1},
				{Key: "notified_at", Value: 1},
			},
			Options: options.Index().SetName("slot_date_notified_at_1"),
		},
		{
			Keys: bson.D{
				{Key: "venue_id", Value: 1},
				{Key: "slot_date", Value: 1},
			},
			Options: options.Index().SetName("venue_id_slot_date_1"),
		},
		{
			Keys: bson.D{
				{Key: "created_at", Value: 1},
			},
			Options: options.Index().SetName("created_at_1"),
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// FindOldUnnotifiedSlots finds court slots that are older than the specified duration and have not been notified
func (s *CourtSlotService) FindOldUnnotifiedSlots(ctx context.Context, olderThan time.Duration) ([]CourtSlot, error) {
	cutoffTime := time.Now().Add(-olderThan)

	filter := bson.M{
		"slot_date":   bson.M{"$lt": cutoffTime},
		"notified_at": bson.M{"$exists": false},
	}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var slots []CourtSlot
	if err = cursor.All(ctx, &slots); err != nil {
		return nil, err
	}

	return slots, nil
}

// DeleteSlotsByIDs deletes court slots by their IDs
func (s *CourtSlotService) DeleteSlotsByIDs(ctx context.Context, slotIDs []string) (int64, error) {
	if len(slotIDs) == 0 {
		return 0, nil
	}

	// Convert string IDs to ObjectIDs if needed, or use as-is if they're strings
	var ids []interface{}
	for _, id := range slotIDs {
		ids = append(ids, id)
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}
	result, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// MarkSlotAsNotified updates a slot to mark it as notified
func (s *CourtSlotService) MarkSlotAsNotified(ctx context.Context, slotID string) error {
	now := time.Now()
	filter := bson.M{"_id": slotID}
	update := bson.M{
		"$set": bson.M{
			"notified_at": now,
			"updated_at":  now,
		},
	}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// Collection returns the name of the MongoDB collection for court slots
func (s *CourtSlotService) Collection() string {
	return "court_slots"
}
