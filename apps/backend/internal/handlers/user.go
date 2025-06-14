package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/models"
)

// UserPreferences represents the user preferences that can be updated
type UserPreferences struct {
	PreferredCourts []string           `json:"preferred_courts,omitempty"`
	PreferredDays   []string           `json:"preferred_days,omitempty"`
	PreferredTimes  []models.TimeRange `json:"preferred_times,omitempty"`
	NotifyBy        []string           `json:"notify_by,omitempty"`
}

// UserHandler handles user-related endpoints
type UserHandler struct {
	userService models.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService models.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UpdatePreferences handles PUT /api/users/preferences - updates user preferences
func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	// Only allow PUT method
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", "PUT")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user claims from context (set by JWT middleware)
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var preferences UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&preferences); err != nil {
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate preferences
	if err := h.validatePreferences(&preferences); err != nil {
		h.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get current user
	user, err := h.userService.FindByID(r.Context(), claims.UserID)
	if err != nil {
		h.writeErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Update user preferences
	user.PreferredCourts = preferences.PreferredCourts
	user.PreferredDays = preferences.PreferredDays
	user.PreferredTimes = preferences.PreferredTimes
	user.NotifyBy = preferences.NotifyBy
	user.UpdatedAt = time.Now()

	// Save updated user
	if err := h.userService.UpdateUser(r.Context(), user); err != nil {
		h.writeErrorResponse(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	// Return updated preferences
	response := UserPreferences{
		PreferredCourts: user.PreferredCourts,
		PreferredDays:   user.PreferredDays,
		PreferredTimes:  user.PreferredTimes,
		NotifyBy:        user.NotifyBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// validatePreferences validates the preferences data
func (h *UserHandler) validatePreferences(prefs *UserPreferences) error {
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

	// Validate notification methods
	validNotifyMethods := map[string]bool{
		"email": true, "sms": true,
	}
	for _, method := range prefs.NotifyBy {
		if !validNotifyMethods[method] {
			return &ValidationError{Field: "notify_by", Message: "invalid notification method: " + method}
		}
	}

	// Validate time ranges
	for i, timeRange := range prefs.PreferredTimes {
		if err := h.validateTimeRange(&timeRange); err != nil {
			return &ValidationError{
				Field:   "preferred_times",
				Message: "invalid time range at index " + string(rune(i)) + ": " + err.Error(),
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

	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResp)
}
