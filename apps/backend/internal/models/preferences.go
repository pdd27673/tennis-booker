package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserPreferences represents user preferences for tennis court booking
type UserPreferences struct {
	ID                   primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	UserID               primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Times                []TimeRange          `bson:"times,omitempty" json:"times,omitempty"`
	MaxPrice             float64              `bson:"max_price,omitempty" json:"max_price,omitempty"`
	PreferredVenues      []string             `bson:"preferred_venues,omitempty" json:"preferred_venues,omitempty"`
	ExcludedVenues       []string             `bson:"excluded_venues,omitempty" json:"excluded_venues,omitempty"`
	PreferredDays        []string             `bson:"preferred_days,omitempty" json:"preferred_days,omitempty"` // "monday", "tuesday", etc.
	NotificationSettings NotificationSettings `bson:"notification_settings,omitempty" json:"notification_settings,omitempty"`
	CreatedAt            time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updated_at"`
}

// NotificationSettings represents notification preferences for court availability alerts
type NotificationSettings struct {
	Email                bool   `bson:"email" json:"email"`
	EmailAddress         string `bson:"email_address,omitempty" json:"email_address,omitempty"`
	InstantAlerts        bool   `bson:"instant_alerts" json:"instant_alerts"`                                       // Receive alerts immediately when courts become available
	MaxAlertsPerHour     int    `bson:"max_alerts_per_hour,omitempty" json:"max_alerts_per_hour,omitempty"`         // Rate limiting (default: 10)
	MaxAlertsPerDay      int    `bson:"max_alerts_per_day,omitempty" json:"max_alerts_per_day,omitempty"`           // Daily limit (default: 50)
	AlertTimeWindowStart string `bson:"alert_time_window_start,omitempty" json:"alert_time_window_start,omitempty"` // e.g., "07:00" - when to start sending alerts
	AlertTimeWindowEnd   string `bson:"alert_time_window_end,omitempty" json:"alert_time_window_end,omitempty"`     // e.g., "22:00" - when to stop sending alerts
	Unsubscribed         bool   `bson:"unsubscribed,omitempty" json:"unsubscribed,omitempty"`                       // User has unsubscribed from all alerts
}

