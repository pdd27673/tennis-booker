package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexInfo represents information about a MongoDB index
type IndexInfo struct {
	Name      string                 `json:"name"`
	Keys      map[string]interface{} `json:"keys"`
	Unique    bool                   `json:"unique,omitempty"`
	Sparse    bool                   `json:"sparse,omitempty"`
	TTL       *int32                 `json:"ttl,omitempty"`
	Size      int64                  `json:"size"`
	UsageInfo *IndexUsageInfo        `json:"usage_info,omitempty"`
}

// IndexUsageInfo represents index usage statistics
type IndexUsageInfo struct {
	Operations int64     `json:"operations"`
	Since      time.Time `json:"since"`
}

// CollectionAnalysis represents analysis results for a collection
type CollectionAnalysis struct {
	Name            string                `json:"name"`
	DocumentCount   int64                 `json:"document_count"`
	DataSize        int64                 `json:"data_size"`
	IndexSize       int64                 `json:"index_size"`
	Indexes         []IndexInfo           `json:"indexes"`
	QueryPatterns   []QueryPattern        `json:"query_patterns"`
	Recommendations []IndexRecommendation `json:"recommendations"`
}

// QueryPattern represents a common query pattern
type QueryPattern struct {
	Description string                 `json:"description"`
	Filter      map[string]interface{} `json:"filter"`
	Sort        map[string]interface{} `json:"sort,omitempty"`
	Projection  map[string]interface{} `json:"projection,omitempty"`
	Frequency   string                 `json:"frequency"`
	Performance string                 `json:"performance"`
}

// IndexRecommendation represents a recommended index
type IndexRecommendation struct {
	Type            string                 `json:"type"`
	Keys            map[string]interface{} `json:"keys"`
	Reason          string                 `json:"reason"`
	Priority        string                 `json:"priority"`
	EstimatedImpact string                 `json:"estimated_impact"`
}

// DatabaseAnalysis represents the complete database analysis
type DatabaseAnalysis struct {
	DatabaseName string               `json:"database_name"`
	Timestamp    time.Time            `json:"timestamp"`
	Collections  []CollectionAnalysis `json:"collections"`
	Summary      AnalysisSummary      `json:"summary"`
}

// AnalysisSummary provides high-level analysis results
type AnalysisSummary struct {
	TotalCollections  int                   `json:"total_collections"`
	TotalIndexes      int                   `json:"total_indexes"`
	TotalDataSize     int64                 `json:"total_data_size"`
	TotalIndexSize    int64                 `json:"total_index_size"`
	UnusedIndexes     []string              `json:"unused_indexes"`
	MissingIndexes    []IndexRecommendation `json:"missing_indexes"`
	PerformanceIssues []string              `json:"performance_issues"`
}

func main() {
	log.Println("🔍 Starting MongoDB Index Analysis...")

	// Get MongoDB connection details from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password@localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "tennis_booking"
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	// Perform analysis
	analysis, err := analyzeDatabase(ctx, db)
	if err != nil {
		log.Fatalf("Failed to analyze database: %v", err)
	}

	// Output results
	outputFile := "mongodb_index_analysis.json"
	if len(os.Args) > 1 {
		outputFile = os.Args[1]
	}

	if err := saveAnalysis(analysis, outputFile); err != nil {
		log.Fatalf("Failed to save analysis: %v", err)
	}

	// Print summary
	printSummary(analysis)

	log.Printf("✅ Analysis complete. Results saved to: %s", outputFile)
}

func analyzeDatabase(ctx context.Context, db *mongo.Database) (*DatabaseAnalysis, error) {
	analysis := &DatabaseAnalysis{
		DatabaseName: db.Name(),
		Timestamp:    time.Now(),
		Collections:  []CollectionAnalysis{},
	}

	// Get list of collections
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	log.Printf("📊 Analyzing %d collections...", len(collections))

	// Analyze each collection
	for _, collName := range collections {
		log.Printf("   Analyzing collection: %s", collName)

		collAnalysis, err := analyzeCollection(ctx, db.Collection(collName))
		if err != nil {
			log.Printf("   ⚠️ Failed to analyze collection %s: %v", collName, err)
			continue
		}

		analysis.Collections = append(analysis.Collections, *collAnalysis)
	}

	// Generate summary
	analysis.Summary = generateSummary(analysis.Collections)

	return analysis, nil
}

