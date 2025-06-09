# 🛠️ Tennis Court Booking System - Utilities

This directory contains utility scripts and tools for managing the tennis court booking system.

## 📁 Directory Contents

### Core Utilities

- **`seed_venues.py`** - Database seeding script for tennis venues
  - Seeds MongoDB with venue data (Victoria Park, Ropemakers Field, Stratford Park)
  - Creates court configurations and provider settings
  - Usage: `python utils/seed_venues.py` or `./scripts/seed_venues.sh`

- **`redis_worker.py`** - Redis notification worker
  - Processes court availability notifications
  - Handles email alerts and user preferences
  - Integrates with the notification system

- **`verify_mongodb_data.py`** - MongoDB data verification
  - Validates venue data integrity
  - Checks court configurations
  - Useful for debugging database issues

- **`verify_venues.py`** - Venue configuration verification
  - Validates venue URLs and accessibility
  - Checks provider-specific configurations
  - Useful for troubleshooting scraping issues

- **`requirements.txt`** - Python dependencies
  - Contains all required Python packages
  - Used by the virtual environment setup

## 🚀 Usage

### Setting up the Python Environment
```bash
# Create virtual environment (done by setup script)
python -m venv scraper-env
source scraper-env/bin/activate
pip install -r utils/requirements.txt
```

### Running Utilities
```bash
# Seed venues (recommended way)
./scripts/seed_venues.sh

# Or directly with Python
source scraper-env/bin/activate
python utils/seed_venues.py

# Verify MongoDB data
python utils/verify_mongodb_data.py

# Verify venue configurations
python utils/verify_venues.py
```

### Redis Worker
```bash
# Start Redis notification worker
source scraper-env/bin/activate
python utils/redis_worker.py
```

## 🔧 Configuration

Most utilities read configuration from environment variables:

```bash
# MongoDB
MONGODB_URI=mongodb://localhost:27017

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password

# Email notifications
SENDGRID_API_KEY=SG....
```

## 📋 Integration with Main System

These utilities are integrated into the main control script:

```bash
# Setup includes venue seeding
./run_tennis_system.sh setup

# System tests include utility validation
./run_tennis_system.sh test-system
```

## 🧪 Testing

Related test files are in the main `tests/` directory:
- `test_integration_mongodb.py` - MongoDB integration tests
- `test_seed_venues.py` - Venue seeding tests
- `test_mongo.py` - Basic MongoDB connectivity tests

## 📝 Notes

- All utilities require the Python virtual environment to be activated
- MongoDB and Redis services should be running before using these utilities
- The master control script (`run_tennis_system.sh`) handles most common operations
- Use these utilities directly for debugging or advanced operations 