package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"tennis-booking-bot/internal/models"
)

// PreferenceHandlers handles all preference-related HTTP requests
type PreferenceHandlers struct {
	preferenceService *models.PreferenceService
}

// NewPreferenceHandlers creates a new preference handlers instance
func NewPreferenceHandlers(preferenceService *models.PreferenceService) *PreferenceHandlers {
	return &PreferenceHandlers{
		preferenceService: preferenceService,
	}
}

// GetPreferences handles GET /api/v1/preferences
// For now, we'll use userID from query params (placeholder for JWT implementation)
func (h *PreferenceHandlers) GetPreferences(c *gin.Context) {
	// Extract userID from query parameter (placeholder for JWT implementation)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id query parameter is required",
		})
		return
	}

	// Convert userID string to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id format",
		})
		return
	}

	// Get user preferences
	preferences, err := h.preferenceService.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve preferences",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": preferences,
	})
}

// UpdatePreferences handles PUT /api/v1/preferences
func (h *PreferenceHandlers) UpdatePreferences(c *gin.Context) {
	// Extract userID from query parameter (placeholder for JWT implementation)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id query parameter is required",
		})
		return
	}

	// Convert userID string to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id format",
		})
		return
	}

	// Bind and validate JSON request body
	var req models.PreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Validate time ranges format if provided
	if req.Times != nil {
		for i, timeRange := range req.Times {
			if err := validateTimeFormat(timeRange.Start); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid time format in times[" + strconv.Itoa(i) + "].start",
					"details": "time must be in HH:MM format (24-hour)",
				})
				return
			}
			if err := validateTimeFormat(timeRange.End); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid time format in times[" + strconv.Itoa(i) + "].end",
					"details": "time must be in HH:MM format (24-hour)",
				})
				return
			}
		}
	}

	// Update user preferences
	updatedPreferences, err := h.preferenceService.UpdateUserPreferences(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update preferences",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "preferences updated successfully",
		"data": updatedPreferences,
	})
}

// AddPreferredVenue handles POST /api/v1/preferences/venues
func (h *PreferenceHandlers) AddPreferredVenue(c *gin.Context) {
	// Extract userID from query parameter (placeholder for JWT implementation)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id query parameter is required",
		})
		return
	}

	// Convert userID string to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id format",
		})
		return
	}

	// Bind and validate JSON request body
	var req models.AddVenueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Determine which list to add the venue to based on venue_type
	var err2 error
	switch req.VenueType {
	case "excluded":
		err2 = h.preferenceService.AddVenueToExcludedList(c.Request.Context(), userID, req.VenueID)
	case "preferred", "": // Default to preferred if not specified
		err2 = h.preferenceService.AddVenueToPreferredList(c.Request.Context(), userID, req.VenueID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid venue_type, must be 'preferred' or 'excluded'",
		})
		return
	}

	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to add venue to preferences",
			"details": err2.Error(),
		})
		return
	}

	listType := req.VenueType
	if listType == "" {
		listType = "preferred"
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "venue added to " + listType + " list successfully",
		"venue_id": req.VenueID,
		"list_type": listType,
	})
}

// RemovePreferredVenue handles DELETE /api/v1/preferences/venues/:venueId
func (h *PreferenceHandlers) RemovePreferredVenue(c *gin.Context) {
	// Extract userID from query parameter (placeholder for JWT implementation)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id query parameter is required",
		})
		return
	}

	// Convert userID string to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id format",
		})
		return
	}

	// Extract venue ID from URL parameter
	venueID := c.Param("venueId")
	if venueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "venue_id parameter is required",
		})
		return
	}

	// Extract list type from query parameter (default to both)
	listType := c.Query("list_type") // "preferred", "excluded", or empty (both)

	var err2 error
	switch listType {
	case "preferred":
		err2 = h.preferenceService.RemoveVenueFromPreferredList(c.Request.Context(), userID, venueID)
	case "excluded":
		err2 = h.preferenceService.RemoveVenueFromExcludedList(c.Request.Context(), userID, venueID)
	case "": // Remove from both lists
		// Try to remove from preferred list first
		_ = h.preferenceService.RemoveVenueFromPreferredList(c.Request.Context(), userID, venueID)
		// Try to remove from excluded list
		err2 = h.preferenceService.RemoveVenueFromExcludedList(c.Request.Context(), userID, venueID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid list_type, must be 'preferred', 'excluded', or empty (both)",
		})
		return
	}

	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to remove venue from preferences",
			"details": err2.Error(),
		})
		return
	}

	message := "venue removed from preferences successfully"
	if listType != "" {
		message = "venue removed from " + listType + " list successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"venue_id": venueID,
	})
}

// validateTimeFormat validates that a time string is in HH:MM format (24-hour)
func validateTimeFormat(timeStr string) error {
	if len(timeStr) != 5 {
		return gin.Error{
			Err:  nil,
			Type: gin.ErrorTypeBind,
			Meta: "time must be in HH:MM format",
		}
	}

	if timeStr[2] != ':' {
		return gin.Error{
			Err:  nil,
			Type: gin.ErrorTypeBind,
			Meta: "time must contain ':' separator",
		}
	}

	// Parse hour
	hour := timeStr[0:2]
	hourInt, err := strconv.Atoi(hour)
	if err != nil || hourInt < 0 || hourInt > 23 {
		return gin.Error{
			Err:  nil,
			Type: gin.ErrorTypeBind,
			Meta: "hour must be between 00 and 23",
		}
	}

	// Parse minute
	minute := timeStr[3:5]
	minuteInt, err := strconv.Atoi(minute)
	if err != nil || minuteInt < 0 || minuteInt > 59 {
		return gin.Error{
			Err:  nil,
			Type: gin.ErrorTypeBind,
			Meta: "minute must be between 00 and 59",
		}
	}

	return nil
} 