func analyzeCollection(ctx context.Context, collection *mongo.Collection) (*CollectionAnalysis, error) {
	collName := collection.Name()

	analysis := &CollectionAnalysis{
		Name:            collName,
		Indexes:         []IndexInfo{},
		QueryPatterns:   getQueryPatterns(collName),
		Recommendations: []IndexRecommendation{},
	}

	// Get collection stats
	var stats bson.M
	err := collection.Database().RunCommand(ctx, bson.D{
		{Key: "collStats", Value: collName},
	}).Decode(&stats)

	if err == nil {
		if count, ok := stats["count"].(int64); ok {
			analysis.DocumentCount = count
		} else if count, ok := stats["count"].(int32); ok {
			analysis.DocumentCount = int64(count)
		}

		if size, ok := stats["size"].(int64); ok {
			analysis.DataSize = size
		} else if size, ok := stats["size"].(int32); ok {
			analysis.DataSize = int64(size)
		}

		if indexSize, ok := stats["totalIndexSize"].(int64); ok {
			analysis.IndexSize = indexSize
		} else if indexSize, ok := stats["totalIndexSize"].(int32); ok {
			analysis.IndexSize = int64(indexSize)
		}
	}

	// Get index information
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var indexDoc bson.M
		if err := cursor.Decode(&indexDoc); err != nil {
			continue
		}

		indexInfo := parseIndexInfo(indexDoc)

		// Try to get index usage stats (MongoDB 3.2+)
		if usage := getIndexUsage(ctx, collection, indexInfo.Name); usage != nil {
			indexInfo.UsageInfo = usage
		}

		analysis.Indexes = append(analysis.Indexes, indexInfo)
	}

	// Generate recommendations
	analysis.Recommendations = generateRecommendations(analysis)

	return analysis, nil
}

func parseIndexInfo(indexDoc bson.M) IndexInfo {
	info := IndexInfo{}

	if name, ok := indexDoc["name"].(string); ok {
		info.Name = name
	}

	if keys, ok := indexDoc["key"].(bson.M); ok {
		info.Keys = make(map[string]interface{})
		for k, v := range keys {
			info.Keys[k] = v
		}
	}

	if unique, ok := indexDoc["unique"].(bool); ok {
		info.Unique = unique
	}

	if sparse, ok := indexDoc["sparse"].(bool); ok {
		info.Sparse = sparse
	}

	if ttl, ok := indexDoc["expireAfterSeconds"].(int32); ok {
		info.TTL = &ttl
	}

	return info
}

