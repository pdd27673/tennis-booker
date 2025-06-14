package database

import (
	"context"
	"testing"
	"time"

	"tennis-booker/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestVenueRepository_Create(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create a test venue
	venue := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub",
		Location: models.Location{
			Address:   "123 Tennis Court Road",
			City:      "London",
			PostCode:  "SW19 5AE",
			Latitude:  51.4341,
			Longitude: -0.2147,
		},
		Courts: []models.Court{
			{
				ID:          "court1",
				Name:        "Court 1",
				Surface:     "hard",
				Indoor:      false,
				Floodlights: true,
			},
			{
				ID:          "court2",
				Name:        "Court 2",
				Surface:     "clay",
				Indoor:      false,
				Floodlights: true,
			},
		},
		BookingWindow:    7,
		ScrapingInterval: 30,
		IsActive:         true,
		ScraperConfig: models.ScraperConfig{
			Type:               "clubspark",
			RequiresLogin:      true,
			CredentialKey:      "lta_credentials",
			RetryCount:         3,
			TimeoutSeconds:     60,
			WaitAfterLoadMs:    1000,
			UseHeadlessBrowser: true,
		},
	}

	// Create the venue
	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to create venue: %v", err)
	}

	// Verify venue ID is set
	if venue.ID.IsZero() {
		t.Errorf("Venue ID should not be zero")
	}

	// Verify timestamps are set
	if venue.CreatedAt.IsZero() || venue.UpdatedAt.IsZero() {
		t.Errorf("Venue timestamps should not be zero")
	}
}

func TestVenueRepository_FindByID(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create a test venue
	venue := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub",
		Location: models.Location{
			Address:  "123 Tennis Court Road",
			City:     "London",
			PostCode: "SW19 5AE",
		},
		Courts: []models.Court{
			{
				ID:     "court1",
				Name:   "Court 1",
				Indoor: false,
			},
		},
		BookingWindow:    7,
		ScrapingInterval: 30,
		IsActive:         true,
	}

	// Create the venue
	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to create venue: %v", err)
	}

	// Find the venue by ID
	foundVenue, err := repo.FindByID(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Failed to find venue by ID: %v", err)
	}

	// Verify venue fields
	if foundVenue.Name != venue.Name {
		t.Errorf("Expected name %s, got %s", venue.Name, foundVenue.Name)
	}
	if foundVenue.Provider != venue.Provider {
		t.Errorf("Expected provider %s, got %s", venue.Provider, foundVenue.Provider)
	}
	if len(foundVenue.Courts) != len(venue.Courts) {
		t.Errorf("Expected %d courts, got %d", len(venue.Courts), len(foundVenue.Courts))
	}
}

func TestVenueRepository_FindByName(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create a test venue
	venue := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub",
		Location: models.Location{
			Address:  "123 Tennis Court Road",
			City:     "London",
			PostCode: "SW19 5AE",
		},
		BookingWindow:    7,
		ScrapingInterval: 30,
		IsActive:         true,
	}

	// Create the venue
	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to create venue: %v", err)
	}

	// Find the venue by name
	foundVenue, err := repo.FindByName(ctx, venue.Name)
	if err != nil {
		t.Fatalf("Failed to find venue by name: %v", err)
	}

	// Verify venue fields
	if foundVenue.ID != venue.ID {
		t.Errorf("Expected ID %s, got %s", venue.ID.Hex(), foundVenue.ID.Hex())
	}
	if foundVenue.Provider != venue.Provider {
		t.Errorf("Expected provider %s, got %s", venue.Provider, foundVenue.Provider)
	}
}

func TestVenueRepository_FindByProvider(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test venues
	venues := []*models.Venue{
		{
			Name:     "LTA Tennis Club 1",
			Provider: "lta",
			URL:      "https://clubspark.lta.org.uk/TestTennisClub1",
			IsActive: true,
		},
		{
			Name:     "LTA Tennis Club 2",
			Provider: "lta",
			URL:      "https://clubspark.lta.org.uk/TestTennisClub2",
			IsActive: true,
		},
		{
			Name:     "Courtsides Tennis Club",
			Provider: "courtsides",
			URL:      "https://courtsides.com/TestTennisClub",
			IsActive: true,
		},
	}

	// Create the venues
	for _, venue := range venues {
		err := repo.Create(ctx, venue)
		if err != nil {
			t.Fatalf("Failed to create venue: %v", err)
		}
	}

	// Find venues by provider
	ltaVenues, err := repo.FindByProvider(ctx, "lta")
	if err != nil {
		t.Fatalf("Failed to find venues by provider: %v", err)
	}

	// Verify venue count
	if len(ltaVenues) != 2 {
		t.Errorf("Expected 2 LTA venues, got %d", len(ltaVenues))
	}

	// Find courtsides venues
	courtsidesVenues, err := repo.FindByProvider(ctx, "courtsides")
	if err != nil {
		t.Fatalf("Failed to find venues by provider: %v", err)
	}

	// Verify venue count
	if len(courtsidesVenues) != 1 {
		t.Errorf("Expected 1 courtsides venue, got %d", len(courtsidesVenues))
	}
}

