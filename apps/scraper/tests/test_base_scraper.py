import sys
import os
import pytest
from unittest.mock import Mock, patch

# Add the src directory to the Python path
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'src'))

from scrapers.base_scraper import BaseScraper

# Create a concrete implementation for testing
class ConcreteScraper(BaseScraper):
    async def scrape_availability(self, target_dates):
        """Concrete implementation for testing."""
        return self.create_scraping_result(True, [], [], 0)

class TestBaseScraper:
    def test_base_scraper_init(self):
        """Test that the BaseScraper initializes correctly."""
        venue = {
            "_id": "test_venue_id",
            "name": "Test Venue",
            "url": "https://example.com",
            "courts": [
                {"id": "court1", "name": "Court 1"},
                {"id": "court2", "name": "Court 2"}
            ],
            "scraper_config": {
                "type": "test_provider",
                "timeoutSeconds": 30,
                "retryCount": 3,
                "waitAfterLoadMs": 2000,
                "useHeadlessBrowser": True
            }
        }
        
        # Use concrete implementation instead of abstract class
        scraper = ConcreteScraper(venue)
        
        assert scraper.venue_id == "test_venue_id"
        assert scraper.venue_name == "Test Venue"
        assert scraper.url == "https://example.com"
        assert scraper.platform == "test_provider"
        assert len(scraper.courts) == 2
        assert scraper.scraper_config["timeoutSeconds"] == 30 