func getIndexUsage(ctx context.Context, collection *mongo.Collection, indexName string) *IndexUsageInfo {
	// Try to get index stats (requires appropriate permissions)
	pipeline := mongo.Pipeline{
		{{Key: "$indexStats", Value: bson.M{}}},
		{{Key: "$match", Value: bson.M{"name": indexName}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var stats bson.M
		if err := cursor.Decode(&stats); err == nil {
			if accesses, ok := stats["accesses"].(bson.M); ok {
				usage := &IndexUsageInfo{}

				if ops, ok := accesses["ops"].(int64); ok {
					usage.Operations = ops
				} else if ops, ok := accesses["ops"].(int32); ok {
					usage.Operations = int64(ops)
				}

				if since, ok := accesses["since"].(primitive.DateTime); ok {
					usage.Since = since.Time()
				}

				return usage
			}
		}
	}

	return nil
}

func getQueryPatterns(collectionName string) []QueryPattern {
	patterns := []QueryPattern{}

	switch collectionName {
	case "court_slots", "slots":
		patterns = []QueryPattern{
			{
				Description: "Find available slots by venue and date range",
				Filter: map[string]interface{}{
					"venue_id":  "ObjectId",
					"available": true,
					"slot_date": map[string]interface{}{"$gte": "date", "$lte": "date"},
				},
				Sort:        map[string]interface{}{"slot_date": 1, "start_time": 1},
				Frequency:   "Very High",
				Performance: "Critical",
			},
			{
				Description: "Find slots by venue, date, and time for booking",
				Filter: map[string]interface{}{
					"venue_id":   "ObjectId",
					"date":       "string",
					"start_time": "string",
				},
				Frequency:   "High",
				Performance: "Critical",
			},
			{
				Description: "Find recent slots for notifications",
				Filter: map[string]interface{}{
					"available":    true,
					"last_scraped": map[string]interface{}{"$gte": "recent_date"},
					"notified":     map[string]interface{}{"$ne": true},
				},
				Sort:        map[string]interface{}{"last_scraped": -1},
				Frequency:   "High",
				Performance: "Important",
			},
			{
				Description: "Cleanup old unavailable slots",
				Filter: map[string]interface{}{
					"available": false,
					"slot_date": map[string]interface{}{"$lt": "cutoff_date"},
				},
				Frequency:   "Daily",
				Performance: "Moderate",
			},
		}

	case "user_preferences":
		patterns = []QueryPattern{
			{
				Description: "Get user preferences by user ID",
				Filter: map[string]interface{}{
					"user_id": "ObjectId",
				},
				Frequency:   "Very High",
				Performance: "Critical",
			},
			{
				Description: "Find active preferences for notifications",
				Filter: map[string]interface{}{
					"notification_settings.unsubscribed": map[string]interface{}{"$ne": true},
					"$or": []interface{}{
						map[string]interface{}{"preferred_venues": map[string]interface{}{"$exists": true, "$ne": []interface{}{}}},
						map[string]interface{}{"preferred_days": map[string]interface{}{"$exists": true, "$ne": []interface{}{}}},
						map[string]interface{}{"max_price": map[string]interface{}{"$gt": 0}},
					},
				},
				Frequency:   "High",
				Performance: "Important",
			},
			{
				Description: "Find preferences by venue for targeted notifications",
				Filter: map[string]interface{}{
					"preferred_venues":                   "venue_id",
					"notification_settings.unsubscribed": map[string]interface{}{"$ne": true},
				},
				Frequency:   "High",
				Performance: "Important",
			},
		}

	case "venues":
		patterns = []QueryPattern{
			{
				Description: "Find active venues for scraping",
				Filter: map[string]interface{}{
					"is_active": true,
				},
				Frequency:   "High",
				Performance: "Important",
			},
			{
				Description: "Find venue by name",
				Filter: map[string]interface{}{
					"name": "string",
				},
				Frequency:   "Medium",
				Performance: "Important",
			},
			{
				Description: "Find venues by provider",
				Filter: map[string]interface{}{
					"provider": "string",
				},
				Frequency:   "Medium",
				Performance: "Moderate",
			},
		}

	case "scraping_logs":
		patterns = []QueryPattern{
			{
				Description: "Find recent logs by venue",
				Filter: map[string]interface{}{
					"venue_id":         "ObjectId",
					"scrape_timestamp": map[string]interface{}{"$gte": "recent_date"},
				},
				Sort:        map[string]interface{}{"scrape_timestamp": -1},
				Frequency:   "High",
				Performance: "Important",
			},
			{
				Description: "Find successful scrapes with slots",
				Filter: map[string]interface{}{
					"success":          true,
					"slots_found":      map[string]interface{}{"$gt": 0},
					"scrape_timestamp": map[string]interface{}{"$gte": "recent_date"},
				},
				Sort:        map[string]interface{}{"scrape_timestamp": -1},
				Frequency:   "Medium",
				Performance: "Important",
			},
			{
				Description: "Cleanup old logs",
				Filter: map[string]interface{}{
					"scrape_timestamp": map[string]interface{}{"$lt": "cutoff_date"},
				},
				Frequency:   "Daily",
				Performance: "Moderate",
			},
		}

	case "deduplication_records":
		patterns = []QueryPattern{
			{
				Description: "Find exact slot match for user",
				Filter: map[string]interface{}{
					"user_id":  "ObjectId",
					"slot_key": "string",
				},
				Frequency:   "Very High",
				Performance: "Critical",
			},
			{
				Description: "Find similar notifications for user",
				Filter: map[string]interface{}{
					"user_id":         "ObjectId",
					"venue_id":        "ObjectId",
					"court_id":        "string",
					"slot_start_time": "string",
					"last_sent_at":    map[string]interface{}{"$gte": "recent_date"},
				},
				Sort:        map[string]interface{}{"last_sent_at": -1},
				Frequency:   "High",
				Performance: "Critical",
			},
		}

	case "notifications":
		patterns = []QueryPattern{
			{
				Description: "Find user notifications",
				Filter: map[string]interface{}{
					"user_id": "ObjectId",
				},
				Sort:        map[string]interface{}{"created_at": -1},
				Frequency:   "High",
				Performance: "Important",
			},
			{
				Description: "Find unread notifications",
				Filter: map[string]interface{}{
					"user_id": "ObjectId",
					"read":    false,
				},
				Sort:        map[string]interface{}{"created_at": -1},
				Frequency:   "Medium",
				Performance: "Important",
			},
		}
	}

	return patterns
}

func generateRecommendations(analysis *CollectionAnalysis) []IndexRecommendation {
	recommendations := []IndexRecommendation{}
	existingIndexes := make(map[string]bool)

	// Build map of existing indexes
	for _, index := range analysis.Indexes {
		key := indexKeysToString(index.Keys)
		existingIndexes[key] = true
	}

	// Generate recommendations based on query patterns
	for _, pattern := range analysis.QueryPatterns {
		rec := generateIndexRecommendation(pattern, existingIndexes)
		if rec != nil {
			recommendations = append(recommendations, *rec)
		}
	}

	// Add collection-specific recommendations
	switch analysis.Name {
	case "court_slots", "slots":
		addSlotIndexRecommendations(&recommendations, existingIndexes)
	case "user_preferences":
		addPreferenceIndexRecommendations(&recommendations, existingIndexes)
	case "scraping_logs":
		addScrapingLogIndexRecommendations(&recommendations, existingIndexes)
	case "deduplication_records":
		addDeduplicationIndexRecommendations(&recommendations, existingIndexes)
	}

	return recommendations
}

func generateIndexRecommendation(pattern QueryPattern, existingIndexes map[string]bool) *IndexRecommendation {
	// Extract fields from filter
	fields := extractFieldsFromFilter(pattern.Filter)

	// Add sort fields
	if pattern.Sort != nil {
		for field := range pattern.Sort {
			if !contains(fields, field) {
				fields = append(fields, field)
			}
		}
	}

	if len(fields) == 0 {
		return nil
	}

	// Create compound index keys
	keys := make(map[string]interface{})
	for _, field := range fields {
		if pattern.Sort != nil && pattern.Sort[field] != nil {
			keys[field] = pattern.Sort[field]
		} else {
			keys[field] = 1
		}
	}

	keyStr := indexKeysToString(keys)
	if existingIndexes[keyStr] {
		return nil // Index already exists
	}

	priority := "Medium"
	if pattern.Performance == "Critical" {
		priority = "High"
	} else if pattern.Performance == "Moderate" {
		priority = "Low"
	}

	return &IndexRecommendation{
		Type:            "Compound",
		Keys:            keys,
		Reason:          fmt.Sprintf("Optimize query: %s", pattern.Description),
		Priority:        priority,
		EstimatedImpact: estimateImpact(pattern.Frequency, pattern.Performance),
	}
}

func addSlotIndexRecommendations(recommendations *[]IndexRecommendation, existingIndexes map[string]bool) {
	// Critical compound index for slot queries
	if !existingIndexes["venue_id_1_slot_date_1_start_time_1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "Compound",
			Keys: map[string]interface{}{
				"venue_id":   1,
				"slot_date":  1,
				"start_time": 1,
			},
			Reason:          "Critical for venue-date-time slot lookups",
			Priority:        "High",
			EstimatedImpact: "Very High - Core booking functionality",
		})
	}

	// Index for availability queries
	if !existingIndexes["available_1_last_scraped_-1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "Compound",
			Keys: map[string]interface{}{
				"available":    1,
				"last_scraped": -1,
			},
			Reason:          "Optimize notification queries for new available slots",
			Priority:        "High",
			EstimatedImpact: "High - Notification performance",
		})
	}

	// TTL index for cleanup
	if !existingIndexes["slot_date_1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "TTL",
			Keys: map[string]interface{}{
				"slot_date": 1,
			},
			Reason:          "Automatic cleanup of old slots",
			Priority:        "Medium",
			EstimatedImpact: "Medium - Reduces storage and improves performance",
		})
	}
}