func TestVenueRepository_Update(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create a test venue
	venue := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub",
		Location: models.Location{
			Address:  "123 Tennis Court Road",
			City:     "London",
			PostCode: "SW19 5AE",
		},
		BookingWindow:    7,
		ScrapingInterval: 30,
		IsActive:         true,
	}

	// Create the venue
	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to create venue: %v", err)
	}

	// Update the venue
	originalUpdateTime := venue.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure updated time is different
	venue.Name = "Updated Tennis Club"
	venue.BookingWindow = 14
	venue.IsActive = false
	err = repo.Update(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to update venue: %v", err)
	}

	// Find the venue by ID to verify update
	foundVenue, err := repo.FindByID(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Failed to find venue by ID: %v", err)
	}

	// Verify venue fields
	if foundVenue.Name != "Updated Tennis Club" {
		t.Errorf("Expected updated name 'Updated Tennis Club', got %s", foundVenue.Name)
	}
	if foundVenue.BookingWindow != 14 {
		t.Errorf("Expected updated booking window 14, got %d", foundVenue.BookingWindow)
	}
	if foundVenue.IsActive != false {
		t.Errorf("Expected updated is_active false, got %v", foundVenue.IsActive)
	}
	if !foundVenue.UpdatedAt.After(originalUpdateTime) {
		t.Errorf("Expected updated time to be after original update time")
	}
}

func TestVenueRepository_UpdateLastScraped(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create a test venue
	venue := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub",
		IsActive: true,
	}

	// Create the venue
	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to create venue: %v", err)
	}

	// Update last scraped
	err = repo.UpdateLastScraped(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Failed to update last scraped: %v", err)
	}

	// Find the venue by ID to verify update
	foundVenue, err := repo.FindByID(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Failed to find venue by ID: %v", err)
	}

	// Verify last scraped is set
	if foundVenue.LastScrapedAt.IsZero() {
		t.Errorf("Expected last scraped time to be set")
	}
}

func TestVenueRepository_Delete(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create a test venue
	venue := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub",
		IsActive: true,
	}

	// Create the venue
	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to create venue: %v", err)
	}

	// Delete the venue
	err = repo.Delete(ctx, venue.ID)
	if err != nil {
		t.Fatalf("Failed to delete venue: %v", err)
	}

	// Try to find the venue by ID
	_, err = repo.FindByID(ctx, venue.ID)
	if err == nil {
		t.Errorf("Expected error when finding deleted venue, got nil")
	}
}

func TestVenueRepository_List(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create multiple test venues
	for i := 0; i < 5; i++ {
		venue := &models.Venue{
			Name:     "Test Tennis Club " + primitive.NewObjectID().Hex(),
			Provider: "lta",
			URL:      "https://clubspark.lta.org.uk/TestTennisClub" + primitive.NewObjectID().Hex(),
			IsActive: true,
		}
		err := repo.Create(ctx, venue)
		if err != nil {
			t.Fatalf("Failed to create venue: %v", err)
		}
	}

	// List all venues
	venues, err := repo.List(ctx, 0, 0)
	if err != nil {
		t.Fatalf("Failed to list venues: %v", err)
	}

	// Verify venue count
	if len(venues) != 5 {
		t.Errorf("Expected 5 venues, got %d", len(venues))
	}

	// Test pagination
	venues, err = repo.List(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Failed to list venues with pagination: %v", err)
	}

	// Verify paginated venue count
	if len(venues) != 2 {
		t.Errorf("Expected 2 venues with pagination, got %d", len(venues))
	}
}

func TestVenueRepository_ListActive(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test venues
	venues := []*models.Venue{
		{
			Name:     "Active Tennis Club 1",
			Provider: "lta",
			URL:      "https://clubspark.lta.org.uk/ActiveTennisClub1",
			IsActive: true,
		},
		{
			Name:     "Active Tennis Club 2",
			Provider: "lta",
			URL:      "https://clubspark.lta.org.uk/ActiveTennisClub2",
			IsActive: true,
		},
		{
			Name:     "Inactive Tennis Club",
			Provider: "lta",
			URL:      "https://clubspark.lta.org.uk/InactiveTennisClub",
			IsActive: false,
		},
	}

	// Create the venues
	for _, venue := range venues {
		err := repo.Create(ctx, venue)
		if err != nil {
			t.Fatalf("Failed to create venue: %v", err)
		}
	}

	// List active venues
	activeVenues, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatalf("Failed to list active venues: %v", err)
	}

	// Verify active venue count
	if len(activeVenues) != 2 {
		t.Errorf("Expected 2 active venues, got %d", len(activeVenues))
	}
}

func TestVenueRepository_CreateIndexes(t *testing.T) {
	_, db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create indexes
	err := repo.CreateIndexes(ctx)
	if err != nil {
		t.Fatalf("Failed to create indexes: %v", err)
	}

	// Create a test venue
	venue1 := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub",
		IsActive: true,
	}
	err = repo.Create(ctx, venue1)
	if err != nil {
		t.Fatalf("Failed to create venue: %v", err)
	}

	// Try to create another venue with the same name
	venue2 := &models.Venue{
		Name:     "Test Tennis Club",
		Provider: "lta",
		URL:      "https://clubspark.lta.org.uk/TestTennisClub2",
		IsActive: true,
	}
	err = repo.Create(ctx, venue2)
	if err == nil {
		t.Errorf("Expected error when creating venue with duplicate name, got nil")
	}
}
