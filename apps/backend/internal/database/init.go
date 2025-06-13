package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// InitDatabase initializes the MongoDB connection and returns a database instance
func InitDatabase(uri, dbName string) (*mongo.Database, error) {
	// Create a context with timeout for the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	// Get database
	db := client.Database(dbName)
	return db, nil
}

// CreateAllIndexes creates all necessary indexes for all collections
func CreateAllIndexes(db *mongo.Database) error {
	// Create a context with timeout for index creation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize repositories
	userRepo := NewUserRepository(db)
	venueRepo := NewVenueRepository(db)
	bookingRepo := NewBookingRepository(db)
	scrapingLogRepo := NewScrapingLogRepository(db)

	// Create indexes for each collection
	log.Println("Creating indexes for users collection...")
	if err := userRepo.CreateIndexes(ctx); err != nil {
		return err
	}

	log.Println("Creating indexes for venues collection...")
	if err := venueRepo.CreateIndexes(ctx); err != nil {
		return err
	}

	log.Println("Creating indexes for bookings collection...")
	if err := bookingRepo.CreateIndexes(ctx); err != nil {
		return err
	}

	log.Println("Creating indexes for scraping_logs collection...")
	if err := scrapingLogRepo.CreateIndexes(ctx); err != nil {
		return err
	}

	log.Println("All indexes created successfully")
	return nil
}

// IndexSummary represents information about an index
type IndexSummary struct {
	Collection string
	IndexName  string
	Keys       map[string]int
	Unique     bool
	TTL        int
}

// GetIndexSummary returns a summary of all indexes in the database
func GetIndexSummary(db *mongo.Database) ([]IndexSummary, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all collection names
	collections, err := db.ListCollectionNames(ctx, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var summaries []IndexSummary

	// For each collection, get its indexes
	for _, collName := range collections {
		coll := db.Collection(collName)
		
		// Get all indexes for the collection
		cursor, err := coll.Indexes().List(ctx)
		if err != nil {
			return nil, err
		}

		// Iterate through indexes and add to summary
		var indexes []map[string]interface{}
		if err = cursor.All(ctx, &indexes); err != nil {
			return nil, err
		}

		for _, idx := range indexes {
			// Extract index information
			name := idx["name"].(string)
			keys := make(map[string]int)
			
			// Extract key information
			if keyDoc, ok := idx["key"].(map[string]interface{}); ok {
				for k, v := range keyDoc {
					// Convert value to int (1 for ascending, -1 for descending)
					if vFloat, ok := v.(float64); ok {
						keys[k] = int(vFloat)
					}
				}
			}
			
			// Check if index is unique
			unique := false
			if u, ok := idx["unique"].(bool); ok {
				unique = u
			}
			
			// Check if index has TTL
			ttl := 0
			if expireAfterSeconds, ok := idx["expireAfterSeconds"].(float64); ok {
				ttl = int(expireAfterSeconds)
			}
			
			// Add to summaries
			summaries = append(summaries, IndexSummary{
				Collection: collName,
				IndexName:  name,
				Keys:       keys,
				Unique:     unique,
				TTL:        ttl,
			})
		}
	}

	return summaries, nil
} 