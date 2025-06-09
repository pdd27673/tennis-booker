# Tennis Booking Bot - Seeding Scripts

This directory contains scripts for seeding the MongoDB database with initial data for the Tennis Booking Bot application.

## Venue Seeding Script

The venue seeding script (`seed_venues.py`) populates the MongoDB venues collection with initial data for:
- LTA/Clubspark venues (Will to Win Regent's Park, Will to Win Hyde Park)
- courtsides.com/tennistowerhamlets venues (Tower Hamlets Tennis)

### Prerequisites

- MongoDB running (via Docker Compose or standalone)
- Python environment set up with required dependencies

### Running the Script

You can run the script in two ways:

#### 1. Using the Shell Script (Recommended)

```bash
# From the project root
./scripts/seed_venues.sh
```

This shell script will:
- Activate the Python virtual environment
- Run the seeding script
- Deactivate the virtual environment when done

#### 2. Directly Using Python

```bash
# Activate the Python virtual environment
source scraper-env/bin/activate

# Run the script
python scripts/python/seed_venues.py

# Deactivate when done
deactivate
```

### Testing and Verification

#### Testing the Seeding Script

A test script is available to verify the venue seeding functionality:

```bash
# Activate the Python virtual environment
source scraper-env/bin/activate

# Run the tests
python scripts/python/test_seed_venues.py

# Deactivate when done
deactivate
```

The test script:
- Creates a separate test database (`tennis_booking_test`)
- Tests all seeding functions without affecting the main database
- Validates the structure of venue data
- Drops the test database when finished

#### Verifying the Seeded Data

To verify the venues that have been seeded in the database:

```bash
# Activate the Python virtual environment
source scraper-env/bin/activate

# Run the verification script
python scripts/python/verify_venues.py

# Deactivate when done
deactivate
```

The verification script:
- Connects to the actual database
- Counts and displays all venues by provider
- Shows venue details including name, URL, location, and court count

### Configuration

The script uses the following environment variables from your `.env` file:

- `MONGO_URI`: MongoDB connection string (default: `mongodb://admin:YOUR_PASSWORD@localhost:27017`)
- `MONGO_DB_NAME`: MongoDB database name (default: `tennis_booking`)

You can modify these in your `.env` file to target a different MongoDB instance.

### Adding New Venues

To add new venues, edit the `seed_venues.py` file and add new venue definitions to either:
- `LTA_CLUBSPARK_VENUES` list for LTA/Clubspark venues
- `COURTSIDES_VENUES` list for courtsides.com venues

Each venue definition should follow the structure in the existing examples, matching the Go `Venue` struct. 