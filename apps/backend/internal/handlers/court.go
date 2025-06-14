package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"tennis-booker/internal/database"
	"tennis-booker/internal/models"
)

// VenueRepositoryInterface defines the interface for venue repository operations
type VenueRepositoryInterface interface {
	ListActive(ctx context.Context) ([]*models.Venue, error)
}

// ScrapingLogRepositoryInterface defines the interface for scraping log repository operations
type ScrapingLogRepositoryInterface interface {
	GetAvailableCourtSlots(ctx context.Context, limit int64) ([]*models.CourtSlot, error)
	GetAvailableCourtSlotsByVenue(ctx context.Context, venueID primitive.ObjectID, limit int64) ([]*models.CourtSlot, error)
	GetAvailableCourtSlotsWithFilters(ctx context.Context, filter models.CourtSlotFilter, limit int64) ([]*models.CourtSlot, error)
}

// CourtHandler handles court and venue related HTTP requests
type CourtHandler struct {
	venueRepo       VenueRepositoryInterface
	scrapingLogRepo ScrapingLogRepositoryInterface
}

// NewCourtHandler creates a new court handler
func NewCourtHandler(venueRepo VenueRepositoryInterface, scrapingLogRepo ScrapingLogRepositoryInterface) *CourtHandler {
	return &CourtHandler{
		venueRepo:       venueRepo,
		scrapingLogRepo: scrapingLogRepo,
	}
}

// NewCourtHandlerWithDB creates a new court handler with database connections
func NewCourtHandlerWithDB(venueRepo *database.VenueRepository, scrapingLogRepo *database.ScrapingLogRepository) *CourtHandler {
	return &CourtHandler{
		venueRepo:       venueRepo,
		scrapingLogRepo: scrapingLogRepo,
	}
}

// ListVenues handles GET /api/venues - returns list of all venues
func (h *CourtHandler) ListVenues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get venues from database
	venues, err := h.venueRepo.ListActive(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve venues", http.StatusInternalServerError)
		return
	}

	// Return venues as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(venues)
}

// ListCourts handles GET /api/courts - returns list of available court slots
func (h *CourtHandler) ListCourts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters for filtering
	query := r.URL.Query()
	filter := models.CourtSlotFilter{}
	hasFilters := false

	// Parse venue ID filter
	if venueIDStr := query.Get("venueId"); venueIDStr != "" {
		if venueID, err := primitive.ObjectIDFromHex(venueIDStr); err == nil {
			filter.VenueID = &venueID
			hasFilters = true
		} else {
			http.Error(w, "Invalid venueId format", http.StatusBadRequest)
			return
		}
	}

	// Parse date filter (YYYY-MM-DD format)
	if dateStr := query.Get("date"); dateStr != "" {
		// Validate date format
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		filter.Date = &dateStr
		hasFilters = true
	}

	// Parse start time filter (HH:MM format)
	if startTimeStr := query.Get("startTime"); startTimeStr != "" {
		// Validate time format
		if _, err := time.Parse("15:04", startTimeStr); err != nil {
			http.Error(w, "Invalid startTime format. Use HH:MM", http.StatusBadRequest)
			return
		}
		filter.StartTime = &startTimeStr
		hasFilters = true
	}

	// Parse end time filter (HH:MM format)
	if endTimeStr := query.Get("endTime"); endTimeStr != "" {
		// Validate time format
		if _, err := time.Parse("15:04", endTimeStr); err != nil {
			http.Error(w, "Invalid endTime format. Use HH:MM", http.StatusBadRequest)
			return
		}
		filter.EndTime = &endTimeStr
		hasFilters = true
	}

	// Parse provider filter
	if providerStr := query.Get("provider"); providerStr != "" {
		filter.Provider = &providerStr
		hasFilters = true
	}

	// Parse price filters
	if minPriceStr := query.Get("minPrice"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filter.MinPrice = &minPrice
			hasFilters = true
		} else {
			http.Error(w, "Invalid minPrice format", http.StatusBadRequest)
			return
		}
	}

	if maxPriceStr := query.Get("maxPrice"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filter.MaxPrice = &maxPrice
			hasFilters = true
		} else {
			http.Error(w, "Invalid maxPrice format", http.StatusBadRequest)
			return
		}
	}

	// Parse limit parameter (optional, defaults to 100)
	limit := int64(100)
	if limitStr := query.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var courtSlots []*models.CourtSlot
	var err error

	// Use filtering method if filters are provided, otherwise use the basic method
	if hasFilters {
		courtSlots, err = h.scrapingLogRepo.GetAvailableCourtSlotsWithFilters(r.Context(), filter, limit)
	} else {
		courtSlots, err = h.scrapingLogRepo.GetAvailableCourtSlots(r.Context(), limit)
	}

	if err != nil {
		http.Error(w, "Failed to retrieve court slots", http.StatusInternalServerError)
		return
	}

	// Return court slots as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courtSlots)
}