func addPreferenceIndexRecommendations(recommendations *[]IndexRecommendation, existingIndexes map[string]bool) {
	// Compound index for notification matching
	if !existingIndexes["notification_settings.unsubscribed_1_preferred_venues_1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "Compound",
			Keys: map[string]interface{}{
				"notification_settings.unsubscribed": 1,
				"preferred_venues":                   1,
			},
			Reason:          "Optimize venue-based notification targeting",
			Priority:        "High",
			EstimatedImpact: "High - Notification matching performance",
		})
	}
}

func addScrapingLogIndexRecommendations(recommendations *[]IndexRecommendation, existingIndexes map[string]bool) {
	// Compound index for venue logs
	if !existingIndexes["venue_id_1_scrape_timestamp_-1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "Compound",
			Keys: map[string]interface{}{
				"venue_id":         1,
				"scrape_timestamp": -1,
			},
			Reason:          "Optimize venue scraping history queries",
			Priority:        "Medium",
			EstimatedImpact: "Medium - Monitoring and debugging",
		})
	}

	// TTL index for log cleanup
	if !existingIndexes["scrape_timestamp_1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "TTL",
			Keys: map[string]interface{}{
				"scrape_timestamp": 1,
			},
			Reason:          "Automatic cleanup of old scraping logs",
			Priority:        "Medium",
			EstimatedImpact: "Medium - Storage management",
		})
	}
}

