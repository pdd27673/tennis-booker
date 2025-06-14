// MongoDB Index Verification Script
// This script verifies that indexes are being used effectively by analyzing query execution plans
// Run with: mongosh < verify_indexes.js

print("🔍 Verifying MongoDB Index Effectiveness");
print("========================================");

// Get current database
const db = db.getSiblingDB('tennis_booking');

// Helper function to format execution stats
function formatExecutionStats(stats) {
    const result = {
        stage: stats.executionStats.executionStages.stage,
        executionTimeMillis: stats.executionStats.executionTimeMillis,
        totalKeysExamined: stats.executionStats.totalKeysExamined,
        totalDocsExamined: stats.executionStats.totalDocsExamined,
        nReturned: stats.executionStats.nReturned,
        indexUsed: stats.executionStats.executionStages.indexName || 'None'
    };
    
    // Calculate efficiency ratios
    result.keyEfficiency = result.nReturned > 0 ? (result.totalKeysExamined / result.nReturned).toFixed(2) : 'N/A';
    result.docEfficiency = result.nReturned > 0 ? (result.totalDocsExamined / result.nReturned).toFixed(2) : 'N/A';
    
    return result;
}

// Helper function to analyze and print query performance
function analyzeQuery(collection, query, description, sortOptions = null) {
    print(`\n📋 Testing: ${description}`);
    print(`   Collection: ${collection}`);
    print(`   Query: ${JSON.stringify(query)}`);
    if (sortOptions) {
        print(`   Sort: ${JSON.stringify(sortOptions)}`);
    }
    
    try {
        let cursor = db[collection].find(query);
        if (sortOptions) {
            cursor = cursor.sort(sortOptions);
        }
        
        const stats = cursor.explain("executionStats");
        const formatted = formatExecutionStats(stats);
        
        print(`   📊 Results:`);
        print(`      Stage: ${formatted.stage}`);
        print(`      Execution Time: ${formatted.executionTimeMillis}ms`);
        print(`      Index Used: ${formatted.indexUsed}`);
        print(`      Keys Examined: ${formatted.totalKeysExamined}`);
        print(`      Docs Examined: ${formatted.totalDocsExamined}`);
        print(`      Documents Returned: ${formatted.nReturned}`);
        print(`      Key Efficiency: ${formatted.keyEfficiency} (keys/doc returned)`);
        print(`      Doc Efficiency: ${formatted.docEfficiency} (docs examined/doc returned)`);
        
        // Performance assessment
        if (formatted.stage === 'IXSCAN') {
            print(`      ✅ GOOD: Using index scan`);
        } else if (formatted.stage === 'COLLSCAN') {
            print(`      ⚠️  WARNING: Using collection scan (no index)`);
        }
        
        if (formatted.executionTimeMillis < 10) {
            print(`      ✅ GOOD: Fast execution (<10ms)`);
        } else if (formatted.executionTimeMillis < 100) {
            print(`      ⚡ OK: Moderate execution (10-100ms)`);
        } else {
            print(`      🐌 SLOW: Slow execution (>100ms)`);
        }
        
        if (parseFloat(formatted.docEfficiency) <= 1.5) {
            print(`      ✅ GOOD: Efficient document examination`);
        } else {
            print(`      ⚠️  WARNING: Examining too many documents`);
        }
        
        return formatted;
        
    } catch (error) {
        print(`      ❌ ERROR: ${error.message}`);
        return null;
    }
}

// Helper function to check if collection exists and has data
function checkCollection(collectionName) {
    try {
        const count = db[collectionName].countDocuments();
        print(`📊 ${collectionName}: ${count} documents`);
        return count > 0;
    } catch (error) {
        print(`📊 ${collectionName}: Collection not found or error - ${error.message}`);
        return false;
    }
}

print("\n📊 Collection Status Check:");
print("===========================");
const collections = ['court_slots', 'slots', 'user_preferences', 'scraping_logs', 'deduplication_records', 'notifications', 'venues', 'users'];
const availableCollections = {};

collections.forEach(coll => {
    availableCollections[coll] = checkCollection(coll);
});

print("\n🔧 Index Verification Tests:");
print("============================");

