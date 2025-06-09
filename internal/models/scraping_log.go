package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ScrapingLog represents a log of a scraping operation
type ScrapingLog struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	VenueID          primitive.ObjectID `bson:"venue_id" json:"venue_id"`
	ScrapeTimestamp  time.Time          `bson:"scrape_timestamp" json:"scrape_timestamp"`
	SlotsFound       []Slot             `bson:"slots_found" json:"slots_found"`
	ScrapeDurationMs int                `bson:"scrape_duration_ms" json:"scrape_duration_ms"`
	Errors           []string           `bson:"errors,omitempty" json:"errors,omitempty"`
	Success          bool               `bson:"success" json:"success"`
	VenueName        string             `bson:"venue_name" json:"venue_name"` // Denormalized for easier querying
	Provider         string             `bson:"provider" json:"provider"`     // Denormalized from venue
	RawResponse      string             `bson:"raw_response,omitempty" json:"raw_response,omitempty"` // Optional raw response for debugging
	ScraperVersion   string             `bson:"scraper_version,omitempty" json:"scraper_version,omitempty"`
	UserAgent        string             `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	IPAddress        string             `bson:"ip_address,omitempty" json:"ip_address,omitempty"`
	RunID            string             `bson:"run_id,omitempty" json:"run_id,omitempty"` // To group multiple scrapes
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
}

// Slot represents an available court slot found during scraping
type Slot struct {
	Date      string  `bson:"date" json:"date"`           // Format: "YYYY-MM-DD"
	Time      string  `bson:"time" json:"time"`           // Format: "HH:MM-HH:MM"
	Court     string  `bson:"court" json:"court"`
	Price     float64 `bson:"price,omitempty" json:"price,omitempty"`
	Available bool    `bson:"available" json:"available"`
	CourtID   string  `bson:"court_id,omitempty" json:"court_id,omitempty"`
	URL       string  `bson:"url,omitempty" json:"url,omitempty"` // Direct booking URL if available
}

// ScrapingLogService provides methods for interacting with scraping logs
type ScrapingLogService struct {
	// Will be implemented later with MongoDB connection
}

// Collection returns the name of the MongoDB collection for scraping logs
func (ScrapingLogService) Collection() string {
	return "scraping_logs"
} 