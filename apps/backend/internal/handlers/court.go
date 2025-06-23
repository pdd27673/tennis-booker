package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// SlotsRepositoryInterface defines the interface for slots repository operations
type SlotsRepositoryInterface interface {
	GetAvailableSlots(ctx context.Context, limit int64) ([]*models.CourtSlot, error)
	GetAvailableSlotsByVenue(ctx context.Context, venueID primitive.ObjectID, limit int64) ([]*models.CourtSlot, error)
	GetAvailableSlotsByDate(ctx context.Context, date string, limit int64) ([]*models.CourtSlot, error)
	CountAvailableSlots(ctx context.Context) (int64, error)
	CountSlotsByDate(ctx context.Context, date string) (int64, error)
	CountSlotsByDateRange(ctx context.Context, startDate, endDate string) (int64, error)
	GetActivePlatforms(ctx context.Context) ([]string, error)
}

// CourtHandler handles court and venue related requests
type CourtHandler struct {
	db              database.Database
	scrapingLogRepo ScrapingLogRepositoryInterface
	slotsRepo       SlotsRepositoryInterface
}

// NewCourtHandler creates a new court handler
func NewCourtHandler(db database.Database) *CourtHandler {
	// Create repositories
	scrapingLogRepo := database.NewScrapingLogRepository(db.GetMongoDB())
	slotsRepo := database.NewSlotsRepository(db.GetMongoDB())

	return &CourtHandler{
		db:              db,
		scrapingLogRepo: scrapingLogRepo,
		slotsRepo:       slotsRepo,
	}
}

// VenueResponse represents venue data for API responses
type VenueResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	PostCode    string `json:"postCode"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Website     string `json:"website"`
	Platform    string `json:"platform"`
	PlatformID  string `json:"platformId"`
	Coordinates struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinates"`
	TotalCourts int `json:"totalCourts"`
}