// ============================================================================
// COURT SLOTS COLLECTION TESTS
// ============================================================================

if (availableCollections['court_slots'] || availableCollections['slots']) {
    const slotsCollection = availableCollections['court_slots'] ? 'court_slots' : 'slots';
    
    print(`\n📁 ${slotsCollection.toUpperCase()} COLLECTION TESTS:`);
    
    // Test 1: Venue-Date-Time compound index
    analyzeQuery(
        slotsCollection,
        { venue_id: "venue123", slot_date: { $gte: new Date("2025-06-14") } },
        "Venue + Date Range Query (should use idx_venue_date_time)",
        { start_time: 1 }
    );
    
    // Test 2: Availability and freshness
    analyzeQuery(
        slotsCollection,
        { is_available: true, last_scraped: { $gte: new Date(Date.now() - 3600000) } },
        "Available Slots + Recent Scrape (should use idx_availability_freshness)"
    );
    
    // Test 3: Court-specific query
    analyzeQuery(
        slotsCollection,
        { venue_id: "venue123", court_id: "court1", slot_date: new Date("2025-06-14") },
        "Venue + Court + Date (should use idx_venue_court_date)"
    );
    
    // Test 4: Date range only (should use TTL index)
    analyzeQuery(
        slotsCollection,
        { slot_date: { $gte: new Date("2025-06-14"), $lte: new Date("2025-06-21") } },
        "Date Range Query (may use idx_slot_date_ttl)"
    );
}

// ============================================================================
// USER PREFERENCES COLLECTION TESTS
// ============================================================================

if (availableCollections['user_preferences']) {
    print(`\n📁 USER PREFERENCES COLLECTION TESTS:`);
    
    // Test 1: User lookup (unique index)
    analyzeQuery(
        'user_preferences',
        { user_id: "user123" },
        "User Preferences Lookup (should use idx_user_id_unique)"
    );
    
    // Test 2: Notification settings
    analyzeQuery(
        'user_preferences',
        { user_id: "user123", "notification_settings.enabled": true },
        "User + Notification Settings (should use idx_user_notifications)"
    );
    
    // Test 3: Venue preferences (sparse index)
    analyzeQuery(
        'user_preferences',
        { preferred_venues: { $in: ["venue123", "venue456"] } },
        "Preferred Venues Query (should use idx_preferred_venues sparse)"
    );
    
    // Test 4: Sport preferences
    analyzeQuery(
        'user_preferences',
        { preferred_sports: "tennis" },
        "Preferred Sports Query (should use idx_preferred_sports sparse)"
    );
}

// ============================================================================
// SCRAPING LOGS COLLECTION TESTS
// ============================================================================

if (availableCollections['scraping_logs']) {
    print(`\n📁 SCRAPING LOGS COLLECTION TESTS:`);
    
    // Test 1: Venue + timestamp monitoring
    analyzeQuery(
        'scraping_logs',
        { venue_id: "venue123", scrape_timestamp: { $gte: new Date(Date.now() - 86400000) } },
        "Venue Logs Last 24h (should use idx_venue_timestamp)"
    );
    
    // Test 2: Status-based queries
    analyzeQuery(
        'scraping_logs',
        { status: "success", scrape_timestamp: { $gte: new Date(Date.now() - 3600000) } },
        "Successful Scrapes Last Hour (should use idx_status_timestamp)"
    );
    
    // Test 3: Recent logs (TTL index)
    analyzeQuery(
        'scraping_logs',
        { scrape_timestamp: { $gte: new Date(Date.now() - 7200000) } },
        "Recent Logs Query (may use idx_scrape_timestamp_ttl)"
    );
}

// ============================================================================
// DEDUPLICATION RECORDS COLLECTION TESTS
// ============================================================================

if (availableCollections['deduplication_records']) {
    print(`\n📁 DEDUPLICATION RECORDS COLLECTION TESTS:`);
    
    // Test 1: Exact match lookup
    analyzeQuery(
        'deduplication_records',
        { user_id: "user123", slot_key: "venue123:2025-06-14:10:00:court1" },
        "Exact Duplicate Check (should use idx_user_slot_key)"
    );
    
    // Test 2: Complex match
    analyzeQuery(
        'deduplication_records',
        { user_id: "user123", venue_id: "venue123", slot_date: new Date("2025-06-14"), start_time: "10:00" },
        "Similar Slot Detection (should use idx_user_venue_date_time)"
    );
    
    // Test 3: Cleanup query (TTL)
    analyzeQuery(
        'deduplication_records',
        { last_sent_at: { $lte: new Date(Date.now() - 604800000) } },
        "Old Records Query (may use idx_last_sent_ttl)"
    );
}