func addDeduplicationIndexRecommendations(recommendations *[]IndexRecommendation, existingIndexes map[string]bool) {
	// Compound index for exact matches
	if !existingIndexes["user_id_1_slot_key_1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "Compound",
			Keys: map[string]interface{}{
				"user_id":  1,
				"slot_key": 1,
			},
			Reason:          "Critical for exact slot deduplication",
			Priority:        "High",
			EstimatedImpact: "Very High - Prevents duplicate notifications",
		})
	}

	// Compound index for similar matches
	if !existingIndexes["user_id_1_venue_id_1_court_id_1_slot_start_time_1_last_sent_at_-1"] {
		*recommendations = append(*recommendations, IndexRecommendation{
			Type: "Compound",
			Keys: map[string]interface{}{
				"user_id":         1,
				"venue_id":        1,
				"court_id":        1,
				"slot_start_time": 1,
				"last_sent_at":    -1,
			},
			Reason:          "Optimize similar notification detection",
			Priority:        "High",
			EstimatedImpact: "High - Reduces notification spam",
		})
	}
}

func extractFieldsFromFilter(filter map[string]interface{}) []string {
	fields := []string{}

	for key, value := range filter {
		if key == "$or" || key == "$and" {
			// Handle logical operators
			if conditions, ok := value.([]interface{}); ok {
				for _, condition := range conditions {
					if condMap, ok := condition.(map[string]interface{}); ok {
						fields = append(fields, extractFieldsFromFilter(condMap)...)
					}
				}
			}
		} else if !strings.HasPrefix(key, "$") {
			// Regular field
			fields = append(fields, key)
		}
	}

	return fields
}

func indexKeysToString(keys map[string]interface{}) string {
	var parts []string

	// Sort keys for consistent string representation
	var sortedKeys []string
	for key := range keys {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		value := keys[key]
		parts = append(parts, fmt.Sprintf("%s_%v", key, value))
	}

	return strings.Join(parts, "_")
}

