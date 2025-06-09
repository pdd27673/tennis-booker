package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"tennis-booking-bot/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// Parse command line flags
	verify := flag.Bool("verify", false, "Only verify indexes without creating them")
	verbose := flag.Bool("verbose", false, "Show detailed index information")
	envFile := flag.String("env", ".env", "Path to .env file")
	flag.Parse()

	// Load environment variables
	err := godotenv.Load(*envFile)
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get MongoDB connection details from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:YOUR_PASSWORD@localhost:27017"
		log.Printf("MONGO_URI not set, using default: %s", mongoURI)
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "tennis_booking"
		log.Printf("MONGO_DB_NAME not set, using default: %s", dbName)
	}

	// Initialize database connection
	db, err := database.InitDatabase(mongoURI, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Create context with timeout (used in database operations)
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify or create indexes
	if *verify {
		// Only verify indexes
		log.Println("Verifying indexes...")
		summaries, err := database.GetIndexSummary(db)
		if err != nil {
			log.Fatalf("Failed to get index summary: %v", err)
		}

		// Group by collection
		collectionIndexes := make(map[string][]database.IndexSummary)
		for _, summary := range summaries {
			collectionIndexes[summary.Collection] = append(collectionIndexes[summary.Collection], summary)
		}

		// Print summary
		fmt.Println("\nIndex Summary:")
		fmt.Println("=============")
		
		for collection, indexes := range collectionIndexes {
			fmt.Printf("\nCollection: %s\n", collection)
			fmt.Printf("  Number of indexes: %d\n", len(indexes))
			
			if *verbose {
				fmt.Println("  Indexes:")
				for _, idx := range indexes {
					fmt.Printf("    - %s\n", idx.IndexName)
					
					// Print keys
					keys := []string{}
					for field, direction := range idx.Keys {
						dir := "asc"
						if direction == -1 {
							dir = "desc"
						}
						keys = append(keys, fmt.Sprintf("%s: %s", field, dir))
					}
					fmt.Printf("      Keys: %s\n", strings.Join(keys, ", "))
					
					// Print properties
					if idx.Unique {
						fmt.Println("      Unique: true")
					}
					if idx.TTL > 0 {
						fmt.Printf("      TTL: %d seconds\n", idx.TTL)
					}
				}
			}
		}

		// Check for required collections
		requiredCollections := []string{"users", "venues", "bookings", "scraping_logs"}
		for _, collection := range requiredCollections {
			if _, exists := collectionIndexes[collection]; !exists {
				fmt.Printf("\nWARNING: Required collection '%s' does not exist or has no indexes\n", collection)
			}
		}
	} else {
		// Create all indexes
		log.Println("Creating all indexes...")
		err = database.CreateAllIndexes(db)
		if err != nil {
			log.Fatalf("Failed to create indexes: %v", err)
		}
		log.Println("All indexes created successfully")

		// If verbose, show the created indexes
		if *verbose {
			summaries, err := database.GetIndexSummary(db)
			if err != nil {
				log.Fatalf("Failed to get index summary: %v", err)
			}

			// Group by collection
			collectionIndexes := make(map[string][]database.IndexSummary)
			for _, summary := range summaries {
				collectionIndexes[summary.Collection] = append(collectionIndexes[summary.Collection], summary)
			}

			// Print summary
			fmt.Println("\nCreated Indexes:")
			fmt.Println("===============")
			
			for collection, indexes := range collectionIndexes {
				fmt.Printf("\nCollection: %s\n", collection)
				fmt.Printf("  Number of indexes: %d\n", len(indexes))
				
				fmt.Println("  Indexes:")
				for _, idx := range indexes {
					fmt.Printf("    - %s\n", idx.IndexName)
					
					// Print keys
					keys := []string{}
					for field, direction := range idx.Keys {
						dir := "asc"
						if direction == -1 {
							dir = "desc"
						}
						keys = append(keys, fmt.Sprintf("%s: %s", field, dir))
					}
					fmt.Printf("      Keys: %s\n", strings.Join(keys, ", "))
					
					// Print properties
					if idx.Unique {
						fmt.Println("      Unique: true")
					}
					if idx.TTL > 0 {
						fmt.Printf("      TTL: %d seconds\n", idx.TTL)
					}
				}
			}
		}
	}
} 