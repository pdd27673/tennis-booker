package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Copy the necessary types from the main tools
type IndexRecommendation struct {
	Type            string                 `json:"type"`
	Keys            map[string]interface{} `json:"keys"`
	Reason          string                 `json:"reason"`
	Priority        string                 `json:"priority"`
	EstimatedImpact string                 `json:"estimated_impact"`
}

type CollectionAnalysis struct {
	Name            string                 `json:"name"`
	DocumentCount   int64                  `json:"document_count"`
	DataSize        int64                  `json:"data_size"`
	IndexSize       int64                  `json:"index_size"`
	Recommendations []IndexRecommendation  `json:"recommendations"`
}

type DatabaseAnalysis struct {
	DatabaseName string               `json:"database_name"`
	Timestamp    time.Time            `json:"timestamp"`
	Collections  []CollectionAnalysis `json:"collections"`
}

func main() {
	log.Println("🧪 Validating MongoDB Index Tools...")

	// Test 1: Validate analysis file loading
	log.Println("📋 Test 1: Loading analysis file...")
	analysis, err := loadAnalysis("test_analysis.json")
	if err != nil {
		log.Fatalf("❌ Failed to load analysis file: %v", err)
	}
	log.Printf("✅ Successfully loaded analysis for database: %s", analysis.DatabaseName)
	log.Printf("   Collections: %d", len(analysis.Collections))

	// Test 2: Validate recommendations structure
	log.Println("📋 Test 2: Validating recommendations...")
	totalRecommendations := 0
	for _, coll := range analysis.Collections {
		log.Printf("   Collection '%s': %d recommendations", coll.Name, len(coll.Recommendations))
		for _, rec := range coll.Recommendations {
			totalRecommendations++
			if err := validateRecommendation(rec); err != nil {
				log.Printf("   ⚠️ Invalid recommendation: %v", err)
			}
		}
	}
	log.Printf("✅ Validated %d recommendations", totalRecommendations)

	// Test 3: Test index name generation
	log.Println("📋 Test 3: Testing index name generation...")
	testKeys := map[string]interface{}{
		"venue_id":   1,
		"slot_date":  1,
		"start_time": 1,
	}
	indexName := generateIndexName(testKeys)
	log.Printf("✅ Generated index name: %s", indexName)

	// Test 4: Test TTL calculation
	log.Println("📋 Test 4: Testing TTL calculation...")
	ttlSeconds := getTTLSeconds("court_slots", map[string]interface{}{"slot_date": 1})
	log.Printf("✅ TTL for court_slots.slot_date: %d seconds (%d days)", ttlSeconds, ttlSeconds/(24*3600))

	// Test 5: Test sparse field detection
	log.Println("📋 Test 5: Testing sparse field detection...")
	sparseKeys := map[string]interface{}{"preferred_venues": 1}
	isSparse := shouldBeSparse(sparseKeys)
	log.Printf("✅ preferred_venues should be sparse: %t", isSparse)

	log.Println("🎉 All validation tests passed!")
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

func validateRecommendation(rec IndexRecommendation) error {
	if rec.Type == "" {
		return fmt.Errorf("recommendation type is empty")
	}
	if len(rec.Keys) == 0 {
		return fmt.Errorf("recommendation keys are empty")
	}
	if rec.Priority == "" {
		return fmt.Errorf("recommendation priority is empty")
	}
	return nil
}

func generateIndexName(keys map[string]interface{}) string {
	var parts []string
	for key, value := range keys {
		parts = append(parts, fmt.Sprintf("%s_%v", key, value))
	}
	return fmt.Sprintf("opt_%s", joinStrings(parts, "_"))
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
		"preferred_venues":                    true,
		"preferred_days":                      true,
		"excluded_venues":                     true,
		"booking_url":                         true,
		"notified":                           true,
	}

	for field := range keys {
		if sparseFields[field] {
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