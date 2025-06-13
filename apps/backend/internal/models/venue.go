package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Venue represents a tennis court venue in the system
type Venue struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Provider         string             `bson:"provider" json:"provider"` // "lta", "courtsides", etc.
	URL              string             `bson:"url" json:"url"`
	Location         Location           `bson:"location" json:"location"`
	Courts           []Court            `bson:"courts" json:"courts"`
	BookingWindow    int                `bson:"booking_window" json:"booking_window"` // Days in advance booking is allowed
	ScraperConfig    ScraperConfig      `bson:"scraper_config" json:"scraper_config"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
	LastScrapedAt    time.Time          `bson:"last_scraped_at,omitempty" json:"last_scraped_at,omitempty"`
	ScrapingInterval int                `bson:"scraping_interval" json:"scraping_interval"` // Minutes between scrapes
	IsActive         bool               `bson:"is_active" json:"is_active"`
}

// Location represents the geographical location of a venue
type Location struct {
	Address   string  `bson:"address" json:"address"`
	City      string  `bson:"city" json:"city"`
	PostCode  string  `bson:"post_code" json:"post_code"`
	Latitude  float64 `bson:"latitude,omitempty" json:"latitude,omitempty"`
	Longitude float64 `bson:"longitude,omitempty" json:"longitude,omitempty"`
}

// Court represents a tennis court within a venue
type Court struct {
	ID          string   `bson:"id" json:"id"`
	Name        string   `bson:"name" json:"name"`
	Surface     string   `bson:"surface,omitempty" json:"surface,omitempty"` // "hard", "clay", "grass", etc.
	Indoor      bool     `bson:"indoor" json:"indoor"`
	Floodlights bool     `bson:"floodlights,omitempty" json:"floodlights,omitempty"`
	CourtType   string   `bson:"court_type,omitempty" json:"court_type,omitempty"` // "singles", "doubles"
	Tags        []string `bson:"tags,omitempty" json:"tags,omitempty"`
}

// ScraperConfig represents configuration for scraping a venue
type ScraperConfig struct {
	Type              string                 `bson:"type" json:"type"`                             // "clubspark", "courtsides", etc.
	RequiresLogin     bool                   `bson:"requires_login" json:"requires_login"`         // Whether login is required to scrape
	CredentialKey     string                 `bson:"credential_key,omitempty" json:"credential_key,omitempty"` // Key to retrieve credentials from Vault
	CustomParameters  map[string]interface{} `bson:"custom_parameters,omitempty" json:"custom_parameters,omitempty"`
	SelectorMappings  map[string]string      `bson:"selector_mappings,omitempty" json:"selector_mappings,omitempty"`
	NavigationSteps   []string               `bson:"navigation_steps,omitempty" json:"navigation_steps,omitempty"`
	RetryCount        int                    `bson:"retry_count" json:"retry_count"`
	TimeoutSeconds    int                    `bson:"timeout_seconds" json:"timeout_seconds"`
	WaitAfterLoadMs   int                    `bson:"wait_after_load_ms" json:"wait_after_load_ms"`
	UserAgent         string                 `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	UseHeadlessBrowser bool                  `bson:"use_headless_browser" json:"use_headless_browser"`
}

// VenueService provides methods for interacting with venues
type VenueService struct {
	// Will be implemented later with MongoDB connection
}

// Collection returns the name of the MongoDB collection for venues
func (VenueService) Collection() string {
	return "venues"
} 