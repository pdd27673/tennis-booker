// MongoDB Index Creation Script
// Run this script in MongoDB shell: mongosh < create_indexes.js
// Or load in MongoDB shell: load("create_indexes.js")

print("🚀 Creating optimized indexes for Tennis Booking application...");
print("================================================================");

// Get current database
const db = db.getSiblingDB('tennis_booking');

// Function to create index with error handling
function createIndexSafely(collection, indexSpec, options, description) {
    print(`\n📋 Creating ${description}...`);
    try {
        const result = db[collection].createIndex(indexSpec, options);
        print(`✅ Success: ${JSON.stringify(result)}`);
        return true;
    } catch (error) {
        if (error.message.includes('already exists')) {
            print(`ℹ️  Index already exists, skipping`);
            return true;
        } else {
            print(`❌ Error: ${error.message}`);
            return false;
        }
    }
}

// Function to show collection stats
function showCollectionStats(collectionName) {
    try {
        const stats = db[collectionName].stats();
        print(`📊 ${collectionName}: ${stats.count} documents, ${(stats.size / 1024 / 1024).toFixed(2)} MB data`);
    } catch (error) {
        print(`📊 ${collectionName}: Collection not found or empty`);
    }
}

print("\n📊 Current Collection Statistics:");
print("================================");
showCollectionStats('court_slots');
showCollectionStats('slots');
showCollectionStats('user_preferences');
showCollectionStats('scraping_logs');
showCollectionStats('deduplication_records');
showCollectionStats('notifications');

print("\n🔧 Creating Indexes...");
print("======================");

// ============================================================================
// COURT SLOTS / SLOTS COLLECTION INDEXES
// ============================================================================

print("\n📁 Court Slots Collection Indexes:");

// Primary compound index for venue-date-time queries
createIndexSafely(
    'court_slots',
    { venue_id: 1, slot_date: 1, start_time: 1 },
    { name: 'idx_venue_date_time', background: true },
    'Venue-Date-Time compound index'
);

// Alternative for 'slots' collection if used instead
createIndexSafely(
    'slots',
    { venue_id: 1, slot_date: 1, start_time: 1 },
    { name: 'idx_venue_date_time', background: true },
    'Venue-Date-Time compound index (slots collection)'
);

// Availability and freshness index
createIndexSafely(
    'court_slots',
    { is_available: 1, last_scraped: 1 },
    { name: 'idx_availability_freshness', background: true },
    'Availability and freshness index'
);

createIndexSafely(
    'slots',
    { is_available: 1, last_scraped: 1 },
    { name: 'idx_availability_freshness', background: true },
    'Availability and freshness index (slots collection)'
);

// TTL index for automatic cleanup (7 days)
createIndexSafely(
    'court_slots',
    { slot_date: 1 },
    { name: 'idx_slot_date_ttl', expireAfterSeconds: 604800, background: true },
    'TTL index for slot cleanup (7 days)'
);

createIndexSafely(
    'slots',
    { slot_date: 1 },
    { name: 'idx_slot_date_ttl', expireAfterSeconds: 604800, background: true },
    'TTL index for slot cleanup (7 days, slots collection)'
);

// Court-specific queries
createIndexSafely(
    'court_slots',
    { venue_id: 1, court_id: 1, slot_date: 1 },
    { name: 'idx_venue_court_date', background: true },
    'Venue-Court-Date index'
);

createIndexSafely(
    'slots',
    { venue_id: 1, court_id: 1, slot_date: 1 },
    { name: 'idx_venue_court_date', background: true },
    'Venue-Court-Date index (slots collection)'
);

// ============================================================================
// USER PREFERENCES COLLECTION INDEXES
// ============================================================================

print("\n📁 User Preferences Collection Indexes:");

// Unique user index
createIndexSafely(
    'user_preferences',
    { user_id: 1 },
    { name: 'idx_user_id_unique', unique: true, background: true },
    'Unique user ID index'
);

// Notification settings compound index
createIndexSafely(
    'user_preferences',
    { user_id: 1, 'notification_settings.enabled': 1 },
    { name: 'idx_user_notifications', background: true },
    'User notification settings index'
);

// Venue preferences (sparse for optional field)
createIndexSafely(
    'user_preferences',
    { preferred_venues: 1 },
    { name: 'idx_preferred_venues', sparse: true, background: true },
    'Preferred venues index (sparse)'
);

// Sport preferences
createIndexSafely(
    'user_preferences',
    { preferred_sports: 1 },
    { name: 'idx_preferred_sports', sparse: true, background: true },
    'Preferred sports index (sparse)'
);

// Time preferences
createIndexSafely(
    'user_preferences',
    { preferred_times: 1 },
    { name: 'idx_preferred_times', sparse: true, background: true },
    'Preferred times index (sparse)'
);

// ============================================================================
// SCRAPING LOGS COLLECTION INDEXES
// ============================================================================

print("\n📁 Scraping Logs Collection Indexes:");

// Venue and timestamp compound index
createIndexSafely(
    'scraping_logs',
    { venue_id: 1, scrape_timestamp: 1 },
    { name: 'idx_venue_timestamp', background: true },
    'Venue-Timestamp compound index'
);

