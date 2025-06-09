package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tennis-booking-bot/internal/models"
)

// VenueHandlers handles venue-related endpoints
type VenueHandlers struct {
	db *mongo.Database
}

// NewVenueHandlers creates a new venue handlers instance
func NewVenueHandlers(db *mongo.Database) *VenueHandlers {
	return &VenueHandlers{
		db: db,
	}
}

// VenueListResponse represents the response for venue list
type VenueListResponse struct {
	Data   []models.Venue `json:"data"`
	Count  int            `json:"count"`
	Status string         `json:"status"`
}

// VenueSlotsResponse represents the response for venue slots
type VenueSlotsResponse struct {
	Data       []VenueSlotData `json:"data"`
	Count      int             `json:"count"`
	VenueID    string          `json:"venue_id"`
	VenueName  string          `json:"venue_name"`
	Provider   string          `json:"provider"`
	LastUpdate time.Time       `json:"last_update"`
	Status     string          `json:"status"`
}

// VenueSlotData represents enriched slot data with venue context
type VenueSlotData struct {
	ID          primitive.ObjectID `json:"scraping_log_id"`
	Date        string             `json:"date"`
	Time        string             `json:"time"`
	Court       string             `json:"court"`
	CourtID     string             `json:"court_id"`
	Price       float64            `json:"price"`
	Available   bool               `json:"available"`
	BookingURL  string             `json:"booking_url"`
	ScrapedAt   time.Time          `json:"scraped_at"`
	VenueName   string             `json:"venue_name"`
	Provider    string             `json:"provider"`
	TimeRange   TimeRangeInfo      `json:"time_range"`
}

// TimeRangeInfo provides parsed time information
type TimeRangeInfo struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Duration  string `json:"duration"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// GetVenues handles GET /api/v1/venues
func (h *VenueHandlers) GetVenues(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("venues")

	// Build filter - only active venues by default
	filter := bson.M{"is_active": true}

	// Add optional provider filter
	if provider := c.Query("provider"); provider != "" {
		filter["provider"] = provider
	}

	// Find venues
	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve venues",
			Status:  "error",
		})
		return
	}
	defer cursor.Close(ctx)

	var venues []models.Venue
	if err = cursor.All(ctx, &venues); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "decode_error",
			Message: "Failed to decode venues",
			Status:  "error",
		})
		return
	}

	// Ensure we return an empty array instead of null
	if venues == nil {
		venues = []models.Venue{}
	}

	c.JSON(http.StatusOK, VenueListResponse{
		Data:   venues,
		Count:  len(venues),
		Status: "success",
	})
}

// GetVenueSlots handles GET /api/v1/venues/{id}/slots
func (h *VenueHandlers) GetVenueSlots(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse venue ID
	venueIDStr := c.Param("id")
	venueID, err := primitive.ObjectIDFromHex(venueIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_venue_id",
			Message: "Invalid venue ID format",
			Status:  "error",
		})
		return
	}

	// Verify venue exists
	venueCollection := h.db.Collection("venues")
	var venue models.Venue
	err = venueCollection.FindOne(ctx, bson.M{"_id": venueID}).Decode(&venue)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "venue_not_found",
				Message: "Venue not found",
				Status:  "error",
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "Failed to retrieve venue",
				Status:  "error",
			})
		}
		return
	}

	// Build filter for scraping logs
	logFilter := bson.M{
		"venue_id": venueID,
		"success":  true,
	}

	// Filter by date range if provided
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if dateTo := c.Query("date_to"); dateTo != "" {
			logFilter["slots_found.date"] = bson.M{
				"$gte": dateFrom,
				"$lte": dateTo,
			}
		} else {
			logFilter["slots_found.date"] = bson.M{"$gte": dateFrom}
		}
	}

	// Filter for recent scrapes (last 24 hours by default)
	hoursBack := 24
	if h := c.Query("hours_back"); h != "" {
		// Could parse custom hours, but keep it simple for now
	}
	since := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	logFilter["scrape_timestamp"] = bson.M{"$gte": since}

	// Query scraping logs
	logsCollection := h.db.Collection("scraping_logs")
	
	// Sort by most recent first
	findOptions := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}).
		SetLimit(1000) // Reasonable limit

	cursor, err := logsCollection.Find(ctx, logFilter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve slot data",
			Status:  "error",
		})
		return
	}
	defer cursor.Close(ctx)

	var scrapingLogs []models.ScrapingLog
	if err = cursor.All(ctx, &scrapingLogs); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "decode_error",
			Message: "Failed to decode slot data",
			Status:  "error",
		})
		return
	}

	// Transform to slot data with enrichment
	var slotData []VenueSlotData
	var lastUpdate time.Time

	for _, log := range scrapingLogs {
		if log.ScrapeTimestamp.After(lastUpdate) {
			lastUpdate = log.ScrapeTimestamp
		}

		for _, slot := range log.SlotsFound {
			// Only include available slots
			if !slot.Available {
				continue
			}

			timeRange := parseTimeRange(slot.Time)
			
			slotData = append(slotData, VenueSlotData{
				ID:         log.ID,
				Date:       slot.Date,
				Time:       slot.Time,
				Court:      slot.Court,
				CourtID:    slot.CourtID,
				Price:      slot.Price,
				Available:  slot.Available,
				BookingURL: slot.URL,
				ScrapedAt:  log.ScrapeTimestamp,
				VenueName:  log.VenueName,
				Provider:   log.Provider,
				TimeRange:  timeRange,
			})
		}
	}

	// Ensure we return an empty array instead of null
	if slotData == nil {
		slotData = []VenueSlotData{}
	}

	c.JSON(http.StatusOK, VenueSlotsResponse{
		Data:       slotData,
		Count:      len(slotData),
		VenueID:    venueIDStr,
		VenueName:  venue.Name,
		Provider:   venue.Provider,
		LastUpdate: lastUpdate,
		Status:     "success",
	})
}

// parseTimeRange parses time range like "09:00-10:00" into structured info
func parseTimeRange(timeStr string) TimeRangeInfo {
	// Simple parsing for now - could be enhanced
	if len(timeStr) == 0 {
		return TimeRangeInfo{
			StartTime: "",
			EndTime:   "",
			Duration:  "",
		}
	}

	// Split on dash
	parts := []string{timeStr, ""} // Default fallback
	if len(timeStr) >= 11 && timeStr[5] == '-' {
		parts = []string{timeStr[:5], timeStr[6:]}
	}

	return TimeRangeInfo{
		StartTime: parts[0],
		EndTime:   parts[1],
		Duration:  calculateDuration(parts[0], parts[1]),
	}
}

 