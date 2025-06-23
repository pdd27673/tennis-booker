package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/database"
	"tennis-booker/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserHandler handles user-related requests
type UserHandler struct {
	db         database.Database
	jwtService *auth.JWTService
}

// NewUserHandler creates a new user handler
func NewUserHandler(db database.Database, jwtService *auth.JWTService) *UserHandler {
	return &UserHandler{
		db:         db,
		jwtService: jwtService,
	}
}

// UserPreferencesResponse represents user preferences for API responses
type UserPreferencesResponse struct {
	ID                   string                      `json:"id"`
	UserID               string                      `json:"userId"`
	Times                []models.TimeRange          `json:"times"`        // Legacy field for backward compatibility
	WeekdayTimes         []models.TimeRange          `json:"weekdayTimes"` // Monday-Friday preferred times
	WeekendTimes         []models.TimeRange          `json:"weekendTimes"` // Saturday-Sunday preferred times
	PreferredVenues      []string                    `json:"preferredVenues"`
	ExcludedVenues       []string                    `json:"excludedVenues"`
	PreferredDays        []string                    `json:"preferredDays"`
	MaxPrice             float64                     `json:"maxPrice"`
	NotificationSettings models.NotificationSettings `json:"notificationSettings"`
	CreatedAt            time.Time                   `json:"createdAt"`
	UpdatedAt            time.Time                   `json:"updatedAt"`
}

// UpdatePreferencesRequest represents a request to update user preferences
type UpdatePreferencesRequest struct {
	Times                []models.TimeRange           `json:"times"`        // Legacy field for backward compatibility
	WeekdayTimes         []models.TimeRange           `json:"weekdayTimes"` // Monday-Friday preferred times
	WeekendTimes         []models.TimeRange           `json:"weekendTimes"` // Saturday-Sunday preferred times
	PreferredVenues      []string                     `json:"preferredVenues"`
	ExcludedVenues       []string                     `json:"excludedVenues"`
	PreferredDays        []string                     `json:"preferredDays"`
	MaxPrice             float64                      `json:"maxPrice"`
	NotificationSettings *models.NotificationSettings `json:"notificationSettings"`
}

