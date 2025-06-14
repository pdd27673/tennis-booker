package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CourtSlot represents a tennis court time slot available for booking
type CourtSlot struct {
	ID            string             `json:"id"`              // Unique identifier for the slot
	VenueID       primitive.ObjectID `json:"venue_id"`        // Reference to the venue
	VenueName     string             `json:"venue_name"`      // Venue name for convenience
	CourtID       string             `json:"court_id"`        // Court identifier
	CourtName     string             `json:"court_name"`      // Human-readable court name
	Date          string             `json:"date"`            // Format: "YYYY-MM-DD"
	StartTime     string             `json:"start_time"`      // Format: "HH:MM"
	EndTime       string             `json:"end_time"`        // Format: "HH:MM"
	Price         float64            `json:"price"`           // Price for the slot
	Currency      string             `json:"currency"`        // Currency code (e.g., "GBP", "USD")
	Available     bool               `json:"available"`       // Whether the slot is available
	BookingURL    string             `json:"booking_url"`     // Direct booking URL if available
	Provider      string             `json:"provider"`        // Provider type (e.g., "lta", "courtsides")
	LastScraped   time.Time          `json:"last_scraped"`    // When this slot was last found
	ScrapingLogID primitive.ObjectID `json:"scraping_log_id"` // Reference to the scraping log
}

// GenerateSlotID creates a unique identifier for a court slot
func (cs *CourtSlot) GenerateSlotID() string {
	return cs.VenueID.Hex() + "_" + cs.CourtID + "_" + cs.Date + "_" + cs.StartTime
}

// CourtSlotFilter represents filtering options for court slots
type CourtSlotFilter struct {
	VenueID   *primitive.ObjectID `json:"venue_id,omitempty"`
	Date      *string             `json:"date,omitempty"`       // Format: "YYYY-MM-DD"
	StartTime *string             `json:"start_time,omitempty"` // Format: "HH:MM"
	EndTime   *string             `json:"end_time,omitempty"`   // Format: "HH:MM"
	Available *bool               `json:"available,omitempty"`
	Provider  *string             `json:"provider,omitempty"`
	MinPrice  *float64            `json:"min_price,omitempty"`
	MaxPrice  *float64            `json:"max_price,omitempty"`
}

// CourtSlotService provides methods for interacting with court slots
type CourtSlotService struct {
	// Will be implemented later with repository dependencies
}

// Collection returns the name of the MongoDB collection for court slots
// Note: Court slots are derived from scraping_logs, not stored separately
func (CourtSlotService) Collection() string {
	return "scraping_logs"
}
