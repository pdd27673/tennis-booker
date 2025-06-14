package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tennis-booker/internal/auth"
)

// HealthResponse represents the response structure for the health endpoint
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// SystemStatusResponse represents the response structure for the system status endpoint
type SystemStatusResponse struct {
	ScrapingStatus   string     `json:"scraping_status"`
	LastRunTimestamp *time.Time `json:"last_run_timestamp,omitempty"`
	ItemsProcessed   int        `json:"items_processed"`
	ErrorCount       int        `json:"error_count"`
	SystemUptime     string     `json:"system_uptime"`
	Timestamp        time.Time  `json:"timestamp"`
}

// SystemControlResponse represents the response structure for system control endpoints
type SystemControlResponse struct {
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// SystemHandler handles system-related endpoints
type SystemHandler struct {
	version   string
	startTime time.Time
	// In a real implementation, these would be injected dependencies
	// For now, we'll simulate the scraping system state
	scrapingStatus string
	lastRunTime    *time.Time
	itemsProcessed int
	errorCount     int
}

// NewSystemHandler creates a new SystemHandler instance
func NewSystemHandler(version string) *SystemHandler {
	return &SystemHandler{
		version:        version,
		startTime:      time.Now(),
		scrapingStatus: "running", // Default status
		itemsProcessed: 0,
		errorCount:     0,
	}
}

// Health handles GET /api/health - returns the health status of the service
func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create health response
	response := HealthResponse{
		Status:    "UP",
		Timestamp: time.Now().UTC(),
		Version:   h.version,
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// If we can't encode the response, something is seriously wrong
		// But at this point we've already written the status code, so we can't change it
		// Log the error if we had a logger available
		return
	}
}

// Status handles GET /api/system/status - returns the current status of the scraping system
func (h *SystemHandler) Status(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user claims from context (set by JWT middleware)
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Any authenticated user can view system status
	_ = claims // Use claims to avoid unused variable error

	// Calculate system uptime
	uptime := time.Since(h.startTime)
	uptimeStr := formatDuration(uptime)

	// Create status response
	response := SystemStatusResponse{
		ScrapingStatus:   h.scrapingStatus,
		LastRunTimestamp: h.lastRunTime,
		ItemsProcessed:   h.itemsProcessed,
		ErrorCount:       h.errorCount,
		SystemUptime:     uptimeStr,
		Timestamp:        time.Now().UTC(),
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.writeErrorResponse(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Pause handles POST /api/system/pause - pauses the scraping system
func (h *SystemHandler) Pause(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user claims from context (set by JWT middleware)
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Any authenticated user can control the system
	_ = claims // Use claims to avoid unused variable error

	// Check if system is already paused
	if h.scrapingStatus == "paused" {
		h.writeErrorResponse(w, "System is already paused", http.StatusBadRequest)
		return
	}

	// Pause the system
	h.scrapingStatus = "paused"

	// In a real implementation, this would trigger actual pause logic:
	// - Stop scraping workers
	// - Cancel ongoing operations
	// - Update database state
	// - Send signals to background processes

	// Create response
	response := SystemControlResponse{
		Message:   "System pausing initiated",
		Status:    h.scrapingStatus,
		Timestamp: time.Now().UTC(),
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.writeErrorResponse(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// Resume handles POST /api/system/resume - resumes the scraping system
func (h *SystemHandler) Resume(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user claims from context (set by JWT middleware)
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// TODO: In a real implementation, check for admin role
	// For now, any authenticated user can control the system
	_ = claims // Use claims to avoid unused variable error

	// Check if system is already running
	if h.scrapingStatus == "running" {
		h.writeErrorResponse(w, "System is already running", http.StatusBadRequest)
		return
	}

	// Resume the system
	h.scrapingStatus = "running"

	// In a real implementation, this would trigger actual resume logic:
	// - Start scraping workers
	// - Resume scheduled operations
	// - Update database state
	// - Send signals to background processes

	// Create response
	response := SystemControlResponse{
		Message:   "System resuming initiated",
		Status:    h.scrapingStatus,
		Timestamp: time.Now().UTC(),
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.writeErrorResponse(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
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

// writeErrorResponse writes an error response in JSON format
func (h *SystemHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResp)
}
