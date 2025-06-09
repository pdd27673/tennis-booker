package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"tennis-booking-bot/internal/models"
)

// CourtsHandlers handles court availability endpoints
type CourtsHandlers struct {
	db *mongo.Database
}

// NewCourtsHandlers creates a new courts handlers instance
func NewCourtsHandlers(db *mongo.Database) *CourtsHandlers {
	return &CourtsHandlers{
		db: db,
	}
}

// CourtAvailabilityResponse represents the response for court availability
type CourtAvailabilityResponse struct {
	Data         []AvailableSlot `json:"data"`
	Count        int             `json:"count"`
	TotalSlots   int             `json:"total_slots"`
	LastUpdated  time.Time       `json:"last_updated"`
	Filters      AppliedFilters  `json:"filters"`
	VenueCounts  []VenueCount    `json:"venue_counts"`
	TimeRange    string          `json:"time_range"`
	Status       string          `json:"status"`
}

// AvailableSlot represents an available court slot with enriched data
type AvailableSlot struct {
	ID             primitive.ObjectID `json:"scraping_log_id"`
	VenueID        string             `json:"venue_id"`
	VenueName      string             `json:"venue_name"`
	Provider       string             `json:"provider"`
	CourtID        string             `json:"court_id"`
	CourtName      string             `json:"court_name"`
	Date           string             `json:"date"`
	StartTime      string             `json:"start_time"`
	EndTime        string             `json:"end_time"`
	TimeSlot       string             `json:"time_slot"`       // Combined display
	Duration       string             `json:"duration"`
	Price          float64            `json:"price"`
	FormattedPrice string             `json:"formatted_price"`
	Currency       string             `json:"currency"`
	BookingURL     string             `json:"booking_url"`
	ScrapedAt      time.Time          `json:"scraped_at"`
	DaysFromNow    int                `json:"days_from_now"`
	IsToday        bool               `json:"is_today"`
	IsTomorrow     bool               `json:"is_tomorrow"`
	IsWeekend      bool               `json:"is_weekend"`
	TimeOfDay      string             `json:"time_of_day"` // morning, afternoon, evening
	VenueLocation  VenueLocationInfo  `json:"venue_location"`
}

// VenueLocationInfo represents venue location details
type VenueLocationInfo struct {
	Address  string  `json:"address"`
	City     string  `json:"city"`
	PostCode string  `json:"post_code"`
	Latitude float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

// AppliedFilters shows what filters were applied
type AppliedFilters struct {
	VenueIDs     []string `json:"venue_ids,omitempty"`
	Provider     string   `json:"provider,omitempty"`
	DateFrom     string   `json:"date_from,omitempty"`
	DateTo       string   `json:"date_to,omitempty"`
	TimeFrom     string   `json:"time_from,omitempty"`
	TimeTo       string   `json:"time_to,omitempty"`
	PriceMin     float64  `json:"price_min,omitempty"`
	PriceMax     float64  `json:"price_max,omitempty"`
	DayOfWeek    []string `json:"day_of_week,omitempty"`
	HoursBack    int      `json:"hours_back"`
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
}

// VenueCount represents slot count per venue
type VenueCount struct {
	VenueID   string `json:"venue_id"`
	VenueName string `json:"venue_name"`
	Count     int    `json:"count"`
}

// GetAvailableCourts handles GET /api/v1/courts/available
func (h *CourtsHandlers) GetAvailableCourts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Parse filters
	filters := h.parseFilters(c)

	// Build MongoDB aggregation pipeline
	pipeline := h.buildAggregationPipeline(filters)

	// Execute aggregation
	collection := h.db.Collection("scraping_logs")
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve court availability",
			Status:  "error",
		})
		return
	}
	defer cursor.Close(ctx)

	// Process results
	var availableSlots []AvailableSlot
	var lastUpdated time.Time
	venueCounts := make(map[string]VenueCount)

	for cursor.Next(ctx) {
		var result struct {
			ScrapingLog models.ScrapingLog `bson:"scraping_log"`
			Venue       models.Venue       `bson:"venue"`
		}

		if err := cursor.Decode(&result); err != nil {
			continue
		}

		log := result.ScrapingLog
		venue := result.Venue

		// Track last updated time
		if log.ScrapeTimestamp.After(lastUpdated) {
			lastUpdated = log.ScrapeTimestamp
		}

		// Process each slot
		for _, slot := range log.SlotsFound {
			if !slot.Available {
				continue
			}

			// Apply additional filters
			if !h.matchesSlotFilters(slot, filters) {
				continue
			}

			// Create enriched slot data
			enrichedSlot := h.createEnrichedSlot(slot, log, venue)
			availableSlots = append(availableSlots, enrichedSlot)

			// Update venue counts
			key := venue.ID.Hex()
			if count, exists := venueCounts[key]; exists {
				count.Count++
				venueCounts[key] = count
			} else {
				venueCounts[key] = VenueCount{
					VenueID:   key,
					VenueName: venue.Name,
					Count:     1,
				}
			}
		}
	}

	// Apply pagination
	totalSlots := len(availableSlots)
	start := filters.Offset
	end := start + filters.Limit
	if start > len(availableSlots) {
		start = len(availableSlots)
	}
	if end > len(availableSlots) {
		end = len(availableSlots)
	}

	paginatedSlots := availableSlots[start:end]

	// Convert venue counts to slice
	var venueCountsSlice []VenueCount
	for _, count := range venueCounts {
		venueCountsSlice = append(venueCountsSlice, count)
	}

	// Generate time range description
	timeRange := h.generateTimeRangeDescription(filters)

	c.JSON(http.StatusOK, CourtAvailabilityResponse{
		Data:         paginatedSlots,
		Count:        len(paginatedSlots),
		TotalSlots:   totalSlots,
		LastUpdated:  lastUpdated,
		Filters:      filters,
		VenueCounts:  venueCountsSlice,
		TimeRange:    timeRange,
		Status:       "success",
	})
}

