# Tennis Court Scraper

This module contains Python scrapers for various tennis court booking platforms.

## Supported Platforms

- ClubSpark (LTA)
- Courtside

## Directory Structure

```
scraper/
├── src/
│   ├── scrapers/
│   │   ├── __init__.py
│   │   ├── base_scraper.py
│   │   ├── clubspark_scraper.py
│   │   ├── courtside_scraper.py
│   │   └── scraper_orchestrator.py
│   └── redis_publisher.py
├── tests/
│   └── test_base_scraper.py
├── Makefile
└── requirements.txt
```

## Setup

```bash
# Create virtual environment and install dependencies
make setup
```

## Usage

```bash
# Run the scraper
make run
```

## Testing

```bash
# Run tests
make test
```

## Development

To add support for a new booking platform:

1. Create a new scraper class that inherits from `BaseScraper`
2. Implement the required methods: `scrape_slots()` and `parse_slots()`
3. Add the new scraper to the `scraper_orchestrator.py` 