import sys
import os
import pytest
from unittest.mock import Mock, patch

# Add the src directory to the Python path
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'src'))

from scrapers.base_scraper import BaseScraper

class TestBaseScraper:
    def test_base_scraper_init(self):
        """Test that the BaseScraper initializes correctly."""
        venue = {
            "name": "Test Venue",
            "url": "https://example.com",
            "provider": "test_provider",
            "scraperConfig": {
                "timeoutSeconds": 30,
                "retryCount": 3,
                "waitAfterLoadMs": 2000,
                "useHeadlessBrowser": True
            }
        }
        
        scraper = BaseScraper(venue)
        
        assert scraper.venue_name == "Test Venue"
        assert scraper.venue_url == "https://example.com"
        assert scraper.provider == "test_provider"
        assert scraper.timeout_seconds == 30
        assert scraper.retry_count == 3
        assert scraper.wait_after_load_ms == 2000
        assert scraper.use_headless == True 