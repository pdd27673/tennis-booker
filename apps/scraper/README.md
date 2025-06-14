# Tennis Booking Scraper

A Python-based web scraper for collecting court slot availability from various tennis venues. Features Redis-based deduplication and MongoDB integration for efficient data processing.

## Features

- **Multi-venue scraping** with configurable scrapers
- **Redis deduplication** to prevent duplicate slot processing
- **MongoDB integration** for data persistence
- **Comprehensive logging** and metrics
- **Docker containerization** for easy deployment
- **Graceful error handling** and retry mechanisms

## Quick Start

```bash
# Set up virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Configure environment
cp .env.example .env
# Edit .env with your configuration

# Run scraper
python -m src.main
```

## Architecture

### Redis Deduplication

The scraper uses Redis for efficient slot deduplication with automatic expiry:

```python
from src.deduplication.redis_deduplicator import RedisDeduplicator

# Initialize deduplicator
deduplicator = RedisDeduplicator(
    host='localhost',
    port=6379,
    db=1,  # Separate DB for deduplication
    expiry_hours=48  # 48-hour expiry for weekend patterns
)

# Check for duplicates
slot_data = {
    'venue_id': 'venue123',
    'court_id': 'court1',
    'slot_date': '2025-06-15',
    'start_time': '10:00'
}

if not deduplicator.is_duplicate_slot(slot_data):
    # Process new slot
    process_slot(slot_data)
else:
    # Skip duplicate
    skip_slot(slot_data)
```

### Key Format

Redis keys use the format: `dedupe:slot:<venueId>:<date>:<startTime>:<courtId>`

This ensures unique identification while allowing efficient pattern matching and cleanup.

### Performance Metrics

The deduplicator tracks comprehensive metrics:

```python
# Get performance statistics
stats = deduplicator.get_cache_statistics()
print(f"Total checks: {stats['total_checks']}")
print(f"Duplicates found: {stats['duplicates_found']}")
print(f"Hit rate: {stats['hit_rate']:.1f}%")
print(f"Memory usage: {stats['memory_usage_mb']:.1f}MB")
```

## Configuration

### Environment Variables

```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=1
REDIS_PASSWORD=  # Optional

# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=tennis_booking

# Scraper Settings
SCRAPER_INTERVAL_MINUTES=30
DEDUP_EXPIRY_HOURS=48
MAX_RETRIES=3
TIMEOUT_SECONDS=30

# Logging
LOG_LEVEL=INFO
LOG_FORMAT=json
```

### Scraper Configuration

Configure venues and scraping parameters in `config/scrapers.json`:

```json
{
  "venues": [
    {
      "id": "venue1",
      "name": "Tennis Club A",
      "scraper_class": "TennisClubAScraper",
      "base_url": "https://venue1.com",
      "enabled": true,
      "interval_minutes": 30
    }
  ],
  "global_settings": {
    "user_agent": "TennisBookingScraper/1.0",
    "timeout_seconds": 30,
    "max_retries": 3
  }
}
```

## Development

### Project Structure

```
src/
├── scrapers/           # Venue-specific scrapers
│   ├── base_scraper.py
│   ├── scraper_orchestrator.py
│   └── venue_scrapers/
├── deduplication/      # Redis deduplication
│   └── redis_deduplicator.py
├── database/          # MongoDB integration
│   └── mongo_client.py
├── utils/             # Utilities
│   ├── logging.py
│   └── config.py
└── main.py           # Entry point

tests/                # Test suite
├── test_redis_deduplicator.py
└── test_scrapers.py
```

### Adding New Scrapers

1. Create a new scraper class inheriting from `BaseScraper`:

```python
from src.scrapers.base_scraper import BaseScraper

class NewVenueScraper(BaseScraper):
    def __init__(self, venue_config):
        super().__init__(venue_config)
    
    def scrape_slots(self, date_range):
        # Implement venue-specific scraping logic
        slots = []
        # ... scraping implementation
        return slots
```

2. Register the scraper in `scrapers.json`
3. Add tests in `tests/`