func estimateImpact(frequency, performance string) string {
	if frequency == "Very High" && performance == "Critical" {
		return "Very High - Critical performance improvement"
	} else if frequency == "High" && performance == "Critical" {
		return "High - Significant performance improvement"
	} else if frequency == "High" || performance == "Critical" {
		return "High - Notable performance improvement"
	} else if frequency == "Medium" || performance == "Important" {
		return "Medium - Moderate performance improvement"
	} else {
		return "Low - Minor performance improvement"
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateSummary(collections []CollectionAnalysis) AnalysisSummary {
	summary := AnalysisSummary{
		TotalCollections:  len(collections),
		UnusedIndexes:     []string{},
		MissingIndexes:    []IndexRecommendation{},
		PerformanceIssues: []string{},
	}

	for _, coll := range collections {
		summary.TotalIndexes += len(coll.Indexes)
		summary.TotalDataSize += coll.DataSize
		summary.TotalIndexSize += coll.IndexSize

		// Check for unused indexes
		for _, index := range coll.Indexes {
			if index.UsageInfo != nil && index.UsageInfo.Operations == 0 && index.Name != "_id_" {
				summary.UnusedIndexes = append(summary.UnusedIndexes, fmt.Sprintf("%s.%s", coll.Name, index.Name))
			}
		}

		// Collect high-priority recommendations
		for _, rec := range coll.Recommendations {
			if rec.Priority == "High" {
				summary.MissingIndexes = append(summary.MissingIndexes, rec)
			}
		}

		// Check for performance issues
		if coll.DocumentCount > 100000 && len(coll.Indexes) <= 2 {
			summary.PerformanceIssues = append(summary.PerformanceIssues,
				fmt.Sprintf("Collection %s has %d documents but only %d indexes",
					coll.Name, coll.DocumentCount, len(coll.Indexes)))
		}
	}

	return summary
}

func saveAnalysis(analysis *DatabaseAnalysis, filename string) error {
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func printSummary(analysis *DatabaseAnalysis) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 MONGODB INDEX ANALYSIS SUMMARY")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("Database: %s\n", analysis.DatabaseName)
	fmt.Printf("Analysis Time: %s\n", analysis.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Collections Analyzed: %d\n", analysis.Summary.TotalCollections)
	fmt.Printf("Total Indexes: %d\n", analysis.Summary.TotalIndexes)
	fmt.Printf("Total Data Size: %s\n", formatBytes(analysis.Summary.TotalDataSize))
	fmt.Printf("Total Index Size: %s\n", formatBytes(analysis.Summary.TotalIndexSize))

	if len(analysis.Summary.UnusedIndexes) > 0 {
		fmt.Printf("\n⚠️  UNUSED INDEXES (%d):\n", len(analysis.Summary.UnusedIndexes))
		for _, index := range analysis.Summary.UnusedIndexes {
			fmt.Printf("   - %s\n", index)
		}
	}

	if len(analysis.Summary.MissingIndexes) > 0 {
		fmt.Printf("\n🚀 HIGH-PRIORITY INDEX RECOMMENDATIONS (%d):\n", len(analysis.Summary.MissingIndexes))
		for _, rec := range analysis.Summary.MissingIndexes {
			fmt.Printf("   - %s: %s\n", indexKeysToString(rec.Keys), rec.Reason)
		}
	}

	if len(analysis.Summary.PerformanceIssues) > 0 {
		fmt.Printf("\n⚡ PERFORMANCE ISSUES (%d):\n", len(analysis.Summary.PerformanceIssues))
		for _, issue := range analysis.Summary.PerformanceIssues {
			fmt.Printf("   - %s\n", issue)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📋 COLLECTION DETAILS")
	fmt.Println(strings.Repeat("=", 60))

	for _, coll := range analysis.Collections {
		fmt.Printf("\n📁 %s:\n", coll.Name)
		fmt.Printf("   Documents: %s\n", formatNumber(coll.DocumentCount))
		fmt.Printf("   Data Size: %s\n", formatBytes(coll.DataSize))
		fmt.Printf("   Index Size: %s\n", formatBytes(coll.IndexSize))
		fmt.Printf("   Indexes: %d\n", len(coll.Indexes))
		fmt.Printf("   Recommendations: %d\n", len(coll.Recommendations))

		if len(coll.Recommendations) > 0 {
			fmt.Printf("   📈 Recommendations:\n")
			for _, rec := range coll.Recommendations {
				fmt.Printf("      - [%s] %s: %s\n", rec.Priority, indexKeysToString(rec.Keys), rec.Reason)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatNumber(num int64) string {
	if num < 1000 {
		return fmt.Sprintf("%d", num)
	} else if num < 1000000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	} else {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	}
}
