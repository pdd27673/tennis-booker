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
	Times                []TimeRange          `bson:"times,omitempty" json:"times,omitempty"`                 // Legacy field for backward compatibility
	WeekdayTimes         []TimeRange          `bson:"weekday_times,omitempty" json:"weekday_times,omitempty"` // Monday-Friday preferred times
	WeekendTimes         []TimeRange          `bson:"weekend_times,omitempty" json:"weekend_times,omitempty"` // Saturday-Sunday preferred times
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
	Times                []TimeRange           `json:"times,omitempty" binding:"dive"`         // Legacy field for backward compatibility
	WeekdayTimes         []TimeRange           `json:"weekday_times,omitempty" binding:"dive"` // Monday-Friday preferred times
	WeekendTimes         []TimeRange           `json:"weekend_times,omitempty" binding:"dive"` // Saturday-Sunday preferred times
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
	if req.WeekdayTimes != nil {
		updateDoc["$set"].(bson.M)["weekday_times"] = req.WeekdayTimes
	}
	if req.WeekendTimes != nil {
		updateDoc["$set"].(bson.M)["weekend_times"] = req.WeekendTimes
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

// GetActiveUserPreferences retrieves all active user preferences for the retention service
// Active preferences are defined as preferences where:
// 1. User has not unsubscribed from notifications
// 2. User has at least one meaningful preference set (times, venues, days, or price limit)
func (s *PreferenceService) GetActiveUserPreferences(ctx context.Context) ([]UserPreferences, error) {
	// Build filter for active preferences
	filter := bson.M{
		"$and": []bson.M{
			// User has not unsubscribed
			{"notification_settings.unsubscribed": bson.M{"$ne": true}},
			// User has at least one meaningful preference
			{"$or": []bson.M{
				{"times": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"preferred_venues": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"excluded_venues": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"preferred_days": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"max_price": bson.M{"$gt": 0}},
			}},
		},
	}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var preferences []UserPreferences
	if err = cursor.All(ctx, &preferences); err != nil {
		return nil, err
	}

	return preferences, nil
}

// CreateIndexes creates the necessary indexes for efficient preference queries
func (s *PreferenceService) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
			Options: options.Index().SetName("user_id_1").SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "notification_settings.unsubscribed", Value: 1},
			},
			Options: options.Index().SetName("notification_unsubscribed_1").SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "preferred_venues", Value: 1},
			},
			Options: options.Index().SetName("preferred_venues_1").SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "preferred_days", Value: 1},
			},
			Options: options.Index().SetName("preferred_days_1").SetSparse(true),
		},
		{
			Keys: bson.D{
				{Key: "updated_at", Value: 1},
			},
			Options: options.Index().SetName("updated_at_1"),
		},
	}

	_, err := s.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// IsActivePreference checks if a user preference is considered active for retention purposes
func (s *PreferenceService) IsActivePreference(pref *UserPreferences) bool {
	// Check if user has unsubscribed
	if pref.NotificationSettings.Unsubscribed {
		return false
	}

	// Check if user has at least one meaningful preference
	return len(pref.Times) > 0 ||
		len(pref.PreferredVenues) > 0 ||
		len(pref.ExcludedVenues) > 0 ||
		len(pref.PreferredDays) > 0 ||
		pref.MaxPrice > 0
}

// GetActiveUserPreferencesCount returns the count of active user preferences
func (s *PreferenceService) GetActiveUserPreferencesCount(ctx context.Context) (int64, error) {
	filter := bson.M{
		"$and": []bson.M{
			{"notification_settings.unsubscribed": bson.M{"$ne": true}},
			{"$or": []bson.M{
				{"times": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"preferred_venues": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"excluded_venues": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"preferred_days": bson.M{"$exists": true, "$ne": []interface{}{}}},
				{"max_price": bson.M{"$gt": 0}},
			}},
		},
	}

	return s.collection.CountDocuments(ctx, filter)
}

// Collection returns the name of the MongoDB collection for user preferences
func (s *PreferenceService) Collection() string {
	return "user_preferences"
}
