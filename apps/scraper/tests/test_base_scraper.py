import sys
import os
import pytest
from unittest.mock import Mock, patch
from datetime import datetime, timedelta

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

    def test_get_target_dates_default(self):
        """Test get_target_dates with default 7 days ahead."""
        venue = {
            "_id": "test_venue_id",
            "name": "Test Venue",
            "url": "https://example.com",
            "courts": [],
            "scraper_config": {"type": "test_provider"}
        }
        
        scraper = ConcreteScraper(venue)
        target_dates = scraper.get_target_dates()
        
        # Should return 7 dates by default
        assert len(target_dates) == 7
        
        # Verify the dates are correct (starting from today)
        today = datetime.now().date()
        for i, date_str in enumerate(target_dates):
            expected_date = today + timedelta(days=i)
            assert date_str == expected_date.strftime("%Y-%m-%d")

    def test_get_target_dates_custom_days(self):
        """Test get_target_dates with custom days ahead parameter."""
        venue = {
            "_id": "test_venue_id",
            "name": "Test Venue",
            "url": "https://example.com",
            "courts": [],
            "scraper_config": {"type": "test_provider"}
        }
        
        scraper = ConcreteScraper(venue)
        
        # Test with 8 days ahead (new requirement)
        target_dates = scraper.get_target_dates(days_ahead=8)
        assert len(target_dates) == 8
        
        # Test with 3 days ahead
        target_dates_3 = scraper.get_target_dates(days_ahead=3)
        assert len(target_dates_3) == 3
        
        # Test with 1 day ahead
        target_dates_1 = scraper.get_target_dates(days_ahead=1)
        assert len(target_dates_1) == 1
        
        # Verify the dates are correct for 8 days
        today = datetime.now().date()
        for i, date_str in enumerate(target_dates):
            expected_date = today + timedelta(days=i)
            assert date_str == expected_date.strftime("%Y-%m-%d")

    def test_get_target_dates_format(self):
        """Test that get_target_dates returns proper YYYY-MM-DD format."""
        venue = {
            "_id": "test_venue_id",
            "name": "Test Venue",
            "url": "https://example.com",
            "courts": [],
            "scraper_config": {"type": "test_provider"}
        }
        
        scraper = ConcreteScraper(venue)
        target_dates = scraper.get_target_dates(days_ahead=5)
        
        # Verify all dates are in YYYY-MM-DD format
        for date_str in target_dates:
            assert len(date_str) == 10
            assert date_str[4] == '-'
            assert date_str[7] == '-'
            # Verify it can be parsed as a date
            datetime.strptime(date_str, "%Y-%m-%d") 