### Testing

```bash
# Run all tests
python -m pytest tests/

# Run specific test
python -m pytest tests/test_redis_deduplicator.py -v

# Run with coverage
python -m pytest tests/ --cov=src --cov-report=html
```

### Performance Testing

Test deduplication performance:

```python
# Generate test data
from tests.test_redis_deduplicator import generate_test_slots

slots = generate_test_slots(1000)
deduplicator = RedisDeduplicator()

# Measure performance
import time
start = time.time()
for slot in slots:
    deduplicator.is_duplicate_slot(slot)
end = time.time()

print(f"Processed {len(slots)} slots in {end-start:.2f}s")
print(f"Throughput: {len(slots)/(end-start):.0f} slots/sec")
```

## Deployment

### Docker

```bash
# Build image
docker build -t tennis-booking-scraper .

# Run container
docker run -d \
  --name scraper \
  --env-file .env \
  tennis-booking-scraper

# View logs
docker logs -f scraper
```

### Docker Compose

```yaml
version: '3.8'
services:
  scraper:
    build: .
    environment:
      - REDIS_HOST=redis
      - MONGODB_URI=mongodb://mongo:27017
    depends_on:
      - redis
      - mongo
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  mongo:
    image: mongo:7
    ports:
      - "27017:27017"
```

## Monitoring

### Key Metrics

Monitor these performance indicators:

- **Scraping Success Rate**: >95%
- **Deduplication Hit Rate**: >90%
- **Processing Throughput**: >100 slots/second
- **Memory Usage**: Stable, no leaks
- **Error Rate**: <5%

### Health Checks

```python
# Check scraper health
from src.utils.health import HealthChecker

health = HealthChecker()
status = health.check_all()

print(f"Redis: {status['redis']}")
print(f"MongoDB: {status['mongodb']}")
print(f"Scrapers: {status['scrapers']}")
```

### Logging

The scraper provides structured logging:

```json
{
  "timestamp": "2025-06-14T10:30:00Z",
  "level": "INFO",
  "component": "scraper_orchestrator",
  "venue_id": "venue123",
  "session_id": "session_456",
  "message": "Scraping completed",
  "metrics": {
    "slots_found": 150,
    "duplicates_skipped": 45,
    "new_slots_stored": 105,
    "duration_seconds": 12.5
  }
}
```

## Troubleshooting

### Common Issues

**Redis Connection Errors**
```bash
# Test Redis connectivity
redis-cli ping

# Check Redis logs
docker logs redis-container
```

**High Memory Usage**
```python
# Check Redis memory usage
import redis
r = redis.Redis()
info = r.info('memory')
print(f"Used memory: {info['used_memory_human']}")
```

**Slow Scraping**
```python
# Enable debug logging
import logging
logging.getLogger('src.scrapers').setLevel(logging.DEBUG)

# Check scraper performance
from src.utils.profiler import profile_scraper
profile_scraper('venue123', date_range)
```

**Duplicate Detection Issues**
```python
# Verify Redis keys
import redis
r = redis.Redis(db=1)
keys = r.keys('dedupe:slot:*')
print(f"Total dedup keys: {len(keys)}")

# Check key expiry
for key in keys[:5]:
    ttl = r.ttl(key)
    print(f"{key}: {ttl}s remaining")
```

## Performance Optimization

### Redis Optimization

```bash
# Redis configuration for optimal performance
redis-cli CONFIG SET maxmemory 256mb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
redis-cli CONFIG SET save ""  # Disable persistence for cache-only usage
```

### Scraper Optimization

```python
# Use connection pooling
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

session = requests.Session()
retry_strategy = Retry(total=3, backoff_factor=1)
adapter = HTTPAdapter(max_retries=retry_strategy, pool_connections=10)
session.mount("http://", adapter)
session.mount("https://", adapter)
```

## Contributing

1. Follow PEP 8 style guidelines
2. Add type hints to new functions
3. Write tests for new features
4. Update documentation
5. Use conventional commits

## License

MIT License - see LICENSE file for details. 