// TTL index for log cleanup (30 days)
createIndexSafely(
    'scraping_logs',
    { scrape_timestamp: 1 },
    { name: 'idx_scrape_timestamp_ttl', expireAfterSeconds: 2592000, background: true },
    'TTL index for log cleanup (30 days)'
);

// Status-based queries
createIndexSafely(
    'scraping_logs',
    { status: 1, scrape_timestamp: 1 },
    { name: 'idx_status_timestamp', background: true },
    'Status-Timestamp index'
);

// ============================================================================
// DEDUPLICATION RECORDS COLLECTION INDEXES
// ============================================================================

print("\n📁 Deduplication Records Collection Indexes:");

// Exact match index for fast duplicate detection
createIndexSafely(
    'deduplication_records',
    { user_id: 1, slot_key: 1 },
    { name: 'idx_user_slot_key', background: true },
    'User-SlotKey exact match index'
);

// Complex match index for similar slot detection
createIndexSafely(
    'deduplication_records',
    { user_id: 1, venue_id: 1, slot_date: 1, start_time: 1 },
    { name: 'idx_user_venue_date_time', background: true },
    'User-Venue-Date-Time complex match index'
);

// TTL index for cleanup (7 days)
createIndexSafely(
    'deduplication_records',
    { last_sent_at: 1 },
    { name: 'idx_last_sent_ttl', expireAfterSeconds: 604800, background: true },
    'TTL index for deduplication cleanup (7 days)'
);

// ============================================================================
// NOTIFICATIONS COLLECTION INDEXES
// ============================================================================

print("\n📁 Notifications Collection Indexes:");

// User notifications index
createIndexSafely(
    'notifications',
    { user_id: 1, created_at: 1 },
    { name: 'idx_user_created', background: true },
    'User-Created timestamp index'
);

// Status-based queries
createIndexSafely(
    'notifications',
    { status: 1, created_at: 1 },
    { name: 'idx_status_created', background: true },
    'Status-Created timestamp index'
);

// TTL index for old notifications (90 days)
createIndexSafely(
    'notifications',
    { created_at: 1 },
    { name: 'idx_created_ttl', expireAfterSeconds: 7776000, background: true },
    'TTL index for notification cleanup (90 days)'
);

// Notification type index
createIndexSafely(
    'notifications',
    { type: 1, user_id: 1 },
    { name: 'idx_type_user', background: true },
    'Type-User index'
);

// ============================================================================
// VENUES COLLECTION INDEXES
// ============================================================================

print("\n📁 Venues Collection Indexes:");

// Location-based queries
createIndexSafely(
    'venues',
    { location: '2dsphere' },
    { name: 'idx_location_geo', background: true },
    'Geospatial location index'
);

// Active venues
createIndexSafely(
    'venues',
    { active: 1, name: 1 },
    { name: 'idx_active_name', background: true },
    'Active venues by name index'
);

// Sport type
createIndexSafely(
    'venues',
    { sports: 1, active: 1 },
    { name: 'idx_sports_active', background: true },
    'Sports-Active index'
);

// ============================================================================
// USERS COLLECTION INDEXES
// ============================================================================

print("\n📁 Users Collection Indexes:");

// Email unique index
createIndexSafely(
    'users',
    { email: 1 },
    { name: 'idx_email_unique', unique: true, background: true },
    'Unique email index'
);

// Active users
createIndexSafely(
    'users',
    { active: 1, created_at: 1 },
    { name: 'idx_active_created', background: true },
    'Active users by creation date'
);

// ============================================================================
// VERIFICATION AND SUMMARY
// ============================================================================

print("\n🔍 Verifying Created Indexes...");
print("===============================");

function showIndexes(collectionName) {
    try {
        const indexes = db[collectionName].getIndexes();
        print(`\n📋 ${collectionName} indexes (${indexes.length}):`);
        indexes.forEach(index => {
            const keys = Object.keys(index.key).map(k => `${k}:${index.key[k]}`).join(', ');
            const options = [];
            if (index.unique) options.push('unique');
            if (index.sparse) options.push('sparse');
            if (index.expireAfterSeconds) options.push(`TTL:${index.expireAfterSeconds}s`);
            const optStr = options.length > 0 ? ` (${options.join(', ')})` : '';
            print(`  • ${index.name}: {${keys}}${optStr}`);
        });
    } catch (error) {
        print(`📋 ${collectionName}: Collection not found`);
    }
}

showIndexes('court_slots');
showIndexes('slots');
showIndexes('user_preferences');
showIndexes('scraping_logs');
showIndexes('deduplication_records');
showIndexes('notifications');
showIndexes('venues');
showIndexes('users');

print("\n✅ Index creation completed!");
print("============================");
print("📊 Use db.collection.explain('executionStats') to verify index usage");
print("🔍 Monitor performance with db.collection.aggregate([{$indexStats: {}}])");
print("⚡ Indexes will improve query performance for:");
print("   • Slot discovery by venue, date, and time");
print("   • User preference lookups");
print("   • Notification targeting");
print("   • Scraping log monitoring");
print("   • Automatic data cleanup via TTL indexes");
print("\n🎉 MongoDB optimization complete!"); 