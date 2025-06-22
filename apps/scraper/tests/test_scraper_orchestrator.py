import sys
import os
import pytest
from unittest.mock import Mock, patch, AsyncMock
from datetime import datetime, timedelta

# Add the src directory to the Python path
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'src'))

from scrapers.scraper_orchestrator import ScraperOrchestrator

class TestScraperOrchestrator:
    
    @patch.dict(os.environ, {}, clear=True)
    def test_scraper_days_ahead_default(self):
        """Test that SCRAPER_DAYS_AHEAD defaults to 7 when not set."""
        # Create a mock venue configuration
        venue_config = {
            '_id': 'test_venue_id',
            'name': 'Test Venue',
            'url': 'https://example.com',
            'courts': [],
            'scraper_config': {'type': 'courtside'}
        }
        
        # Mock the base scraper to capture the days_ahead parameter
        with patch('scrapers.scraper_orchestrator.CourtsideScraper') as mock_scraper_class:
            mock_scraper = Mock()
            mock_scraper.get_target_dates.return_value = ['2024-01-01', '2024-01-02', '2024-01-03', '2024-01-04', '2024-01-05', '2024-01-06', '2024-01-07']
            mock_scraper.scrape_availability = AsyncMock()
            mock_scraper_class.return_value = mock_scraper
            
            orchestrator = ScraperOrchestrator()
            
            # This should use the default value of 7 since SCRAPER_DAYS_AHEAD is not set
            # We can't easily test the exact environment variable read, but we can verify the behavior
            # by ensuring the mock get_target_dates is called (which happens when target_dates is None)
            import asyncio
            asyncio.run(orchestrator.scrape_venue(venue_config, target_dates=None))
            
            # Verify that get_target_dates was called (meaning target_dates was None and default was used)
            mock_scraper.get_target_dates.assert_called_once()
            # The default should be 7 days, so we check that the call was made with that parameter
            call_args = mock_scraper.get_target_dates.call_args
            assert call_args[1]['days_ahead'] == 7  # Default value when env var not set

    @patch.dict(os.environ, {'SCRAPER_DAYS_AHEAD': '8'}, clear=True)
    def test_scraper_days_ahead_environment_variable(self):
        """Test that SCRAPER_DAYS_AHEAD environment variable is respected."""
        venue_config = {
            '_id': 'test_venue_id',
            'name': 'Test Venue',
            'url': 'https://example.com',
            'courts': [],
            'scraper_config': {'type': 'courtside'}
        }
        
        with patch('scrapers.scraper_orchestrator.CourtsideScraper') as mock_scraper_class:
            mock_scraper = Mock()
            mock_scraper.get_target_dates.return_value = ['2024-01-01', '2024-01-02', '2024-01-03', '2024-01-04', '2024-01-05', '2024-01-06', '2024-01-07', '2024-01-08']
            mock_scraper.scrape_availability = AsyncMock()
            mock_scraper_class.return_value = mock_scraper
            
            orchestrator = ScraperOrchestrator()
            
            import asyncio
            asyncio.run(orchestrator.scrape_venue(venue_config, target_dates=None))
            
            # Verify that get_target_dates was called with 8 days
            mock_scraper.get_target_dates.assert_called_once()
            call_args = mock_scraper.get_target_dates.call_args
            assert call_args[1]['days_ahead'] == 8  # Environment variable value

    @patch.dict(os.environ, {'SCRAPER_DAYS_AHEAD': '10'}, clear=True)
    def test_scraper_days_ahead_custom_value(self):
        """Test that SCRAPER_DAYS_AHEAD works with any custom value."""
        venue_config = {
            '_id': 'test_venue_id',
            'name': 'Test Venue',
            'url': 'https://example.com',
            'courts': [],
            'scraper_config': {'type': 'courtside'}
        }
        
        with patch('scrapers.scraper_orchestrator.CourtsideScraper') as mock_scraper_class:
            mock_scraper = Mock()
            # Mock 10 days worth of dates
            mock_dates = [f'2024-01-{i:02d}' for i in range(1, 11)]
            mock_scraper.get_target_dates.return_value = mock_dates
            mock_scraper.scrape_availability = AsyncMock()
            mock_scraper_class.return_value = mock_scraper
            
            orchestrator = ScraperOrchestrator()
            
            import asyncio
            asyncio.run(orchestrator.scrape_venue(venue_config, target_dates=None))
            
            # Verify that get_target_dates was called with 10 days
            mock_scraper.get_target_dates.assert_called_once()
            call_args = mock_scraper.get_target_dates.call_args
            assert call_args[1]['days_ahead'] == 10

    @patch.dict(os.environ, {'SCRAPER_DAYS_AHEAD': '8'}, clear=True)
    def test_test_single_venue_days_ahead(self):
        """Test that test_single_venue method also respects SCRAPER_DAYS_AHEAD."""
        with patch('scrapers.scraper_orchestrator.CourtsideScraper') as mock_scraper_class:
            mock_scraper = Mock()
            mock_scraper.get_target_dates.return_value = ['2024-01-01', '2024-01-02', '2024-01-03', '2024-01-04', '2024-01-05', '2024-01-06', '2024-01-07', '2024-01-08']
            mock_scraper.scrape_availability = AsyncMock()
            mock_scraper_class.return_value = mock_scraper
            
            orchestrator = ScraperOrchestrator()
            
            # Mock the database operations
            with patch.object(orchestrator, 'connect_mongodb'), \
                 patch.object(orchestrator, 'disconnect_mongodb'), \
                 patch.object(orchestrator, 'load_venues') as mock_load_venues, \
                 patch.object(orchestrator, 'store_scraping_result') as mock_store:
                
                mock_load_venues.return_value = [{
                    '_id': 'test_venue_id',
                    'name': 'Test Venue',
                    'url': 'https://example.com',
                    'courts': [],
                    'scraper_config': {'type': 'courtside'}
                }]
                
                import asyncio
                asyncio.run(orchestrator.test_single_venue('Test Venue', days_ahead=None))
                
                # Verify that get_target_dates was called with 8 days (from environment)
                mock_scraper.get_target_dates.assert_called_once()
                call_args = mock_scraper.get_target_dates.call_args
                assert call_args[1]['days_ahead'] == 8 