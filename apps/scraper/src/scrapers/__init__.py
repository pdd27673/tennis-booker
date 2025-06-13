"""
Tennis court scrapers package.

This package contains platform-specific scrapers for different tennis booking systems:
- CourtsideScraper: For Courtside platform (Victoria Park, Ropemakers Field)
- ClubSparkScraper: For ClubSpark platform (Stratford Park)
"""

from .base_scraper import BaseScraper, ScrapedSlot, ScrapingResult
from .courtside_scraper import CourtsideScraper
from .clubspark_scraper import ClubSparkScraper
from .scraper_orchestrator import ScraperOrchestrator

__all__ = [
    'BaseScraper',
    'ScrapedSlot', 
    'ScrapingResult',
    'CourtsideScraper',
    'ClubSparkScraper',
    'ScraperOrchestrator'
] 