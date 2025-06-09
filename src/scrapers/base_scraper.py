#!/usr/bin/env python3

"""
Base scraper class for tennis court availability scraping.
Defines the interface that platform-specific scrapers must implement.
"""

import logging
import time
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import List, Dict, Any, Optional
from datetime import datetime, timedelta

@dataclass
class ScrapedSlot:
    """Represents a scraped tennis court time slot"""
    venue_id: str
    venue_name: str
    court_id: str
    court_name: str
    date: str  # YYYY-MM-DD format
    start_time: str  # HH:MM format
    end_time: str  # HH:MM format
    price: Optional[float]
    currency: str = "GBP"
    available: bool = True
    booking_url: Optional[str] = None
    scraped_at: datetime = None
    
    def __post_init__(self):
        if self.scraped_at is None:
            self.scraped_at = datetime.now()

@dataclass
class ScrapingResult:
    """Result of a scraping operation"""
    venue_id: str
    venue_name: str
    platform: str
    success: bool
    slots_found: List[ScrapedSlot]
    errors: List[str]
    duration_ms: int
    scraped_at: datetime
    
class BaseScraper(ABC):
    """Base class for platform-specific scrapers"""
    
    def __init__(self, venue_config: Dict[str, Any]):
        self.venue_config = venue_config
        self.venue_id = venue_config['_id']
        self.venue_name = venue_config['name']
        self.platform = venue_config['scraper_config']['type']
        self.url = venue_config['url']
        self.courts = venue_config['courts']
        self.scraper_config = venue_config['scraper_config']
        
        self.logger = logging.getLogger(f"{self.__class__.__name__}_{self.venue_name}")
        
    @abstractmethod
    async def scrape_availability(self, target_dates: List[str]) -> ScrapingResult:
        """
        Scrape court availability for the given dates.
        
        Args:
            target_dates: List of dates in YYYY-MM-DD format
            
        Returns:
            ScrapingResult with found slots and metadata
        """
        pass
        
    def get_target_dates(self, days_ahead: int = 7) -> List[str]:
        """Get list of target dates to scrape (default 7 days ahead)"""
        target_dates = []
        today = datetime.now().date()
        
        for i in range(days_ahead):
            target_date = today + timedelta(days=i)
            target_dates.append(target_date.strftime("%Y-%m-%d"))
            
        return target_dates
        
    def parse_time_slot(self, time_text: str) -> tuple[str, str]:
        """
        Parse time slot text into start and end times.
        
        Args:
            time_text: Time text like "14:00-15:00" or "2:00 PM - 3:00 PM"
            
        Returns:
            Tuple of (start_time, end_time) in HH:MM format
        """
        import re
        
        # Handle 24-hour format: "14:00-15:00"
        match = re.match(r'(\d{1,2}):(\d{2})-(\d{1,2}):(\d{2})', time_text)
        if match:
            start_hour, start_min, end_hour, end_min = match.groups()
            return f"{int(start_hour):02d}:{start_min}", f"{int(end_hour):02d}:{end_min}"
            
        # Handle 12-hour format: "2:00 PM - 3:00 PM"
        match = re.match(r'(\d{1,2}):(\d{2})\s*(AM|PM)\s*-\s*(\d{1,2}):(\d{2})\s*(AM|PM)', time_text, re.IGNORECASE)
        if match:
            start_hour, start_min, start_period, end_hour, end_min, end_period = match.groups()
            
            # Convert to 24-hour format
            start_hour = int(start_hour)
            end_hour = int(end_hour)
            
            if start_period.upper() == 'PM' and start_hour != 12:
                start_hour += 12
            elif start_period.upper() == 'AM' and start_hour == 12:
                start_hour = 0
                
            if end_period.upper() == 'PM' and end_hour != 12:
                end_hour += 12
            elif end_period.upper() == 'AM' and end_hour == 12:
                end_hour = 0
                
            return f"{start_hour:02d}:{start_min}", f"{end_hour:02d}:{end_min}"
            
        # Fallback: assume 1-hour slot if only start time given
        match = re.match(r'(\d{1,2}):(\d{2})', time_text)
        if match:
            start_hour, start_min = match.groups()
            start_hour = int(start_hour)
            end_hour = start_hour + 1
            return f"{start_hour:02d}:{start_min}", f"{end_hour:02d}:{start_min}"
            
        self.logger.warning(f"Could not parse time slot: {time_text}")
        return "00:00", "01:00"
        
    def parse_price(self, price_text: str) -> Optional[float]:
        """
        Parse price from text.
        
        Args:
            price_text: Price text like "£25.00" or "$30"
            
        Returns:
            Price as float or None if not found
        """
        import re
        
        if not price_text:
            return None
            
        # Remove currency symbols and extract number
        match = re.search(r'[\d,]+\.?\d*', price_text.replace(',', ''))
        if match:
            try:
                return float(match.group())
            except ValueError:
                pass
                
        return None
        
    def create_scraping_result(self, success: bool, slots: List[ScrapedSlot], 
                             errors: List[str], duration_ms: int) -> ScrapingResult:
        """Create a ScrapingResult object"""
        return ScrapingResult(
            venue_id=self.venue_id,
            venue_name=self.venue_name,
            platform=self.platform,
            success=success,
            slots_found=slots,
            errors=errors,
            duration_ms=duration_ms,
            scraped_at=datetime.now()
        ) 