package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexRecommendation represents a recommended index (from analysis)
type IndexRecommendation struct {
	Type            string                 `json:"type"`
	Keys            map[string]interface{} `json:"keys"`
	Reason          string                 `json:"reason"`
	Priority        string                 `json:"priority"`
	EstimatedImpact string                 `json:"estimated_impact"`
}

// CollectionAnalysis represents analysis results for a collection (from analysis)
type CollectionAnalysis struct {
	Name            string                `json:"name"`
	DocumentCount   int64                 `json:"document_count"`
	DataSize        int64                 `json:"data_size"`
	IndexSize       int64                 `json:"index_size"`
	Recommendations []IndexRecommendation `json:"recommendations"`
}

// DatabaseAnalysis represents the complete database analysis (from analysis)
type DatabaseAnalysis struct {
	DatabaseName string               `json:"database_name"`
	Timestamp    time.Time            `json:"timestamp"`
	Collections  []CollectionAnalysis `json:"collections"`
}

// OptimizationResult represents the result of applying an index
type OptimizationResult struct {
	Collection string `json:"collection"`
	IndexName  string `json:"index_name"`
	Keys       string `json:"keys"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Duration   string `json:"duration"`
	Priority   string `json:"priority"`
}

// OptimizationSummary represents the overall optimization results
type OptimizationSummary struct {
	Timestamp         time.Time            `json:"timestamp"`
	TotalIndexes      int                  `json:"total_indexes"`
	SuccessfulIndexes int                  `json:"successful_indexes"`
	FailedIndexes     int                  `json:"failed_indexes"`
	Results           []OptimizationResult `json:"results"`
	Duration          string               `json:"duration"`
}

func main() {
	log.Println("🚀 Starting MongoDB Index Optimization...")

	// Check for analysis file
	analysisFile := "mongodb_index_analysis.json"
	if len(os.Args) > 1 {
		analysisFile = os.Args[1]
	}

	// Load analysis results
	analysis, err := loadAnalysis(analysisFile)
	if err != nil {
		log.Fatalf("Failed to load analysis file %s: %v", analysisFile, err)
	}

	log.Printf("📊 Loaded analysis for database: %s", analysis.DatabaseName)

	// Get MongoDB connection details from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = analysis.DatabaseName
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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
	log.Printf("✅ Connected to MongoDB database: %s", dbName)

	// Apply optimizations
	startTime := time.Now()
	summary := applyOptimizations(ctx, db, analysis)
	summary.Duration = time.Since(startTime).String()

	// Save results
	resultsFile := "mongodb_optimization_results.json"
	if err := saveResults(summary, resultsFile); err != nil {
		log.Printf("⚠️ Failed to save results: %v", err)
	}

	// Print summary
	printOptimizationSummary(summary)

	log.Printf("✅ Optimization complete. Results saved to: %s", resultsFile)
}

func loadAnalysis(filename string) (*DatabaseAnalysis, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var analysis DatabaseAnalysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &analysis, nil
}

func applyOptimizations(ctx context.Context, db *mongo.Database, analysis *DatabaseAnalysis) *OptimizationSummary {
	summary := &OptimizationSummary{
		Timestamp: time.Now(),
		Results:   []OptimizationResult{},
	}

	// Process each collection
	for _, collAnalysis := range analysis.Collections {
		log.Printf("📁 Processing collection: %s", collAnalysis.Name)

		collection := db.Collection(collAnalysis.Name)

		// Apply recommendations by priority (High first, then Medium, then Low)
		priorities := []string{"High", "Medium", "Low"}

		for _, priority := range priorities {
			for _, rec := range collAnalysis.Recommendations {
				if rec.Priority != priority {
					continue
				}

				summary.TotalIndexes++
				result := applyIndexRecommendation(ctx, collection, rec)
				summary.Results = append(summary.Results, result)

				if result.Success {
					summary.SuccessfulIndexes++
					log.Printf("   ✅ [%s] Created index: %s", priority, result.Keys)
				} else {
					summary.FailedIndexes++
					log.Printf("   ❌ [%s] Failed to create index %s: %s", priority, result.Keys, result.Error)
				}
			}
		}
	}

	return summary
}

func applyIndexRecommendation(ctx context.Context, collection *mongo.Collection, rec IndexRecommendation) OptimizationResult {
	startTime := time.Now()

	result := OptimizationResult{
		Collection: collection.Name(),
		Keys:       formatIndexKeys(rec.Keys),
		Priority:   rec.Priority,
	}

	// Create index model
	indexModel := mongo.IndexModel{}

	// Convert keys to bson.D
	var keys bson.D
	for key, value := range rec.Keys {
		keys = append(keys, bson.E{Key: key, Value: value})
	}
	indexModel.Keys = keys

	// Set index options based on type and recommendation
	opts := options.Index()

	// Generate index name
	indexName := generateIndexName(rec.Keys)
	opts.SetName(indexName)
	result.IndexName = indexName

	// Set TTL if this is a TTL index
	if rec.Type == "TTL" {
		// Set appropriate TTL based on collection and field
		ttlSeconds := getTTLSeconds(collection.Name(), rec.Keys)
		if ttlSeconds > 0 {
			opts.SetExpireAfterSeconds(int32(ttlSeconds))
			log.Printf("   🕐 Setting TTL: %d seconds for %s", ttlSeconds, indexName)
		}
	}

	// Set sparse for certain indexes
	if shouldBeSparse(rec.Keys) {
		opts.SetSparse(true)
		log.Printf("   🔍 Setting sparse: true for %s", indexName)
	}

	indexModel.Options = opts

	// Create the index
	_, err := collection.Indexes().CreateOne(ctx, indexModel)

	result.Duration = time.Since(startTime).String()

	if err != nil {
		result.Success = false
		result.Error = err.Error()

		// Check if error is because index already exists
		if isIndexExistsError(err) {
			result.Error = "Index already exists (skipped)"
			log.Printf("   ℹ️ Index %s already exists, skipping", indexName)
		}
	} else {
		result.Success = true
		log.Printf("   ⏱️ Created index %s in %s", indexName, result.Duration)
	}

	return result
}

func generateIndexName(keys map[string]interface{}) string {
	var parts []string
	for key, value := range keys {
		parts = append(parts, fmt.Sprintf("%s_%v", key, value))
	}
	return fmt.Sprintf("opt_%s", joinStrings(parts, "_"))
}

func formatIndexKeys(keys map[string]interface{}) string {
	var parts []string
	for key, value := range keys {
		parts = append(parts, fmt.Sprintf("%s: %v", key, value))
	}
	return fmt.Sprintf("{%s}", joinStrings(parts, ", "))
}

func getTTLSeconds(collectionName string, keys map[string]interface{}) int32 {
	// Define TTL values for different collections and fields
	ttlMap := map[string]map[string]int32{
		"court_slots": {
			"slot_date": 7 * 24 * 3600, // 7 days for old slots
		},
		"slots": {
			"slot_date": 7 * 24 * 3600, // 7 days for old slots
		},
		"scraping_logs": {
			"scrape_timestamp": 30 * 24 * 3600, // 30 days for logs
		},
		"deduplication_records": {
			"last_sent_at": 7 * 24 * 3600, // 7 days for deduplication records
		},
		"notifications": {
			"created_at": 90 * 24 * 3600, // 90 days for old notifications
		},
	}

	if collTTL, exists := ttlMap[collectionName]; exists {
		for field := range keys {
			if ttl, fieldExists := collTTL[field]; fieldExists {
				return ttl
			}
		}
	}

	return 0 // No TTL
}

func shouldBeSparse(keys map[string]interface{}) bool {
	// Fields that should have sparse indexes
	sparseFields := map[string]bool{
		"notification_settings.unsubscribed": true,
		"preferred_venues":                   true,
		"preferred_days":                     true,
		"excluded_venues":                    true,
		"booking_url":                        true,
		"notified":                           true,
	}

	for field := range keys {
		if sparseFields[field] {
			return true
		}
	}

	return false
}

func isIndexExistsError(err error) bool {
	// Check if the error indicates the index already exists
	errStr := err.Error()
	return contains(errStr, "already exists") ||
		contains(errStr, "IndexOptionsConflict") ||
		contains(errStr, "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func saveResults(summary *OptimizationSummary, filename string) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func printOptimizationSummary(summary *OptimizationSummary) {
	fmt.Println("\n" + repeatString("=", 60))
	fmt.Println("🚀 MONGODB INDEX OPTIMIZATION SUMMARY")
	fmt.Println(repeatString("=", 60))

	fmt.Printf("Optimization Time: %s\n", summary.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Duration: %s\n", summary.Duration)
	fmt.Printf("Total Indexes Processed: %d\n", summary.TotalIndexes)
	fmt.Printf("Successful: %d\n", summary.SuccessfulIndexes)
	fmt.Printf("Failed: %d\n", summary.FailedIndexes)

	if summary.TotalIndexes > 0 {
		successRate := float64(summary.SuccessfulIndexes) / float64(summary.TotalIndexes) * 100
		fmt.Printf("Success Rate: %.1f%%\n", successRate)
	}

	fmt.Println("\n" + repeatString("=", 60))
	fmt.Println("📋 DETAILED RESULTS")
	fmt.Println(repeatString("=", 60))

	// Group results by collection
	collectionResults := make(map[string][]OptimizationResult)
	for _, result := range summary.Results {
		collectionResults[result.Collection] = append(collectionResults[result.Collection], result)
	}

	for collection, results := range collectionResults {
		fmt.Printf("\n📁 %s:\n", collection)

		for _, result := range results {
			status := "✅"
			if !result.Success {
				status = "❌"
			}

			fmt.Printf("   %s [%s] %s (%s)\n", status, result.Priority, result.Keys, result.Duration)
			if !result.Success && result.Error != "" {
				fmt.Printf("      Error: %s\n", result.Error)
			}
		}
	}

	fmt.Println(repeatString("=", 60))
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
