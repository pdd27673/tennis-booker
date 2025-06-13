#!/usr/bin/env python3

"""
Courtside platform scraper for Victoria Park and Ropemakers Field.
Based on Firecrawl analysis of the Courtside booking system.
"""

import asyncio
import time
import re
from typing import List, Dict, Any
from playwright.async_api import async_playwright, Page, Browser
from .base_scraper import BaseScraper, ScrapedSlot, ScrapingResult

class CourtsideScraper(BaseScraper):
    """Scraper for Courtside platform (Victoria Park, Ropemakers Field)"""
    
    def __init__(self, venue_config: Dict[str, Any]):
        super().__init__(venue_config)
        self.selectors = self.scraper_config.get('selector_mappings', {})
        self.navigation_steps = self.scraper_config.get('navigation_steps', [])
        self.timeout = self.scraper_config.get('timeout_seconds', 30) * 1000
        self.wait_after_load = self.scraper_config.get('wait_after_load_ms', 2000)
        
    async def scrape_availability(self, target_dates: List[str]) -> ScrapingResult:
        """Scrape court availability for Courtside platform"""
        start_time = time.time()
        slots = []
        errors = []
        
        try:
            async with async_playwright() as p:
                browser = await p.chromium.launch(
                    headless=self.scraper_config.get('use_headless_browser', True)
                )
                
                try:
                    page = await browser.new_page()
                    await page.set_viewport_size({"width": 1280, "height": 720})
                    
                    # Set user agent if specified
                    user_agent = self.scraper_config.get('user_agent')
                    if user_agent:
                        await page.set_extra_http_headers({'User-Agent': user_agent})
                    
                    for date in target_dates:
                        try:
                            date_slots = await self._scrape_date(page, date)
                            slots.extend(date_slots)
                            
                            # Rate limiting between dates
                            await asyncio.sleep(1)
                            
                        except Exception as e:
                            error_msg = f"Error scraping {date}: {str(e)}"
                            self.logger.error(error_msg)
                            errors.append(error_msg)
                            
                finally:
                    await browser.close()
                    
        except Exception as e:
            error_msg = f"Browser setup error: {str(e)}"
            self.logger.error(error_msg)
            errors.append(error_msg)
            
        duration_ms = int((time.time() - start_time) * 1000)
        success = len(errors) == 0
        
        self.logger.info(f"Scraped {len(slots)} slots for {self.venue_name} in {duration_ms}ms")
        
        return self.create_scraping_result(success, slots, errors, duration_ms)
        
    async def _scrape_date(self, page: Page, date: str) -> List[ScrapedSlot]:
        """Scrape availability for a specific date"""
        slots = []
        
        # Navigate to the venue booking page
        date_url = self._build_date_url(date)
        self.logger.info(f"Navigating to {date_url}")
        
        await page.goto(date_url, timeout=self.timeout)
        await page.wait_for_timeout(self.wait_after_load)
        
        # Check if courts are closed
        closed_selector = self.selectors.get('closed_message', '.closed-today')
        closed_elements = await page.query_selector_all(closed_selector)
        
        if closed_elements:
            self.logger.info(f"Courts closed on {date}")
            return slots
            
        # Wait for court widget to load
        court_widget_selector = self.selectors.get('court_widget', '.court-widget')
        try:
            await page.wait_for_selector(court_widget_selector, timeout=10000)
        except Exception:
            self.logger.warning(f"Court widget not found for {date}")
            return slots
            
        # Extract available time slots - NEW LOGIC based on Firecrawl analysis
        # Look for checkboxes with class "bookable" which indicate available slots
        available_checkboxes = await page.query_selector_all('input.bookable')
        
        self.logger.info(f"Found {len(available_checkboxes)} available slots for {date}")
        
        for checkbox in available_checkboxes:
            try:
                slot_data = await self._extract_slot_data_from_checkbox(page, checkbox, date)
                if slot_data:
                    slots.append(slot_data)
            except Exception as e:
                self.logger.warning(f"Error extracting slot data: {e}")
                
        return slots
        
    async def _extract_slot_data_from_checkbox(self, page: Page, checkbox, date: str) -> ScrapedSlot:
        """Extract slot data from an available checkbox element"""
        
        # Get the checkbox value which contains venue_id, court_id, date, and time
        # Format: "254_164_2025-06-10_14:00" 
        checkbox_value = await checkbox.get_attribute('value')
        if not checkbox_value:
            return None
            
        # Parse the checkbox value
        value_parts = checkbox_value.split('_')
        if len(value_parts) < 4:
            self.logger.warning(f"Unexpected checkbox value format: {checkbox_value}")
            return None
            
        venue_id_part = value_parts[0]
        court_id_part = value_parts[1] 
        date_part = value_parts[2]
        time_part = value_parts[3]
        
        # Extract price from data-price attribute
        price_str = await checkbox.get_attribute('data-price')
        price = float(price_str) if price_str else None
        
        # Find the parent row to get the time
        time_cell = await checkbox.evaluate("""
            (checkbox) => {
                const row = checkbox.closest('tr');
                if (row) {
                    const timeCell = row.querySelector('th.time');
                    return timeCell ? timeCell.textContent.trim() : null;
                }
                return null;
            }
        """)
        
        # Parse time (e.g., "2pm" -> "14:00-15:00")
        start_time, end_time = self._parse_courtside_time(time_cell) if time_cell else (None, None)
        if not start_time:
            # Fallback to parsing from checkbox value time part
            start_time, end_time = self._parse_time_from_value(time_part)
            
        # Find the court name from the button text
        court_name = await checkbox.evaluate("""
            (checkbox) => {
                const label = checkbox.closest('label');
                if (label) {
                    const button = label.querySelector('span.button.available');
                    if (button) {
                        const text = button.childNodes[0].textContent;
                        return text ? text.trim() : null;
                    }
                }
                return null;
            }
        """)
        
        if not court_name:
            court_name = f"Court {court_id_part}"
            
        court_id = court_name.lower().replace(" ", "-")
        
        # Build booking URL - use the existing URL structure
        booking_url = self._build_booking_url_from_value(checkbox_value)
        
        return ScrapedSlot(
            venue_id=self.venue_id,
            venue_name=self.venue_name,
            court_id=court_id,
            court_name=f"{self.venue_name} {court_name}",
            date=date,
            start_time=start_time,
            end_time=end_time,
            price=price,
            currency="GBP", 
            available=True,
            booking_url=booking_url
        )
        
    def _parse_courtside_time(self, time_text: str) -> tuple:
        """Parse Courtside time format like '2pm' to '14:00-15:00'"""
        if not time_text:
            return None, None
            
        # Remove any extra whitespace
        time_text = time_text.strip().lower()
        
        # Handle am/pm format
        if 'am' in time_text or 'pm' in time_text:
            # Extract the hour
            hour_match = re.search(r'(\d+)', time_text)
            if not hour_match:
                return None, None
                
            hour = int(hour_match.group(1))
            
            # Convert to 24-hour format
            if 'pm' in time_text and hour != 12:
                hour += 12
            elif 'am' in time_text and hour == 12:
                hour = 0
                
            start_time = f"{hour:02d}:00"
            end_time = f"{(hour + 1):02d}:00"
            
            return start_time, end_time
            
        return None, None
        
    def _parse_time_from_value(self, time_part: str) -> tuple:
        """Parse time from checkbox value like '14:00' to '14:00-15:00'"""
        if not time_part or ':' not in time_part:
            return None, None
            
        try:
            hour, minute = time_part.split(':')
            hour = int(hour)
            start_time = f"{hour:02d}:{minute}"
            end_time = f"{(hour + 1):02d}:{minute}"
            return start_time, end_time
        except:
            return None, None
            
    def _build_date_url(self, date: str) -> str:
        """Build URL for specific date"""
        # Courtside URLs typically follow pattern: /book/courts/venue-name/YYYY-MM-DD#book
        base_url = self.url.split('#')[0]  # Remove fragment if present
        return f"{base_url}/{date}#book"
        
    def _build_booking_url_from_value(self, checkbox_value: str) -> str:
        """Build booking URL from checkbox value"""
        # For now, return the base booking page
        # In a real implementation, you'd want to construct the actual booking URL
        base_url = self.url.split('#')[0]
        return f"{base_url}?booking={checkbox_value}" 