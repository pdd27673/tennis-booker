package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestInitDatabase(t *testing.T) {
	// Skip this test if MongoDB is not available
	if testing.Short() {
		t.Skip("Skipping MongoDB test in short mode")
	}

	// Test with valid URI
	db, err := InitDatabase("mongodb://localhost:27017", "test_db")
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Test with invalid URI
	db, err = InitDatabase("mongodb://invalid:27017", "test_db")
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestCreateAllIndexes(t *testing.T) {
	// Skip this test if MongoDB is not available
	if testing.Short() {
		t.Skip("Skipping MongoDB test in short mode")
	}

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

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
	err = client.Database(dbName).Drop(ctx)
	assert.NoError(t, err)
}

func TestGetIndexSummary(t *testing.T) {
	// Skip this test if MongoDB is not available
	if testing.Short() {
		t.Skip("Skipping MongoDB test in short mode")
	}

	// Connect to MongoDB
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)
	defer client.Disconnect(ctx)

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
	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	require.NoError(t, err)

	// Get index summary
	summaries, err := GetIndexSummary(db)
	assert.NoError(t, err)
	assert.NotEmpty(t, summaries)

	// Verify the index was found
	found := false
	for _, summary := range summaries {
		if summary.Collection == "test_collection" && len(summary.Keys) > 0 {
			if _, ok := summary.Keys["test_field"]; ok {
				found = true
				assert.True(t, summary.Unique)
				break
			}
		}
	}
	assert.True(t, found, "Index on test_field should be found")

	// Clean up - drop the test database
	err = client.Database(dbName).Drop(ctx)
	assert.NoError(t, err)
} 