// parseFilters extracts and validates query parameters
func (h *CourtsHandlers) parseFilters(c *gin.Context) AppliedFilters {
	filters := AppliedFilters{
		HoursBack: 24, // Default to last 24 hours
		Limit:     100, // Default limit
		Offset:    0,
	}

	// Venue filters
	if venueIDs := c.Query("venue_ids"); venueIDs != "" {
		filters.VenueIDs = strings.Split(venueIDs, ",")
	}
	if provider := c.Query("provider"); provider != "" {
		filters.Provider = provider
	}

	// Date filters
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters.DateFrom = dateFrom
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		filters.DateTo = dateTo
	}

	// Time filters
	if timeFrom := c.Query("time_from"); timeFrom != "" {
		filters.TimeFrom = timeFrom
	}
	if timeTo := c.Query("time_to"); timeTo != "" {
		filters.TimeTo = timeTo
	}

	// Price filters
	if priceMin := c.Query("price_min"); priceMin != "" {
		if p, err := strconv.ParseFloat(priceMin, 64); err == nil {
			filters.PriceMin = p
		}
	}
	if priceMax := c.Query("price_max"); priceMax != "" {
		if p, err := strconv.ParseFloat(priceMax, 64); err == nil {
			filters.PriceMax = p
		}
	}

	// Day of week filter
	if dayOfWeek := c.Query("day_of_week"); dayOfWeek != "" {
		filters.DayOfWeek = strings.Split(dayOfWeek, ",")
	}

	// Pagination
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			filters.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filters.Offset = o
		}
	}

	// Hours back filter
	if hoursBack := c.Query("hours_back"); hoursBack != "" {
		if h, err := strconv.Atoi(hoursBack); err == nil && h > 0 && h <= 168 { // Max 1 week
			filters.HoursBack = h
		}
	}

	return filters
}

// buildAggregationPipeline creates MongoDB aggregation pipeline
func (h *CourtsHandlers) buildAggregationPipeline(filters AppliedFilters) []bson.M {
	pipeline := []bson.M{}

	// Match stage for scraping logs
	matchStage := bson.M{
		"success": true,
		"scrape_timestamp": bson.M{
			"$gte": time.Now().Add(-time.Duration(filters.HoursBack) * time.Hour),
		},
	}

	// Add provider filter if specified
	if filters.Provider != "" {
		matchStage["provider"] = filters.Provider
	}

	// Add venue filter if specified
	if len(filters.VenueIDs) > 0 {
		var venueObjectIDs []primitive.ObjectID
		for _, id := range filters.VenueIDs {
			if objID, err := primitive.ObjectIDFromHex(id); err == nil {
				venueObjectIDs = append(venueObjectIDs, objID)
			}
		}
		if len(venueObjectIDs) > 0 {
			matchStage["venue_id"] = bson.M{"$in": venueObjectIDs}
		}
	}

	pipeline = append(pipeline, bson.M{"$match": matchStage})

	// Lookup venue details
	pipeline = append(pipeline, bson.M{
		"$lookup": bson.M{
			"from":         "venues",
			"localField":   "venue_id",
			"foreignField": "_id",
			"as":           "venue",
		},
	})

	// Unwind venue (should be exactly one)
	pipeline = append(pipeline, bson.M{
		"$unwind": "$venue",
	})

	// Sort by most recent scrapes first
	pipeline = append(pipeline, bson.M{
		"$sort": bson.M{"scrape_timestamp": -1},
	})

	// Group by venue to get the most recent scrape per venue
	pipeline = append(pipeline, bson.M{
		"$group": bson.M{
			"_id":          "$venue_id",
			"scraping_log": bson.M{"$first": "$$ROOT"},
			"venue":        bson.M{"$first": "$venue"},
		},
	})

	// Project the final structure
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"scraping_log": 1,
			"venue":        1,
		},
	})

	return pipeline
}

