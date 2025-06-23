package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestInitDatabase(t *testing.T) {
	// Skip integration tests if MongoDB is not available
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB integration tests - SKIP_MONGODB_TESTS=true")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	// Test with valid URI - use short timeout to fail fast
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to connect first to see if MongoDB is available
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - failed to connect: %v", err)
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		client.Disconnect(context.Background())
		t.Skipf("Skipping MongoDB integration tests - failed to ping: %v", err)
	}
	client.Disconnect(context.Background())

	// Now test the actual InitDatabase function
	db, err := InitDatabase(mongoURI, "test_db")
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Test with invalid URI
	db, err = InitDatabase("mongodb://invalid:27017", "test_db")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestCreateAllIndexes(t *testing.T) {
	// Skip integration tests if MongoDB is not available
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB integration tests - SKIP_MONGODB_TESTS=true")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	// Connect to MongoDB with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - failed to connect: %v", err)
	}
	defer client.Disconnect(context.Background())

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - failed to ping: %v", err)
	}

	// Create a unique database name for this test
	dbName := "test_db_indexes"
	db := client.Database(dbName)

	// Create all indexes
	err = CreateAllIndexes(db)
	assert.NoError(t, err)

	// Get index summary
	summaries, err := GetIndexSummary(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, summaries)

	// Verify indexes for each collection
	collections := map[string]bool{
		"users":         false,
		"venues":        false,
		"bookings":      false,
		"scraping_logs": false,
	}

	// Check that each collection has at least one index (the _id index)
	for _, summary := range summaries {
		collections[summary.Collection] = true
	}

	// Check that all collections have indexes
	for collection, hasIndexes := range collections {
		assert.True(t, hasIndexes, "Collection %s should have indexes", collection)
	}

	// Clean up - drop the test database
	err = client.Database(dbName).Drop(context.Background())
	assert.NoError(t, err)
}

func TestGetIndexSummary(t *testing.T) {
	// Skip integration tests if MongoDB is not available
	if os.Getenv("SKIP_MONGODB_TESTS") == "true" {
		t.Skip("Skipping MongoDB integration tests - SKIP_MONGODB_TESTS=true")
	}

	mongoURI := os.Getenv("MONGODB_TEST_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	// Connect to MongoDB with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - failed to connect: %v", err)
	}
	defer client.Disconnect(context.Background())

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()

	err = client.Ping(pingCtx, nil)
	if err != nil {
		t.Skipf("Skipping MongoDB integration tests - failed to ping: %v", err)
	}

	// Create a unique database name for this test
	dbName := "test_db_summary"
	db := client.Database(dbName)

	// Create a collection with an index
	coll := db.Collection("test_collection")
	indexModel := mongo.IndexModel{
		Keys: map[string]interface{}{
			"test_field": 1,
		},
		Options: options.Index().SetUnique(true),
	}
	_, err = coll.Indexes().CreateOne(context.Background(), indexModel)
	require.NoError(t, err)

	// Get index summary
	summaries, err := GetIndexSummary(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, summaries)

	// Debug: Print all indexes found
	t.Logf("Found %d index summaries:", len(summaries))
	for _, summary := range summaries {
		t.Logf("Collection: %s, Index: %s, Keys: %+v, Unique: %t",
			summary.Collection, summary.IndexName, summary.Keys, summary.Unique)
	}

	// Verify the index was found
	found := false
	for _, summary := range summaries {
		if summary.Collection == "test_collection" {
			t.Logf("Found test_collection index: %s with keys %+v", summary.IndexName, summary.Keys)
			if _, ok := summary.Keys["test_field"]; ok {
				found = true
				assert.True(t, summary.Unique)
				break
			}
		}
	}
	assert.True(t, found, "Index on test_field should be found")

	// Clean up - drop the test database
	err = client.Database(dbName).Drop(context.Background())
	assert.NoError(t, err)
}