// GetPreferences handles GET /api/users/preferences
func (h *UserHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("user_preferences")
	var preferences models.UserPreferences
	err = collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&preferences)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create default preferences if none exist
			preferences = models.UserPreferences{
				ID:              primitive.NewObjectID(),
				UserID:          userID,
				Times:           []models.TimeRange{},
				PreferredVenues: []string{},
				ExcludedVenues:  []string{},
				PreferredDays:   []string{},
				MaxPrice:        100.0,
				NotificationSettings: models.NotificationSettings{
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
			}

			// Insert default preferences
			_, err = collection.InsertOne(ctx, preferences)
			if err != nil {
				http.Error(w, "Failed to create default preferences", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Failed to fetch preferences", http.StatusInternalServerError)
			return
		}
	}

	// Convert to response format
	response := UserPreferencesResponse{
		ID:                   preferences.ID.Hex(),
		UserID:               preferences.UserID.Hex(),
		Times:                preferences.Times,
		WeekdayTimes:         preferences.WeekdayTimes,
		WeekendTimes:         preferences.WeekendTimes,
		PreferredVenues:      preferences.PreferredVenues,
		ExcludedVenues:       preferences.ExcludedVenues,
		PreferredDays:        preferences.PreferredDays,
		MaxPrice:             preferences.MaxPrice,
		NotificationSettings: preferences.NotificationSettings,
		CreatedAt:            preferences.CreatedAt,
		UpdatedAt:            preferences.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdatePreferences handles PUT /api/users/preferences
func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("user_preferences")

	// Check if preferences exist
	var existingPreferences models.UserPreferences
	err = collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&existingPreferences)

	if err != nil && err != mongo.ErrNoDocuments {
		http.Error(w, "Failed to check existing preferences", http.StatusInternalServerError)
		return
	}

	if err == mongo.ErrNoDocuments {
		// Create new preferences
		preferences := models.UserPreferences{
			ID:              primitive.NewObjectID(),
			UserID:          userID,
			Times:           req.Times,
			WeekdayTimes:    req.WeekdayTimes,
			WeekendTimes:    req.WeekendTimes,
			PreferredVenues: req.PreferredVenues,
			ExcludedVenues:  req.ExcludedVenues,
			PreferredDays:   req.PreferredDays,
			MaxPrice:        req.MaxPrice,
			NotificationSettings: func() models.NotificationSettings {
				if req.NotificationSettings != nil {
					return *req.NotificationSettings
				}
				return models.NotificationSettings{
					Email:                true,
					InstantAlerts:        true,
					MaxAlertsPerHour:     10,
					MaxAlertsPerDay:      50,
					AlertTimeWindowStart: "07:00",
					AlertTimeWindowEnd:   "22:00",
					Unsubscribed:         false,
				}
			}(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = collection.InsertOne(ctx, preferences)
		if err != nil {
			http.Error(w, "Failed to create preferences", http.StatusInternalServerError)
			return
		}

		// Return created preferences
		response := UserPreferencesResponse{
			ID:                   preferences.ID.Hex(),
			UserID:               preferences.UserID.Hex(),
			Times:                preferences.Times,
			WeekdayTimes:         preferences.WeekdayTimes,
			WeekendTimes:         preferences.WeekendTimes,
			PreferredVenues:      preferences.PreferredVenues,
			ExcludedVenues:       preferences.ExcludedVenues,
			PreferredDays:        preferences.PreferredDays,
			MaxPrice:             preferences.MaxPrice,
			NotificationSettings: preferences.NotificationSettings,
			CreatedAt:            preferences.CreatedAt,
			UpdatedAt:            preferences.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update existing preferences
	updateFields := bson.M{
		"updatedAt": time.Now(),
	}

	// Only update fields that are provided
	if req.Times != nil {
		updateFields["times"] = req.Times
	}
	if req.WeekdayTimes != nil {
		updateFields["weekday_times"] = req.WeekdayTimes
	}
	if req.WeekendTimes != nil {
		updateFields["weekend_times"] = req.WeekendTimes
	}
	if req.PreferredVenues != nil {
		updateFields["preferred_venues"] = req.PreferredVenues
	}
	if req.ExcludedVenues != nil {
		updateFields["excluded_venues"] = req.ExcludedVenues
	}
	if req.PreferredDays != nil {
		updateFields["preferred_days"] = req.PreferredDays
	}
	updateFields["max_price"] = req.MaxPrice
	if req.NotificationSettings != nil {
		updateFields["notification_settings"] = *req.NotificationSettings
	}

	update := bson.M{"$set": updateFields}

	result, err := collection.UpdateOne(ctx, bson.M{"userId": userID}, update)
	if err != nil {
		http.Error(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Preferences not found", http.StatusNotFound)
		return
	}

	// Fetch updated preferences
	var updatedPreferences models.UserPreferences
	err = collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&updatedPreferences)
	if err != nil {
		http.Error(w, "Failed to fetch updated preferences", http.StatusInternalServerError)
		return
	}

	// Return updated preferences
	response := UserPreferencesResponse{
		ID:                   updatedPreferences.ID.Hex(),
		UserID:               updatedPreferences.UserID.Hex(),
		Times:                updatedPreferences.Times,
		WeekdayTimes:         updatedPreferences.WeekdayTimes,
		WeekendTimes:         updatedPreferences.WeekendTimes,
		PreferredVenues:      updatedPreferences.PreferredVenues,
		ExcludedVenues:       updatedPreferences.ExcludedVenues,
		PreferredDays:        updatedPreferences.PreferredDays,
		MaxPrice:             updatedPreferences.MaxPrice,
		NotificationSettings: updatedPreferences.NotificationSettings,
		CreatedAt:            updatedPreferences.CreatedAt,
		UpdatedAt:            updatedPreferences.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// validatePreferences validates the preferences data
func (h *UserHandler) validatePreferences(prefs *models.UserPreferences) error {
	// Validate preferred days
	validDays := map[string]bool{
		"monday": true, "tuesday": true, "wednesday": true, "thursday": true,
		"friday": true, "saturday": true, "sunday": true,
	}
	for _, day := range prefs.PreferredDays {
		if !validDays[day] {
			return &ValidationError{Field: "preferred_days", Message: "invalid day: " + day}
		}
	}

	// Validate time ranges
	for i, timeRange := range prefs.Times {
		if err := h.validateTimeRange(&timeRange); err != nil {
			return &ValidationError{
				Field:   "preferred_times",
				Message: "invalid time range at index " + fmt.Sprintf("%d", i) + ": " + err.Error(),
			}
		}
	}

	return nil
}

// validateTimeRange validates a time range
func (h *UserHandler) validateTimeRange(tr *models.TimeRange) error {
	// Basic format validation (HH:MM)
	if len(tr.Start) != 5 || tr.Start[2] != ':' {
		return &ValidationError{Field: "start", Message: "invalid time format, expected HH:MM"}
	}
	if len(tr.End) != 5 || tr.End[2] != ':' {
		return &ValidationError{Field: "end", Message: "invalid time format, expected HH:MM"}
	}

	// Parse and validate time values
	startTime, err := time.Parse("15:04", tr.Start)
	if err != nil {
		return &ValidationError{Field: "start", Message: "invalid time value"}
	}

	endTime, err := time.Parse("15:04", tr.End)
	if err != nil {
		return &ValidationError{Field: "end", Message: "invalid time value"}
	}

	// Ensure start time is before end time
	if !startTime.Before(endTime) {
		return &ValidationError{Field: "time_range", Message: "start time must be before end time"}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// writeErrorResponse writes an error response in JSON format
func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	json.NewEncoder(w).Encode(errorResp)
}
