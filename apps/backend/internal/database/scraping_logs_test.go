package database

import (
	"context"
	"os"
	"testing"
	"time"

	"tennis-booker/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupScrapingLogTest(t *testing.T) (*mongo.Database, *ScrapingLogRepository, func()) {
	// Skip integration tests if MongoDB is not available
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB integration tests - SKIP_MONGODB_TESTS=true")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	// Use a unique database name for each test to ensure isolation
	dbName := "test_db_" + primitive.NewObjectID().Hex()

	// Connect to MongoDB with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - failed to connect: %v", err)
	}

	// Ping the database with short timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		client.Disconnect(context.Background())
		t.Skipf("Skipping MongoDB integration tests - failed to ping: %v", err)
	}

	// Create a new database for testing
	db := client.Database(dbName)

	// Create repository
	repo := NewScrapingLogRepository(db)

	// Return cleanup function
	cleanup := func() {
		err := db.Drop(context.Background())
		assert.NoError(t, err)
		err = client.Disconnect(context.Background())
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
	assert.Equal(t, int64(4), deleted) // Should delete logs from 6 to 9 days ago (4 logs)

	// Verify count
	count, err := repo.CountByVenue(ctx, venueID)
	assert.NoError(t, err)
	assert.Equal(t, int64(6), count) // Should have 6 logs left (0 to 5 days ago)
}
