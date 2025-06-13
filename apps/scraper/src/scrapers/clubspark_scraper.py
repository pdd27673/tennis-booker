#!/usr/bin/env python3

"""
ClubSpark platform scraper for Stratford Park.
Based on Firecrawl analysis of the ClubSpark booking system.
"""

import asyncio
import time
import re
from typing import List, Dict, Any
from playwright.async_api import async_playwright, Page, Browser
from .base_scraper import BaseScraper, ScrapedSlot, ScrapingResult

class ClubSparkScraper(BaseScraper):
    """Scraper for ClubSpark platform (Stratford Park)"""
    
    def __init__(self, venue_config: Dict[str, Any]):
        super().__init__(venue_config)
        self.selectors = self.scraper_config.get('selector_mappings', {})
        self.navigation_steps = self.scraper_config.get('navigation_steps', [])
        self.timeout = self.scraper_config.get('timeout_seconds', 45) * 1000
        self.wait_after_load = self.scraper_config.get('wait_after_load_ms', 3000)
        self.custom_params = self.scraper_config.get('custom_parameters', {})
        
    async def scrape_availability(self, target_dates: List[str]) -> ScrapingResult:
        """Scrape court availability for ClubSpark platform"""
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
                            await asyncio.sleep(2)
                            
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
        
        # Navigate to the venue booking page with date
        date_url = self._build_date_url(date)
        self.logger.info(f"Navigating to {date_url}")
        
        await page.goto(date_url, timeout=self.timeout)
        await page.wait_for_timeout(self.wait_after_load)
        
        # Wait for booking sheet to load (ClubSpark specific selector)
        try:
            await page.wait_for_selector('.booking-sheet', timeout=15000)
        except Exception:
            self.logger.warning(f"Booking sheet not found for {date}")
            return slots
            
        # Find all available booking intervals using correct ClubSpark selectors
        available_slots_selector = 'a.book-interval.not-booked'
        available_elements = await page.query_selector_all(available_slots_selector)
        
        self.logger.info(f"Found {len(available_elements)} available slots for {date}")
        
        for slot_element in available_elements:
            try:
                slot_data = await self._extract_slot_data(slot_element, date)
                if slot_data:
                    slots.append(slot_data)
            except Exception as e:
                self.logger.warning(f"Error extracting slot data: {e}")
                
        return slots
        
    async def _extract_slot_data(self, slot_element, date: str) -> ScrapedSlot:
        """Extract slot data from an available booking slot element"""
        
        # Extract price - prioritize data-session-cost from parent resource-session
        price = None
        try:
            # First try to get price from parent resource-session data-session-cost attribute
            resource_session = await slot_element.evaluate_handle("""
                (element) => {
                    let current = element;
                    while (current && current.parentElement) {
                        current = current.parentElement;
                        if (current.classList.contains('resource-session')) {
                            return current;
                        }
                    }
                    return null;
                }
            """)
            
            if resource_session:
                session_cost = await resource_session.get_attribute('data-session-cost')
                if session_cost:
                    price = float(session_cost)
                    
            # Fallback to span.cost if data-session-cost not found
            if price is None:
                cost_element = await slot_element.query_selector('span.cost')
                if cost_element:
                    price_text = await cost_element.text_content()
                    price = self.parse_price(price_text)  # Uses base class method
                    
        except Exception as e:
            self.logger.warning(f"Could not extract price: {e}")
            
        # Extract time information from span.available-booking-slot
        start_time = None
        end_time = None
        try:
            time_element = await slot_element.query_selector('span.available-booking-slot')
            if time_element:
                time_text = await time_element.text_content()
                # Parse "Book at 08:00 - 09:00" format
                time_match = re.search(r'Book at (\d{2}:\d{2}) - (\d{2}:\d{2})', time_text)
                if time_match:
                    start_time = time_match.group(1)
                    end_time = time_match.group(2)
        except Exception as e:
            self.logger.warning(f"Could not extract time: {e}")
            
        if not start_time or not end_time:
            self.logger.warning("Could not parse time information")
            return None
            
        # Extract court information from parent elements
        court_name = "Court 1"  # Default
        court_id = "court-1"    # Default
        
        try:
            # Navigate up to find the court resource container
            # ClubSpark structure: slot is inside .resource-interval -> .resource-session -> .resource
            court_element = await slot_element.evaluate_handle("""
                (element) => {
                    // Walk up the DOM to find the resource container
                    let current = element;
                    while (current && current.parentElement) {
                        current = current.parentElement;
                        if (current.classList.contains('resource')) {
                            return current;
                        }
                    }
                    return null;
                }
            """)
            
            if court_element:
                # Extract court name from data attributes or header
                court_name_attr = await court_element.get_attribute('data-resource-name')
                if court_name_attr:
                    court_name = court_name_attr
                    # Generate court_id from name
                    court_id = re.sub(r'[^a-zA-Z0-9]', '-', court_name.lower())
                else:
                    # Fallback: look for court header
                    header_element = await court_element.query_selector('.resource-header h3')
                    if header_element:
                        court_header_text = await header_element.text_content()
                        if court_header_text:
                            court_name = court_header_text.strip()
                            court_id = re.sub(r'[^a-zA-Z0-9]', '-', court_name.lower())
                            
        except Exception as e:
            self.logger.warning(f"Could not extract court information: {e}")
            
        # Extract booking URL from href
        booking_url = None
        try:
            href = await slot_element.get_attribute('href')
            if href and href != '#':
                # Check if it's a relative URL and make it absolute
                if href.startswith('/') or not href.startswith('http'):
                    base_url = self.url.split('/Booking')[0]
                    booking_url = f"{base_url}{href}"
                else:
                    booking_url = href
        except Exception as e:
            self.logger.warning(f"Could not extract booking URL: {e}")
            
        # Create slot object
        slot = ScrapedSlot(
            venue_id=self.venue_id,
            venue_name=self.venue_name,
            court_id=court_id,
            court_name=court_name,
            date=date,
            start_time=start_time,
            end_time=end_time,
            price=price,
            currency='GBP',
            available=True,
            booking_url=booking_url
        )
        
        return slot
        
    def _build_date_url(self, date: str) -> str:
        """Build URL for specific date on ClubSpark platform"""
        # ClubSpark URL format: /Booking/BookByDate#?date=YYYY-MM-DD&role=guest
        base_url = self.url.split('#')[0]  # Remove any existing fragment
        return f"{base_url}#?date={date}&role=guest"
        
    async def _build_booking_url(self, date: str, start_time: str, court_id: str) -> str:
        """Build booking URL for ClubSpark platform"""
        # This would need to be customized based on ClubSpark's booking flow
        base_url = self.url.split('/Booking')[0]
        return f"{base_url}/Booking/BookSlot?date={date}&time={start_time}&court={court_id}" 