// ============================================================================
// NOTIFICATIONS COLLECTION TESTS
// ============================================================================

if (availableCollections['notifications']) {
    print(`\n📁 NOTIFICATIONS COLLECTION TESTS:`);
    
    // Test 1: User notifications
    analyzeQuery(
        'notifications',
        { user_id: "user123", created_at: { $gte: new Date(Date.now() - 86400000) } },
        "User Notifications Last 24h (should use idx_user_created)"
    );
    
    // Test 2: Status-based
    analyzeQuery(
        'notifications',
        { status: "sent", created_at: { $gte: new Date(Date.now() - 3600000) } },
        "Sent Notifications Last Hour (should use idx_status_created)"
    );
    
    // Test 3: Notification type
    analyzeQuery(
        'notifications',
        { type: "slot_available", user_id: "user123" },
        "Notification Type + User (should use idx_type_user)"
    );
}

// ============================================================================
// VENUES COLLECTION TESTS
// ============================================================================

if (availableCollections['venues']) {
    print(`\n📁 VENUES COLLECTION TESTS:`);
    
    // Test 1: Geospatial query
    analyzeQuery(
        'venues',
        { location: { $near: { $geometry: { type: "Point", coordinates: [-122.4194, 37.7749] }, $maxDistance: 5000 } } },
        "Nearby Venues (should use idx_location_geo)"
    );
    
    // Test 2: Active venues
    analyzeQuery(
        'venues',
        { active: true, name: { $regex: "Tennis", $options: "i" } },
        "Active Tennis Venues (should use idx_active_name)"
    );
    
    // Test 3: Sport filtering
    analyzeQuery(
        'venues',
        { sports: "tennis", active: true },
        "Tennis Venues (should use idx_sports_active)"
    );
}

// ============================================================================
// USERS COLLECTION TESTS
// ============================================================================

if (availableCollections['users']) {
    print(`\n📁 USERS COLLECTION TESTS:`);
    
    // Test 1: Email lookup (unique)
    analyzeQuery(
        'users',
        { email: "user@example.com" },
        "User by Email (should use idx_email_unique)"
    );
    
    // Test 2: Active users
    analyzeQuery(
        'users',
        { active: true, created_at: { $gte: new Date(Date.now() - 2592000000) } },
        "Active Users Last 30 Days (should use idx_active_created)"
    );
}

// ============================================================================
// SUMMARY AND RECOMMENDATIONS
// ============================================================================

print("\n📋 VERIFICATION SUMMARY:");
print("========================");
print("✅ Completed index effectiveness verification");
print("📊 Key Performance Indicators to Monitor:");
print("   • Execution Time: <10ms (good), 10-100ms (ok), >100ms (needs optimization)");
print("   • Stage: IXSCAN (good), COLLSCAN (needs index)");
print("   • Doc Efficiency: ≤1.5 (good), >1.5 (examining too many docs)");
print("   • Key Efficiency: ≤2.0 (good), >2.0 (index not selective enough)");

print("\n🔧 OPTIMIZATION RECOMMENDATIONS:");
print("=================================");
print("1. If queries show COLLSCAN, ensure indexes are created correctly");
print("2. If execution time >100ms, consider adding more specific indexes");
print("3. If doc efficiency >1.5, review index field order (ESR rule)");
print("4. Monitor index usage with: db.collection.aggregate([{$indexStats: {}}])");
print("5. Use MongoDB Compass or profiler for ongoing performance monitoring");

print("\n🎯 NEXT STEPS:");
print("===============");
print("1. Run this script after creating indexes with create_indexes.js");
print("2. Compare performance before/after index creation");
print("3. Monitor production queries and adjust indexes as needed");
print("4. Set up automated performance monitoring");

print("\n✅ Index verification completed!"); 