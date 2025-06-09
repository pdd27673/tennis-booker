# MongoDB Database Schema and Indexes

This document describes the MongoDB schema design and indexes for the Tennis Booking Bot application.

## Collections

The application uses the following MongoDB collections:

1. **users** - Stores user information
2. **venues** - Stores tennis court venue information
3. **bookings** - Stores booking requests and their status
4. **scraping_logs** - Stores logs from court availability scraping operations

## Indexes

### Users Collection

| Index Name | Fields | Properties | Purpose |
|------------|--------|------------|---------|
| _id_ | `_id` | Default | MongoDB's default primary key |
| email_1 | `email` | Unique | Ensures email uniqueness and speeds up user lookups by email |

### Venues Collection

| Index Name | Fields | Properties | Purpose |
|------------|--------|------------|---------|
| _id_ | `_id` | Default | MongoDB's default primary key |
| name_1 | `name` | Unique | Ensures venue name uniqueness and speeds up venue lookups by name |
| provider_1 | `provider` | - | Speeds up queries that filter venues by provider (e.g., "lta_clubspark") |
| is_active_1 | `is_active` | - | Optimizes queries for active venues |

### Bookings Collection

| Index Name | Fields | Properties | Purpose |
|------------|--------|------------|---------|
| _id_ | `_id` | Default | MongoDB's default primary key |
| user_id_1 | `user_id` | - | Speeds up queries for bookings by a specific user |
| venue_id_1 | `venue_id` | - | Speeds up queries for bookings at a specific venue |
| date_1 | `date` | - | Optimizes date-based booking queries |
| status_1 | `status` | - | Optimizes queries that filter bookings by status |
| venue_id_1_court_id_1_date_1_start_time_1 | `venue_id`, `court_id`, `date`, `start_time` | Unique | Prevents double-booking the same court at the same time |

### Scraping Logs Collection

| Index Name | Fields | Properties | Purpose |
|------------|--------|------------|---------|
| _id_ | `_id` | Default | MongoDB's default primary key |
| venue_id_1 | `venue_id` | - | Speeds up queries for logs by venue |
| scrape_timestamp_-1 | `scrape_timestamp` | Descending | Optimizes time-based queries, newest first |
| venue_id_1_scrape_timestamp_-1 | `venue_id`, `scrape_timestamp` | Compound | Optimizes queries for logs by venue sorted by time |
| success_1 | `success` | - | Speeds up queries that filter logs by success/failure |
| run_id_1 | `run_id` | - | Speeds up queries for logs from a specific scraping run |
| provider_1_scrape_timestamp_-1 | `provider`, `scrape_timestamp` | Compound | Optimizes provider-specific time-based queries |
| created_at_1 | `created_at` | TTL: 30 days | Automatically deletes logs older than 30 days |

## Managing Indexes

The application includes tools for managing database indexes:

### Command Line Tool

Use the `ensure-indexes` tool to create or verify indexes:

```bash
# Create all indexes
./bin/ensure-indexes

# Verify indexes
./bin/ensure-indexes --verify

# Verify indexes with detailed output
./bin/ensure-indexes --verify --verbose
```

### Makefile Targets

The project includes Makefile targets for index management:

```bash
# Create all indexes
make db-ensure-indexes

# Verify indexes
make db-verify-indexes

# Verify indexes with detailed output
make db-verify-indexes-verbose
```

### Programmatic Usage

The database package provides functions for index management:

```go
import "tennis-booking-bot/internal/database"

// Initialize database connection
db, err := database.InitDatabase(mongoURI, dbName)
if err != nil {
    log.Fatal(err)
}

// Create all indexes
err = database.CreateAllIndexes(db)
if err != nil {
    log.Fatal(err)
}

// Get index summary
summaries, err := database.GetIndexSummary(db)
if err != nil {
    log.Fatal(err)
}

// Print index information
for _, summary := range summaries {
    fmt.Printf("Collection: %s, Index: %s\n", summary.Collection, summary.IndexName)
}
``` 