// PreferenceRequest represents the request payload for updating preferences
type PreferenceRequest struct {
	Times                []TimeRange           `json:"times,omitempty" binding:"dive"`
	MaxPrice             *float64              `json:"max_price,omitempty" binding:"omitempty,gte=0"`
	PreferredVenues      []string              `json:"preferred_venues,omitempty"`
	ExcludedVenues       []string              `json:"excluded_venues,omitempty"`
	PreferredDays        []string              `json:"preferred_days,omitempty" binding:"dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	NotificationSettings *NotificationSettings `json:"notification_settings,omitempty"`
}

// AddVenueRequest represents the request payload for adding a venue to preferences
type AddVenueRequest struct {
	VenueID   string `json:"venue_id" binding:"required"`
	VenueType string `json:"venue_type,omitempty" binding:"omitempty,oneof=preferred excluded"`
}

// PreferenceService provides methods for interacting with user preferences
type PreferenceService struct {
	collection *mongo.Collection
}

// NewPreferenceService creates a new preference service
func NewPreferenceService(db *mongo.Database) *PreferenceService {
	return &PreferenceService{
		collection: db.Collection("user_preferences"),
	}
}

// GetUserPreferences retrieves preferences for a specific user
func (s *PreferenceService) GetUserPreferences(ctx context.Context, userID primitive.ObjectID) (*UserPreferences, error) {
	var preferences UserPreferences

	filter := bson.M{"user_id": userID}
	err := s.collection.FindOne(ctx, filter).Decode(&preferences)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Return default preferences if none exist
			return &UserPreferences{
				UserID:          userID,
				Times:           []TimeRange{},
				MaxPrice:        0,
				PreferredVenues: []string{},
				ExcludedVenues:  []string{},
				PreferredDays:   []string{},
				NotificationSettings: NotificationSettings{
					Email:                true,
					InstantAlerts:        true,
					MaxAlertsPerHour:     10,
					MaxAlertsPerDay:      50,
					AlertTimeWindowStart: "07:00",
					AlertTimeWindowEnd:   "22:00",
					Unsubscribed:         false,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, err
	}

	return &preferences, nil
}

// UpdateUserPreferences updates or creates user preferences
func (s *PreferenceService) UpdateUserPreferences(ctx context.Context, userID primitive.ObjectID, req *PreferenceRequest) (*UserPreferences, error) {
	now := time.Now()

	// Build update document
	updateDoc := bson.M{
		"$set": bson.M{
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"user_id":    userID,
			"created_at": now,
		},
	}

	// Add fields that were provided in the request
	if req.Times != nil {
		updateDoc["$set"].(bson.M)["times"] = req.Times
	}
	if req.MaxPrice != nil {
		updateDoc["$set"].(bson.M)["max_price"] = *req.MaxPrice
	}
	if req.PreferredVenues != nil {
		updateDoc["$set"].(bson.M)["preferred_venues"] = req.PreferredVenues
	}
	if req.ExcludedVenues != nil {
		updateDoc["$set"].(bson.M)["excluded_venues"] = req.ExcludedVenues
	}
	if req.PreferredDays != nil {
		updateDoc["$set"].(bson.M)["preferred_days"] = req.PreferredDays
	}
	if req.NotificationSettings != nil {
		updateDoc["$set"].(bson.M)["notification_settings"] = *req.NotificationSettings
	}

	// Upsert the document
	filter := bson.M{"user_id": userID}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result UserPreferences
	err := s.collection.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AddVenueToPreferredList adds a venue to the user's preferred venues list
func (s *PreferenceService) AddVenueToPreferredList(ctx context.Context, userID primitive.ObjectID, venueID string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$addToSet": bson.M{
			"preferred_venues": venueID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"user_id":         userID,
			"created_at":      time.Now(),
			"times":           []TimeRange{},
			"max_price":       0,
			"excluded_venues": []string{},
			"preferred_days":  []string{},
			"notification_settings": NotificationSettings{
				Email:                true,
				InstantAlerts:        true,
				MaxAlertsPerHour:     10,
				MaxAlertsPerDay:      50,
				AlertTimeWindowStart: "07:00",
				AlertTimeWindowEnd:   "22:00",
				Unsubscribed:         false,
			},
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// AddVenueToExcludedList adds a venue to the user's excluded venues list
func (s *PreferenceService) AddVenueToExcludedList(ctx context.Context, userID primitive.ObjectID, venueID string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$addToSet": bson.M{
			"excluded_venues": venueID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"user_id":          userID,
			"created_at":       time.Now(),
			"times":            []TimeRange{},
			"max_price":        0,
			"preferred_venues": []string{},
			"preferred_days":   []string{},
			"notification_settings": NotificationSettings{
				Email:                true,
				InstantAlerts:        true,
				MaxAlertsPerHour:     10,
				MaxAlertsPerDay:      50,
				AlertTimeWindowStart: "07:00",
				AlertTimeWindowEnd:   "22:00",
				Unsubscribed:         false,
			},
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// RemoveVenueFromPreferredList removes a venue from the user's preferred venues list
func (s *PreferenceService) RemoveVenueFromPreferredList(ctx context.Context, userID primitive.ObjectID, venueID string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$pull": bson.M{
			"preferred_venues": venueID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// RemoveVenueFromExcludedList removes a venue from the user's excluded venues list
func (s *PreferenceService) RemoveVenueFromExcludedList(ctx context.Context, userID primitive.ObjectID, venueID string) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$pull": bson.M{
			"excluded_venues": venueID,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteUserPreferences removes all preferences for a user
func (s *PreferenceService) DeleteUserPreferences(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	_, err := s.collection.DeleteOne(ctx, filter)
	return err
}

// Collection returns the name of the MongoDB collection for user preferences
func (s *PreferenceService) Collection() string {
	return "user_preferences"
}
