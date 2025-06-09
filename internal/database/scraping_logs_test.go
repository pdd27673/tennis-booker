package database

import (
	"context"
	"testing"
	"time"

	"tennis-booking-bot/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupScrapingLogTest(t *testing.T) (*mongo.Database, *ScrapingLogRepository, func()) {
	// Use a unique database name for each test to ensure isolation
	dbName := "test_db_" + primitive.NewObjectID().Hex()
	
	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	// Create a new database for testing
	db := client.Database(dbName)
	
	// Create repository
	repo := NewScrapingLogRepository(db)
	
	// Return cleanup function
	cleanup := func() {
		err := db.Drop(ctx)
		assert.NoError(t, err)
		err = client.Disconnect(ctx)
		assert.NoError(t, err)
	}

	return db, repo, cleanup
}

func TestScrapingLogRepository_Create(t *testing.T) {
	_, repo, cleanup := setupScrapingLogTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create a test log
	venueID := primitive.NewObjectID()
	log := &models.ScrapingLog{
		VenueID:         venueID,
		ScrapeTimestamp: time.Now(),
		SlotsFound: []models.Slot{
			{
				Date:      "2023-06-15",
				Time:      "10:00-11:00",
				Court:     "Court 1",
				Available: true,
				Price:     10.50,
			},
		},
		ScrapeDurationMs: 1500,
		Success:          true,
		VenueName:        "Test Venue",
		Provider:         "lta_clubspark",
	}

	// Test creating the log
	err := repo.Create(ctx, log)
	assert.NoError(t, err)
	assert.False(t, log.ID.IsZero(), "Expected ID to be set")

	// Verify we can retrieve it
	retrieved, err := repo.FindByID(ctx, log.ID)
	assert.NoError(t, err)
	assert.Equal(t, log.VenueID, retrieved.VenueID)
	assert.Equal(t, log.VenueName, retrieved.VenueName)
	assert.Equal(t, log.Provider, retrieved.Provider)
	assert.Equal(t, log.Success, retrieved.Success)
	assert.Equal(t, len(log.SlotsFound), len(retrieved.SlotsFound))
}

func TestScrapingLogRepository_FindByVenueID(t *testing.T) {
	_, repo, cleanup := setupScrapingLogTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create venue IDs
	venue1ID := primitive.NewObjectID()
	venue2ID := primitive.NewObjectID()

	// Create test logs for two different venues
	for i := 0; i < 5; i++ {
		log1 := &models.ScrapingLog{
			VenueID:         venue1ID,
			ScrapeTimestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Success:         true,
			VenueName:       "Venue 1",
		}
		err := repo.Create(ctx, log1)
		require.NoError(t, err)
		
		log2 := &models.ScrapingLog{
			VenueID:         venue2ID,
			ScrapeTimestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Success:         true,
			VenueName:       "Venue 2",
		}
		err = repo.Create(ctx, log2)
		require.NoError(t, err)
	}

	// Test retrieving logs for venue 1
	logs, err := repo.FindByVenueID(ctx, venue1ID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(logs))
	for _, log := range logs {
		assert.Equal(t, venue1ID, log.VenueID)
	}

	// Test pagination
	logs, err = repo.FindByVenueID(ctx, venue1ID, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(logs))
}

func TestScrapingLogRepository_FindByTimeRange(t *testing.T) {
	_, repo, cleanup := setupScrapingLogTest(t)
	defer cleanup()

	ctx := context.Background()
	venueID := primitive.NewObjectID()
	
	now := time.Now()
	
	// Create logs at different times
	for i := 0; i < 10; i++ {
		log := &models.ScrapingLog{
			VenueID:         venueID,
			ScrapeTimestamp: now.Add(-time.Duration(i) * time.Hour),
			Success:         true,
			VenueName:       "Test Venue",
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}
	
	// Test finding logs within time range (last 5 hours)
	startTime := now.Add(-5 * time.Hour)
	endTime := now
	
	logs, err := repo.FindByTimeRange(ctx, startTime, endTime, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(logs)) // Includes logs at now, -1h, -2h, -3h, -4h, -5h
}

func TestScrapingLogRepository_FindSuccessfulByVenue(t *testing.T) {
	_, repo, cleanup := setupScrapingLogTest(t)
	defer cleanup()

	ctx := context.Background()
	venueID := primitive.NewObjectID()
	
	// Create both successful and failed logs
	for i := 0; i < 5; i++ {
		log := &models.ScrapingLog{
			VenueID:         venueID,
			ScrapeTimestamp: time.Now(),
			Success:         true,
			VenueName:       "Test Venue",
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)
		
		log = &models.ScrapingLog{
			VenueID:         venueID,
			ScrapeTimestamp: time.Now(),
			Success:         false,
			VenueName:       "Test Venue",
			Errors:          []string{"Test error"},
		}
		err = repo.Create(ctx, log)
		require.NoError(t, err)
	}
	
	// Test finding only successful logs
	logs, err := repo.FindSuccessfulByVenue(ctx, venueID, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(logs))
	for _, log := range logs {
		assert.True(t, log.Success)
	}
}

func TestScrapingLogRepository_CountByVenue(t *testing.T) {
	_, repo, cleanup := setupScrapingLogTest(t)
	defer cleanup()

	ctx := context.Background()
	venue1ID := primitive.NewObjectID()
	venue2ID := primitive.NewObjectID()
	
	// Create logs for two venues
	for i := 0; i < 3; i++ {
		log := &models.ScrapingLog{
			VenueID:         venue1ID,
			ScrapeTimestamp: time.Now(),
			VenueName:       "Venue 1",
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}
	
	for i := 0; i < 5; i++ {
		log := &models.ScrapingLog{
			VenueID:         venue2ID,
			ScrapeTimestamp: time.Now(),
			VenueName:       "Venue 2",
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}
	
	// Test count
	count, err := repo.CountByVenue(ctx, venue1ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
	
	count, err = repo.CountByVenue(ctx, venue2ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestScrapingLogRepository_DeleteOlderThan(t *testing.T) {
	_, repo, cleanup := setupScrapingLogTest(t)
	defer cleanup()

	ctx := context.Background()
	venueID := primitive.NewObjectID()
	
	now := time.Now()
	
	// Create logs at different times
	for i := 0; i < 10; i++ {
		log := &models.ScrapingLog{
			VenueID:         venueID,
			ScrapeTimestamp: now.Add(-time.Duration(i) * 24 * time.Hour), // days ago
			Success:         true,
			VenueName:       "Test Venue",
		}
		err := repo.Create(ctx, log)
		require.NoError(t, err)
	}
	
	// Delete logs older than 5 days
	deleteCutoff := now.Add(-5 * 24 * time.Hour)
	deleted, err := repo.DeleteOlderThan(ctx, deleteCutoff)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), deleted) // Should delete logs from 5 to 9 days ago (5 logs)
	
	// Verify count
	count, err := repo.CountByVenue(ctx, venueID)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count) // Should have 5 logs left (0 to 4 days ago)
} 