// matchesSlotFilters applies additional filters to individual slots
func (h *CourtsHandlers) matchesSlotFilters(slot models.Slot, filters AppliedFilters) bool {
	// Date range filter
	if filters.DateFrom != "" && slot.Date < filters.DateFrom {
		return false
	}
	if filters.DateTo != "" && slot.Date > filters.DateTo {
		return false
	}

	// Time range filter
	if filters.TimeFrom != "" || filters.TimeTo != "" {
		startTime := strings.Split(slot.Time, "-")[0]
		if filters.TimeFrom != "" && startTime < filters.TimeFrom {
			return false
		}
		if filters.TimeTo != "" && startTime > filters.TimeTo {
			return false
		}
	}

	// Price range filter
	if filters.PriceMin > 0 && slot.Price < filters.PriceMin {
		return false
	}
	if filters.PriceMax > 0 && slot.Price > filters.PriceMax {
		return false
	}

	// Day of week filter
	if len(filters.DayOfWeek) > 0 {
		if slotDate, err := time.Parse("2006-01-02", slot.Date); err == nil {
			dayOfWeek := strings.ToLower(slotDate.Weekday().String())
			found := false
			for _, filterDay := range filters.DayOfWeek {
				if strings.ToLower(filterDay) == dayOfWeek {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// createEnrichedSlot creates an enriched slot with additional metadata
func (h *CourtsHandlers) createEnrichedSlot(slot models.Slot, log models.ScrapingLog, venue models.Venue) AvailableSlot {
	// Parse time components
	timeParts := strings.Split(slot.Time, "-")
	startTime := ""
	endTime := ""
	if len(timeParts) == 2 {
		startTime = timeParts[0]
		endTime = timeParts[1]
	}

	// Calculate date metadata
	now := time.Now()
	slotDate, _ := time.Parse("2006-01-02", slot.Date)
	daysFromNow := int(slotDate.Sub(now).Hours() / 24)
	isToday := slotDate.Format("2006-01-02") == now.Format("2006-01-02")
	isTomorrow := slotDate.Format("2006-01-02") == now.AddDate(0, 0, 1).Format("2006-01-02")
	isWeekend := slotDate.Weekday() == time.Saturday || slotDate.Weekday() == time.Sunday

	// Determine time of day
	timeOfDay := h.getTimeOfDay(startTime)

	// Format price
	formattedPrice := formatPrice(slot.Price, "GBP") // Default to GBP

	// Calculate duration
	duration := calculateDuration(startTime, endTime)

	return AvailableSlot{
		ID:             log.ID,
		VenueID:        venue.ID.Hex(),
		VenueName:      venue.Name,
		Provider:       venue.Provider,
		CourtID:        slot.CourtID,
		CourtName:      slot.Court,
		Date:           slot.Date,
		StartTime:      startTime,
		EndTime:        endTime,
		TimeSlot:       slot.Time,
		Duration:       duration,
		Price:          slot.Price,
		FormattedPrice: formattedPrice,
		Currency:       "GBP",
		BookingURL:     slot.URL,
		ScrapedAt:      log.ScrapeTimestamp,
		DaysFromNow:    daysFromNow,
		IsToday:        isToday,
		IsTomorrow:     isTomorrow,
		IsWeekend:      isWeekend,
		TimeOfDay:      timeOfDay,
		VenueLocation: VenueLocationInfo{
			Address:   venue.Location.Address,
			City:      venue.Location.City,
			PostCode:  venue.Location.PostCode,
			Latitude:  venue.Location.Latitude,
			Longitude: venue.Location.Longitude,
		},
	}
}

// getTimeOfDay categorizes time into morning/afternoon/evening
func (h *CourtsHandlers) getTimeOfDay(timeStr string) string {
	if timeStr == "" {
		return ""
	}

	if t, err := time.Parse("15:04", timeStr); err == nil {
		hour := t.Hour()
		switch {
		case hour >= 6 && hour < 12:
			return "morning"
		case hour >= 12 && hour < 18:
			return "afternoon"
		case hour >= 18 && hour < 22:
			return "evening"
		default:
			return "night"
		}
	}

	return ""
}

// generateTimeRangeDescription creates a human-readable time range description
func (h *CourtsHandlers) generateTimeRangeDescription(filters AppliedFilters) string {
	parts := []string{}

	if filters.DateFrom != "" && filters.DateTo != "" {
		parts = append(parts, "from "+filters.DateFrom+" to "+filters.DateTo)
	} else if filters.DateFrom != "" {
		parts = append(parts, "from "+filters.DateFrom)
	} else if filters.DateTo != "" {
		parts = append(parts, "until "+filters.DateTo)
	}

	if filters.TimeFrom != "" && filters.TimeTo != "" {
		parts = append(parts, "between "+filters.TimeFrom+" and "+filters.TimeTo)
	} else if filters.TimeFrom != "" {
		parts = append(parts, "from "+filters.TimeFrom)
	} else if filters.TimeTo != "" {
		parts = append(parts, "until "+filters.TimeTo)
	}

	if len(filters.DayOfWeek) > 0 {
		parts = append(parts, "on "+strings.Join(filters.DayOfWeek, ", "))
	}

	if len(parts) == 0 {
		return "last " + strconv.Itoa(filters.HoursBack) + " hours"
	}

	return strings.Join(parts, ", ")
} 