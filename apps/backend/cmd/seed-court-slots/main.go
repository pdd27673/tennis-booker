package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"tennis-booker/internal/config"
	"tennis-booker/internal/database"
	"tennis-booker/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	mongoDb, err := database.InitDatabase(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()

	// Get existing venues
	venueCollection := mongoDb.Collection("venues")
	cursor, err := venueCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to fetch venues: %v", err)
	}
	defer cursor.Close(ctx)

	var venues []models.Venue
	if err := cursor.All(ctx, &venues); err != nil {
		log.Fatalf("Failed to decode venues: %v", err)
	}

	if len(venues) == 0 {
		log.Fatalf("No venues found in database. Please seed venues first.")
	}

	// Create scraping logs collection
	scrapingLogCollection := mongoDb.Collection("scraping_logs")

	now := time.Now()
	totalSlots := 0

	// Generate court slots for next 3 days
	dates := []string{
		now.Format("2006-01-02"),
		now.AddDate(0, 0, 1).Format("2006-01-02"),
		now.AddDate(0, 0, 2).Format("2006-01-02"),
	}

	for _, venue := range venues {
		log.Printf("Creating court slots for venue: %s", venue.Name)
		
		for dateIndex, date := range dates {
			var slots []models.Slot
			
			// Generate slots based on number of courts
			numCourts := len(venue.Courts)
			if numCourts == 0 {
				// Default to 2 courts if none specified
				numCourts = 2
			}

			// Generate time slots from 9 AM to 9 PM (12 hours)
			for hour := 9; hour < 21; hour++ {
				for court := 1; court <= numCourts; court++ {
					slot := models.Slot{
						Date:      date,
						Time:      fmt.Sprintf("%02d:00-%02d:00", hour, hour+1),
						Court:     fmt.Sprintf("Court %d", court),
						Price:     25.0 + float64(court*2), // Varying prices
						Available: true,
						CourtID:   fmt.Sprintf("court_%d", court),
						URL:       fmt.Sprintf("https://booking.example.com/venue/%s/court/%d", venue.ID.Hex(), court),
					}
					slots = append(slots, slot)
				}
			}

			// Create scraping log for this venue and date
			scrapingLog := models.ScrapingLog{
				ID:               primitive.NewObjectID(),
				VenueID:          venue.ID,
				VenueName:        venue.Name,
				Provider:         venue.Provider,
				ScrapeTimestamp:  now.Add(-time.Duration(dateIndex) * time.Hour), // Stagger timestamps
				SlotsFound:       slots,
				ScrapeDurationMs: 2500,
				Success:          true,
				Errors:           []string{},
				CreatedAt:        now,
			}

			// Insert the scraping log
			_, err := scrapingLogCollection.InsertOne(ctx, scrapingLog)
			if err != nil {
				log.Printf("Failed to insert scraping log for %s on %s: %v", venue.Name, date, err)
				continue
			}

			log.Printf("✅ Created %d slots for %s on %s", len(slots), venue.Name, date)
			totalSlots += len(slots)
		}
	}

	log.Printf("🎾 Successfully seeded court slots!")
	log.Printf("📊 Total available slots created: %d", totalSlots)
	log.Printf("🏟️ Venues with slots: %d", len(venues))
	log.Printf("Court slots seeded for:")
	for _, venue := range venues {
		numCourts := len(venue.Courts)
		if numCourts == 0 {
			numCourts = 2
		}
		log.Printf("  - %s (%d courts, %s provider)", venue.Name, numCourts, venue.Provider)
	}
} 