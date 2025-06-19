package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tennis-booker/internal/models"
)

func main() {
	log.Println("Starting venue seeding process...")

	// Get MongoDB URI from environment or use default
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "tennis_booking"
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(dbName)
	log.Printf("Connected to MongoDB database: %s", dbName)

	// Clear existing data
	log.Println("Clearing existing data...")

	collections := []string{"venues", "scraping_logs", "bookings", "slots"}
	for _, collName := range collections {
		result, err := db.Collection(collName).DeleteMany(ctx, bson.M{})
		if err != nil {
			log.Printf("Warning: Failed to clear collection %s: %v", collName, err)
		} else {
			log.Printf("Cleared %d documents from %s collection", result.DeletedCount, collName)
		}
	}

	// Seed venues
	log.Println("Seeding venues...")

	venues := []models.Venue{
		{
			ID:       primitive.NewObjectID(),
			Name:     "Victoria Park",
			Provider: "courtsides",
			URL:      "https://tennistowerhamlets.com/book/courts/victoria-park#book",
			Location: models.Location{
				Address:  "Victoria Park, London",
				City:     "London",
				PostCode: "E9 7DE",
			},
			Courts: []models.Court{
				{ID: "1", Name: "Court 1", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "2", Name: "Court 2", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "3", Name: "Court 3", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "4", Name: "Court 4", Surface: "Hard", Indoor: false, Floodlights: true},
			},
			BookingWindow: 7,
			ScraperConfig: models.ScraperConfig{
				Type:               "courtside",
				RequiresLogin:      false,
				RetryCount:         3,
				TimeoutSeconds:     30,
				WaitAfterLoadMs:    2000,
				UseHeadlessBrowser: true,
				CustomParameters: map[string]interface{}{
					"date_selector":      ".day-picker",
					"court_selector":     ".court-widget",
					"booking_selector":   "input.bookable",
					"price_selector":     "[data-price]",
					"available_selector": "span.button.available",
				},
			},
			ScrapingInterval: 30,
			IsActive:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Stratford Park",
			Provider: "lta_clubspark",
			URL:      "https://stratford.newhamparkstennis.org.uk/Booking/BookByDate#?date=2025-06-09&role=guest",
			Location: models.Location{
				Address:  "Stratford Park, London",
				City:     "London",
				PostCode: "E15 1DA",
			},
			Courts: []models.Court{
				{ID: "1", Name: "Court 1", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "2", Name: "Court 2", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "3", Name: "Court 3", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "4", Name: "Court 4", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "5", Name: "Court 5", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "6", Name: "Court 6", Surface: "Hard", Indoor: false, Floodlights: true},
			},
			BookingWindow: 7,
			ScraperConfig: models.ScraperConfig{
				Type:               "clubspark",
				RequiresLogin:      false,
				RetryCount:         3,
				TimeoutSeconds:     30,
				WaitAfterLoadMs:    2000,
				UseHeadlessBrowser: true,
				CustomParameters: map[string]interface{}{
					"booking_grid_selector": ".booking-grid",
					"slot_selector":         "a.book-interval.not-booked",
					"price_selector":        "span.cost",
					"data_test_id":          "data-test-id",
					"guest_role":            true,
					"session_cost_selector": "[data-session-cost]",
				},
			},
			ScrapingInterval: 15,
			IsActive:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Ropemakers Field",
			Provider: "courtsides",
			URL:      "https://tennistowerhamlets.com/book/courts/ropemakers-field#book",
			Location: models.Location{
				Address:  "Ropemakers Field, London",
				City:     "London",
				PostCode: "E14 0JY",
			},
			Courts: []models.Court{
				{ID: "1", Name: "Court 1", Surface: "Hard", Indoor: false, Floodlights: true},
				{ID: "2", Name: "Court 2", Surface: "Hard", Indoor: false, Floodlights: true},
			},
			BookingWindow: 7,
			ScraperConfig: models.ScraperConfig{
				Type:               "courtside",
				RequiresLogin:      false,
				RetryCount:         3,
				TimeoutSeconds:     30,
				WaitAfterLoadMs:    2000,
				UseHeadlessBrowser: true,
				CustomParameters: map[string]interface{}{
					"date_selector":      ".day-picker",
					"court_selector":     ".court-widget",
					"booking_selector":   "input.bookable",
					"price_selector":     "[data-price]",
					"available_selector": "span.button.available",
				},
			},
			ScrapingInterval: 30,
			IsActive:         true,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	// Insert venues
	venueCollection := db.Collection("venues")
	for _, venue := range venues {
		_, err := venueCollection.InsertOne(ctx, venue)
		if err != nil {
			log.Fatalf("Failed to insert venue %s: %v", venue.Name, err)
		}
		log.Printf("✅ Inserted venue: %s (%s provider)", venue.Name, venue.Provider)
	}

	log.Printf("🎾 Successfully seeded %d venues!", len(venues))
	log.Println("Venues seeded:")
	for _, venue := range venues {
		log.Printf("  - %s (%d courts, %s provider, %d min intervals)",
			venue.Name, len(venue.Courts), venue.Provider, venue.ScrapingInterval)
	}
}
