package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AlertHistory represents a notification alert sent to a user
type AlertHistory struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	VenueID       string             `bson:"venue_id" json:"venue_id"`
	VenueName     string             `bson:"venue_name" json:"venue_name"`
	CourtID       string             `bson:"court_id" json:"court_id"`
	CourtName     string             `bson:"court_name" json:"court_name"`
	SlotDate      string             `bson:"slot_date" json:"slot_date"`             // YYYY-MM-DD
	SlotStartTime string             `bson:"slot_start_time" json:"slot_start_time"` // HH:MM
	SlotEndTime   string             `bson:"slot_end_time" json:"slot_end_time"`     // HH:MM
	Price         float64            `bson:"price" json:"price"`
	Currency      string             `bson:"currency" json:"currency"`
	BookingURL    string             `bson:"booking_url" json:"booking_url"`
	EmailAddress  string             `bson:"email_address" json:"email_address"`
	AlertSentAt   time.Time          `bson:"alert_sent_at" json:"alert_sent_at"`
	EmailStatus   string             `bson:"email_status" json:"email_status"` // sent, delivered, failed, bounced
	SlotKey       string             `bson:"slot_key" json:"slot_key"`         // Unique key for deduplication
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// CourtAvailabilityEvent represents a court availability event from Redis
type CourtAvailabilityEvent struct {
	VenueID      string    `json:"venue_id"`
	VenueName    string    `json:"venue_name"`
	CourtID      string    `json:"court_id"`
	CourtName    string    `json:"court_name"`
	Date         string    `json:"date"`       // Format: "YYYY-MM-DD"
	StartTime    string    `json:"start_time"` // Format: "HH:MM"
	EndTime      string    `json:"end_time"`   // Format: "HH:MM"
	Price        float64   `json:"price"`
	Currency     string    `json:"currency"`
	BookingURL   string    `json:"booking_url"`
	DiscoveredAt time.Time `json:"discovered_at"`
	ScrapeLogID  string    `json:"scrape_log_id"`
}

// GenerateSlotKey creates a unique identifier for a court slot
func (e *CourtAvailabilityEvent) GenerateSlotKey() string {
	return e.VenueID + ":" + e.CourtID + ":" + e.Date + ":" + e.StartTime
}

// AlertHistoryService provides methods for managing notification alert history
type AlertHistoryService struct {
	collection *mongo.Collection
}

// NewAlertHistoryService creates a new alert history service
func NewAlertHistoryService(db *mongo.Database) *AlertHistoryService {
	return &AlertHistoryService{
		collection: db.Collection("alert_history"),
	}
}

// CreateAlert creates a new alert history record
func (s *AlertHistoryService) CreateAlert(ctx context.Context, alert *AlertHistory) error {
	alert.CreatedAt = time.Now()
	alert.AlertSentAt = time.Now()

	_, err := s.collection.InsertOne(ctx, alert)
	return err
}

// HasRecentAlert checks if user was recently alerted about the same slot
func (s *AlertHistoryService) HasRecentAlert(ctx context.Context, userID primitive.ObjectID, slotKey string, withinHours int) (bool, error) {
	since := time.Now().Add(time.Duration(-withinHours) * time.Hour)

	filter := bson.M{
		"user_id":       userID,
		"slot_key":      slotKey,
		"alert_sent_at": bson.M{"$gte": since},
	}

	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserAlertCount returns the number of alerts sent to a user in the last period
func (s *AlertHistoryService) GetUserAlertCount(ctx context.Context, userID primitive.ObjectID, withinHours int) (int64, error) {
	since := time.Now().Add(time.Duration(-withinHours) * time.Hour)

	filter := bson.M{
		"user_id":       userID,
		"alert_sent_at": bson.M{"$gte": since},
	}

	return s.collection.CountDocuments(ctx, filter)
}

// UpdateEmailStatus updates the delivery status of an alert
func (s *AlertHistoryService) UpdateEmailStatus(ctx context.Context, alertID primitive.ObjectID, status string) error {
	filter := bson.M{"_id": alertID}
	update := bson.M{
		"$set": bson.M{
			"email_status": status,
		},
	}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// CleanupOldAlerts removes alert history older than the specified number of days
func (s *AlertHistoryService) CleanupOldAlerts(ctx context.Context, olderThanDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	filter := bson.M{
		"created_at": bson.M{"$lt": cutoff},
	}

	result, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// GetUserAlertHistory retrieves recent alert history for a user
func (s *AlertHistoryService) GetUserAlertHistory(ctx context.Context, userID primitive.ObjectID, limit int) ([]AlertHistory, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.M{"alert_sent_at": -1}).
		SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []AlertHistory
	if err = cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}

	return alerts, nil
}

// Collection returns the MongoDB collection name
func (s *AlertHistoryService) Collection() string {
	return "alert_history"
}
