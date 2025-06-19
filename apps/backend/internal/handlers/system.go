package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"tennis-booker/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SystemStatusResponse represents the response structure for the system status endpoint
type SystemStatusResponse struct {
	Status         string     `json:"status"`
	ScrapingStatus string     `json:"scrapingStatus"`
	LastUpdate     time.Time  `json:"lastUpdate"`
	LastScrapeTime *time.Time `json:"lastScrapeTime,omitempty"`
	ActiveJobs     int        `json:"activeJobs"`
	QueuedJobs     int        `json:"queuedJobs"`
	CompletedJobs  int        `json:"completedJobs"`
	ErroredJobs    int        `json:"erroredJobs"`
	SystemHealth   string     `json:"systemHealth"`
	Message        string     `json:"message"`
}

// SystemControlRequest represents system control requests
type SystemControlRequest struct {
	Action string `json:"action"`
	Reason string `json:"reason,omitempty"`
}

// SystemControlResponse represents system control responses
type SystemControlResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// SystemHandler handles system control requests
type SystemHandler struct {
	db database.Database
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(db database.Database) *SystemHandler {
	return &SystemHandler{
		db: db,
	}
}

// GetStatus handles GET /api/system/status
func (h *SystemHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get system status from database
	statusCollection := h.db.Collection("system_status")
	var systemStatus bson.M
	err := statusCollection.FindOne(ctx, bson.M{}).Decode(&systemStatus)
	
	// Default status if no record exists
	response := SystemStatusResponse{
		Status:         "running",
		ScrapingStatus: "active",
		LastUpdate:     time.Now(),
		LastScrapeTime: nil,
		ActiveJobs:     0,
		QueuedJobs:     0,
		CompletedJobs:  0,
		ErroredJobs:    0,
		SystemHealth:   "healthy",
		Message:        "System is operational",
	}

	if err == nil {
		// Parse existing status
		if status, ok := systemStatus["status"].(string); ok {
			response.Status = status
		}
		if scrapingStatus, ok := systemStatus["scrapingStatus"].(string); ok {
			response.ScrapingStatus = scrapingStatus
		}
		if lastUpdate, ok := systemStatus["lastUpdate"].(primitive.DateTime); ok {
			response.LastUpdate = lastUpdate.Time()
		}
		if activeJobs, ok := systemStatus["activeJobs"].(int32); ok {
			response.ActiveJobs = int(activeJobs)
		}
		if queuedJobs, ok := systemStatus["queuedJobs"].(int32); ok {
			response.QueuedJobs = int(queuedJobs)
		}
		if completedJobs, ok := systemStatus["completedJobs"].(int32); ok {
			response.CompletedJobs = int(completedJobs)
		}
		if erroredJobs, ok := systemStatus["erroredJobs"].(int32); ok {
			response.ErroredJobs = int(erroredJobs)
		}
		if systemHealth, ok := systemStatus["systemHealth"].(string); ok {
			response.SystemHealth = systemHealth
		}
		if message, ok := systemStatus["message"].(string); ok {
			response.Message = message
		}
	}

	// Get last scrape time from scraper status
	var scraperStatus bson.M
	err = statusCollection.FindOne(ctx, bson.M{"_id": "scraper_status"}).Decode(&scraperStatus)
	if err == nil {
		if lastScrapeTime, ok := scraperStatus["last_scrape_time"].(primitive.DateTime); ok {
			t := lastScrapeTime.Time()
			response.LastScrapeTime = &t
		}
	}

	// Update job counts from actual collections
	h.updateJobCounts(ctx, &response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PauseScraping handles POST /api/system/pause
func (h *SystemHandler) PauseScraping(w http.ResponseWriter, r *http.Request) {
	var req SystemControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body provided, just use default action
		req.Action = "pause"
		req.Reason = "Manual pause requested"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update system status
	statusCollection := h.db.Collection("system_status")
	update := bson.M{
		"$set": bson.M{
			"status":         "paused",
			"scrapingStatus": "paused",
			"lastUpdate":     time.Now(),
			"message":        "Scraping paused: " + req.Reason,
		},
	}

	_, err := statusCollection.UpdateOne(
		ctx,
		bson.M{},
		update,
		&options.UpdateOptions{Upsert: &[]bool{true}[0]},
	)

	response := SystemControlResponse{
		Success: err == nil,
		Status:  "paused",
	}

	if err != nil {
		response.Message = "Failed to pause scraping: " + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Message = "Scraping paused successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResumeScraping handles POST /api/system/resume
func (h *SystemHandler) ResumeScraping(w http.ResponseWriter, r *http.Request) {
	var req SystemControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body provided, just use default action
		req.Action = "resume"
		req.Reason = "Manual resume requested"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update system status
	statusCollection := h.db.Collection("system_status")
	update := bson.M{
		"$set": bson.M{
			"status":         "running",
			"scrapingStatus": "active",
			"lastUpdate":     time.Now(),
			"message":        "Scraping resumed: " + req.Reason,
		},
	}

	_, err := statusCollection.UpdateOne(
		ctx,
		bson.M{},
		update,
		&options.UpdateOptions{Upsert: &[]bool{true}[0]},
	)

	response := SystemControlResponse{
		Success: err == nil,
		Status:  "running",
	}

	if err != nil {
		response.Message = "Failed to resume scraping: " + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Message = "Scraping resumed successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RestartSystem handles POST /api/system/restart
func (h *SystemHandler) RestartSystem(w http.ResponseWriter, r *http.Request) {
	var req SystemControlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If no body provided, just use default action
		req.Action = "restart"
		req.Reason = "Manual restart requested"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update system status
	statusCollection := h.db.Collection("system_status")
	update := bson.M{
		"$set": bson.M{
			"status":         "restarting",
			"scrapingStatus": "restarting",
			"lastUpdate":     time.Now(),
			"message":        "System restart initiated: " + req.Reason,
		},
	}

	_, err := statusCollection.UpdateOne(
		ctx,
		bson.M{},
		update,
		&options.UpdateOptions{Upsert: &[]bool{true}[0]},
	)

	response := SystemControlResponse{
		Success: err == nil,
		Status:  "restarting",
	}

	if err != nil {
		response.Message = "Failed to restart system: " + err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Message = "System restart initiated successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// updateJobCounts updates job counts from database collections
func (h *SystemHandler) updateJobCounts(ctx context.Context, response *SystemStatusResponse) {
	// Count scraping jobs
	scrapingLogsCollection := h.db.Collection("scraping_logs")

	// Active jobs (running status)
	activeCount, err := scrapingLogsCollection.CountDocuments(ctx, bson.M{"status": "running"})
	if err == nil {
		response.ActiveJobs = int(activeCount)
	}

	// Queued jobs (pending status)
	queuedCount, err := scrapingLogsCollection.CountDocuments(ctx, bson.M{"status": "pending"})
	if err == nil {
		response.QueuedJobs = int(queuedCount)
	}

	// Completed jobs (completed status)
	completedCount, err := scrapingLogsCollection.CountDocuments(ctx, bson.M{"status": "completed"})
	if err == nil {
		response.CompletedJobs = int(completedCount)
	}

	// Errored jobs (error status)
	erroredCount, err := scrapingLogsCollection.CountDocuments(ctx, bson.M{"status": "error"})
	if err == nil {
		response.ErroredJobs = int(erroredCount)
	}
}

// GetScrapingLogs returns recent scraping logs for monitoring
func (h *SystemHandler) GetScrapingLogs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get query parameters
	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")
	venueID := query.Get("venueId")

	// Parse limit and offset
	limit := int64(50) // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 64); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := int64(0)
	if offsetStr != "" {
		if parsedOffset, err := strconv.ParseInt(offsetStr, 10, 64); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Query the database directly to handle the current schema
	scrapingLogsCollection := h.db.Collection("scraping_logs")
	
	// Build filter
	filter := bson.M{}
	if venueID != "" {
		venueObjectID, err := primitive.ObjectIDFromHex(venueID)
		if err != nil {
			http.Error(w, "Invalid venue ID", http.StatusBadRequest)
			return
		}
		filter["venue_id"] = venueObjectID
	}

	// Set up options
	opts := options.Find().
		SetSort(bson.D{{Key: "scrape_timestamp", Value: -1}}).
		SetSkip(offset).
		SetLimit(limit)

	cursor, err := scrapingLogsCollection.Find(ctx, filter, opts)
	if err != nil {

		http.Error(w, fmt.Sprintf("Failed to fetch scraping logs: %v", err), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Transform logs for frontend consumption
	type ScrapingLogResponse struct {
		ID               string    `json:"id"`
		VenueID          string    `json:"venueId"`
		VenueName        string    `json:"venueName"`
		Provider         string    `json:"provider"`
		Platform         string    `json:"platform"`
		ScrapeTimestamp  time.Time `json:"scrapeTimestamp"`
		Success          bool      `json:"success"`
		SlotsFound       int       `json:"slotsFound"`
		ScrapeDurationMs int       `json:"scrapeDurationMs"`
		Errors           []string  `json:"errors"`
		CreatedAt        time.Time `json:"createdAt"`
	}

	var response []ScrapingLogResponse
	for cursor.Next(ctx) {
		var rawLog bson.M
		if err := cursor.Decode(&rawLog); err != nil {
			continue // Skip invalid logs
		}

		// Extract fields safely
		id, _ := rawLog["_id"].(primitive.ObjectID)
		venueID, _ := rawLog["venue_id"].(primitive.ObjectID)
		venueName, _ := rawLog["venue_name"].(string)
		provider, _ := rawLog["provider"].(string)
		platform, _ := rawLog["platform"].(string)
		scrapeTimestamp, _ := rawLog["scrape_timestamp"].(primitive.DateTime)
		success, _ := rawLog["success"].(bool)
		slotsFound, _ := rawLog["slots_found"].(int32)
		scrapeDurationMs, _ := rawLog["scrape_duration_ms"].(int32)
		createdAt, _ := rawLog["created_at"].(primitive.DateTime)
		
		// Handle errors field
		var errors []string
		if errorsInterface, ok := rawLog["errors"]; ok {
			if errorsArray, ok := errorsInterface.(primitive.A); ok {
				for _, err := range errorsArray {
					if errStr, ok := err.(string); ok {
						errors = append(errors, errStr)
					}
				}
			}
		}

		response = append(response, ScrapingLogResponse{
			ID:               id.Hex(),
			VenueID:          venueID.Hex(),
			VenueName:        venueName,
			Provider:         provider,
			Platform:         platform,
			ScrapeTimestamp:  scrapeTimestamp.Time(),
			Success:          success,
			SlotsFound:       int(slotsFound),
			ScrapeDurationMs: int(scrapeDurationMs),
			Errors:           errors,
			CreatedAt:        createdAt.Time(),
		})
	}

	if err := cursor.Err(); err != nil {
		
		http.Error(w, fmt.Sprintf("Error reading scraping logs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