// CourtSlotResponse represents court slot data for API responses
type CourtSlotResponse struct {
	ID         string    `json:"id"`
	VenueID    string    `json:"venueId"`
	VenueName  string    `json:"venueName"`
	CourtID    string    `json:"courtId"`
	CourtName  string    `json:"courtName"`
	Date       string    `json:"date"`
	StartTime  string    `json:"startTime"`
	EndTime    string    `json:"endTime"`
	Duration   int       `json:"duration"`
	Price      float64   `json:"price"`
	Currency   string    `json:"currency"`
	Available  bool      `json:"available"`
	Platform   string    `json:"platform"`
	BookingURL string    `json:"bookingUrl"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// DashboardStatsResponse represents dashboard statistics
type DashboardStatsResponse struct {
	TotalVenues     int `json:"totalVenues"`
	TotalCourtSlots int `json:"totalCourtSlots"`
	AvailableSlots  int `json:"availableSlots"`
	TodaySlots      int `json:"todaySlots"`
	WeekSlots       int `json:"weekSlots"`
	ActivePlatforms int `json:"activePlatforms"`
}

// GetVenues handles the GET /api/venues endpoint
func (h *CourtHandler) GetVenues(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get query parameters
	query := r.URL.Query()
	platform := query.Get("platform")
	city := query.Get("city")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	// Build filter
	filter := bson.M{}
	if platform != "" {
		filter["platform"] = platform
	}
	if city != "" {
		// Sanitize city input to prevent regex injection
		sanitizedCity := strings.ReplaceAll(strings.ReplaceAll(city, "\\", "\\\\"), "$", "\\$")
		filter["city"] = bson.M{"$regex": sanitizedCity, "$options": "i"}
	}

	// Set up options
	opts := options.Find()
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			opts.SetLimit(int64(limit))
		}
	}
	if offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			opts.SetSkip(int64(offset))
		}
	}

	// Sort by name
	opts.SetSort(bson.D{{Key: "name", Value: 1}})

	// Query venues
	collection := h.db.Collection("venues")
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		http.Error(w, "Failed to fetch venues", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var venues []models.Venue
	if err := cursor.All(ctx, &venues); err != nil {
		http.Error(w, "Failed to decode venues", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := make([]VenueResponse, len(venues))
	for i, venue := range venues {
		response[i] = VenueResponse{
			ID:         venue.ID.Hex(),
			Name:       venue.Name,
			Address:    venue.Location.Address,
			City:       venue.Location.City,
			PostCode:   venue.Location.PostCode,
			Phone:      "", // Not available in current model
			Email:      "", // Not available in current model
			Website:    venue.URL,
			Platform:   venue.Provider,
			PlatformID: venue.ID.Hex(), // Use venue ID as platform ID
			Coordinates: struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			}{
				Lat: venue.Location.Latitude,
				Lng: venue.Location.Longitude,
			},
			TotalCourts: len(venue.Courts),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCourtSlots handles the GET /api/courts endpoint
func (h *CourtHandler) GetCourtSlots(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get query parameters
	query := r.URL.Query()
	venueID := query.Get("venueId")
	date := query.Get("date")
	limitStr := query.Get("limit")

	// Parse limit
	limit := int64(100) // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = int64(parsedLimit)
		}
	}

	var courtSlots []*models.CourtSlot
	var err error

	// If venue ID is specified, use venue-specific query
	if venueID != "" {
		venueObjID, err := primitive.ObjectIDFromHex(venueID)
		if err != nil {
			http.Error(w, "Invalid venue ID", http.StatusBadRequest)
			return
		}
		courtSlots, err = h.slotsRepo.GetAvailableSlotsByVenue(ctx, venueObjID, limit)
	} else if date != "" {
		// If date is specified, use date-specific query
		courtSlots, err = h.slotsRepo.GetAvailableSlotsByDate(ctx, date, limit)
	} else {
		// General query for all available slots
		courtSlots, err = h.slotsRepo.GetAvailableSlots(ctx, limit)
	}

	if err != nil {
		http.Error(w, "Failed to fetch court slots", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := make([]CourtSlotResponse, len(courtSlots))
	for i, slot := range courtSlots {
		response[i] = CourtSlotResponse{
			ID:         slot.ID,
			VenueID:    slot.VenueID.Hex(),
			VenueName:  slot.VenueName,
			CourtID:    slot.CourtID,
			CourtName:  slot.CourtName,
			Date:       slot.Date,
			StartTime:  slot.StartTime,
			EndTime:    slot.EndTime,
			Duration:   calculateDuration(slot.StartTime, slot.EndTime),
			Price:      slot.Price,
			Currency:   slot.Currency,
			Available:  slot.Available,
			Platform:   slot.Provider,
			BookingURL: slot.BookingURL,
			CreatedAt:  slot.LastScraped,
			UpdatedAt:  slot.LastScraped,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDashboardStats provides statistics for the dashboard
func (h *CourtHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stats := DashboardStatsResponse{}

	// Count total venues
	venueCollection := h.db.Collection("venues")
	totalVenues, err := venueCollection.CountDocuments(ctx, bson.M{})
	if err == nil {
		stats.TotalVenues = int(totalVenues)
	}

	// Get available court slots count
	availableCount, err := h.slotsRepo.CountAvailableSlots(ctx)
	if err == nil {
		stats.TotalCourtSlots = int(availableCount)
		stats.AvailableSlots = int(availableCount) // All counted slots are available

		// Count today's slots
		today := time.Now().Format("2006-01-02")
		todayCount, todayErr := h.slotsRepo.CountSlotsByDate(ctx, today)
		if todayErr == nil {
			stats.TodaySlots = int(todayCount)
		}

		// Count this week's slots using efficient date range query
		now := time.Now()
		weekStart := now.AddDate(0, 0, -int(now.Weekday()))
		weekEnd := weekStart.AddDate(0, 0, 7)
		weekStartStr := weekStart.Format("2006-01-02")
		weekEndStr := weekEnd.Format("2006-01-02")

		weekCount, weekErr := h.slotsRepo.CountSlotsByDateRange(ctx, weekStartStr, weekEndStr)
		if weekErr == nil {
			stats.WeekSlots = int(weekCount)
		}

		// Count active platforms
		platforms, platformErr := h.slotsRepo.GetActivePlatforms(ctx)
		if platformErr == nil {
			stats.ActivePlatforms = len(platforms)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// calculateDuration calculates the duration in minutes between start and end time
func calculateDuration(startTime, endTime string) int {
	// Parse time format "HH:MM"
	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return 0
	}

	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return 0
	}

	// Handle case where end time is next day (e.g., 23:00 to 01:00)
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}

	duration := end.Sub(start)
	return int(duration.Minutes())
}
