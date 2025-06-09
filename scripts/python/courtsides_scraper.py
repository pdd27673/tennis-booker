#!/usr/bin/env python3
"""
Courtside/Tower Hamlets Tennis Court Scraper

This module handles scraping court availability from Courtside-managed venues
like those in Tower Hamlets. Based on the website structure analysis:

Website Structure:
- Time slots are shown in a table format with rows for each hour (5pm, 6pm, 7pm, 8pm)
- Each row contains cells for different courts (Court 1, Court 2, etc.)
- Court status can be: "booked", "Reserved" (for group coaching), or available
- Date navigation allows checking multiple days

URL Pattern: https://tennistowerhamlets.com/book/courts/{venue-name}#book
"""

import re
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Any, Optional
from urllib.parse import urljoin, urlparse

import requests
from bs4 import BeautifulSoup
from firecrawl import FirecrawlApp


class CourtsidesVenueScraper:
    """
    Scraper for Courtside-managed tennis venues.
    
    Handles the specific HTML structure and booking interface used by
    Courtside for their Tower Hamlets venues.
    """
    
    def __init__(self, firecrawl_client: Optional[FirecrawlApp] = None):
        """
        Initialize the Courtside scraper.
        
        Args:
            firecrawl_client: Optional Firecrawl client for JavaScript rendering
        """
        self.logger = logging.getLogger(__name__)
        self.firecrawl_client = firecrawl_client
        
        # Session for cookie persistence if needed
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8',
            'Accept-Language': 'en-US,en;q=0.5',
            'Accept-Encoding': 'gzip, deflate',
            'Connection': 'keep-alive',
            'Upgrade-Insecure-Requests': '1'
        })
    
    def scrape_venue_availability(self, venue_config: Dict[str, Any], 
                                 target_date: Optional[datetime] = None) -> Dict[str, Any]:
        """
        Scrape court availability for a Courtside venue.
        
        Args:
            venue_config: Venue configuration dictionary
            target_date: Date to check (defaults to today)
            
        Returns:
            Dictionary containing scraped availability data
        """
        venue_id = venue_config.get('_id')
        venue_name = venue_config.get('name')
        base_url = venue_config.get('url')
        
        if not base_url:
            self.logger.error(f"No URL provided for venue {venue_name}")
            return self._create_error_result(venue_id, "No URL provided")
        
        if target_date is None:
            target_date = datetime.now()
        
        self.logger.info(f"Scraping {venue_name} for {target_date.strftime('%Y-%m-%d')}")
        
        try:
            # Construct URL with date parameter if needed
            scrape_url = self._build_date_specific_url(base_url, target_date)
            
            # Use requests for scraping (Firecrawl analysis already done)
            content = self._scrape_with_requests(scrape_url)
            
            # Only use Firecrawl as fallback if requests fails and it's available
            if not content and self.firecrawl_client:
                self.logger.info("Requests failed, trying Firecrawl as fallback")
                content = self._scrape_with_firecrawl(scrape_url)
            
            if not content:
                return self._create_error_result(venue_id, "Failed to retrieve content")
            
            # Parse the HTML content
            soup = BeautifulSoup(content, 'html.parser')
            
            # Extract availability data
            availability_data = self._parse_availability_table(soup, venue_config)
            
            # Structure the result
            result = {
                'venue_id': venue_id,
                'venue_name': venue_name,
                'scrape_date': datetime.now().isoformat(),
                'target_date': target_date.isoformat(),
                'url': scrape_url,
                'status': 'success',
                'court_availability': availability_data.get('courts', []),
                'time_slots': availability_data.get('time_slots', []),
                'metadata': {
                    'total_courts': len(venue_config.get('courts', [])),
                    'total_slots_checked': len(availability_data.get('time_slots', [])),
                    'scraper_type': 'courtsides'
                }
            }
            
            self.logger.info(f"Successfully scraped {venue_name}: {len(result['court_availability'])} courts")
            return result
            
        except Exception as e:
            self.logger.error(f"Error scraping venue {venue_name}: {e}")
            return self._create_error_result(venue_id, str(e))
    
    def _scrape_with_firecrawl(self, url: str) -> Optional[str]:
        """
        Scrape content using Firecrawl for JavaScript rendering.
        
        Args:
            url: URL to scrape
            
        Returns:
            HTML content or None if failed
        """
        try:
            self.logger.debug(f"Scraping with Firecrawl: {url}")
            
            result = self.firecrawl_client.scrape_url(
                url,
                params={
                    'formats': ['html'],
                    'waitFor': 3000,  # Wait 3 seconds for JS to load
                    'timeout': 10000
                }
            )
            
            if result and 'html' in result:
                return result['html']
            else:
                self.logger.warning("Firecrawl returned no HTML content")
                return None
                
        except Exception as e:
            self.logger.warning(f"Firecrawl scraping failed: {e}")
            return None
    
    def _scrape_with_requests(self, url: str) -> Optional[str]:
        """
        Scrape content using requests as fallback.
        
        Args:
            url: URL to scrape
            
        Returns:
            HTML content or None if failed
        """
        try:
            self.logger.debug(f"Scraping with requests: {url}")
            
            response = self.session.get(url, timeout=10)
            response.raise_for_status()
            
            return response.text
            
        except Exception as e:
            self.logger.warning(f"Requests scraping failed: {e}")
            return None
    
    def _build_date_specific_url(self, base_url: str, target_date: datetime) -> str:
        """
        Build URL for specific date if the venue supports date navigation.
        
        Args:
            base_url: Base venue URL
            target_date: Target date for booking
            
        Returns:
            URL with date parameters if applicable
        """
        # For now, return base URL - date navigation logic can be added later
        # when we analyze the actual date selection mechanism
        return base_url
    
    def _parse_availability_table(self, soup: BeautifulSoup, 
                                 venue_config: Dict[str, Any]) -> Dict[str, Any]:
        """
        Parse the booking table to extract court availability.
        
        Args:
            soup: BeautifulSoup object of the page
            venue_config: Venue configuration
            
        Returns:
            Dictionary with parsed availability data
        """
        courts_data = []
        time_slots = []
        
        # Expected venue courts from config
        expected_courts = venue_config.get('courts', [])
        
        try:
            # Find the booking table - look for the availability form table
            booking_table = soup.find('table')
            
            if not booking_table:
                self.logger.warning("No booking table found")
                return {'courts': courts_data, 'time_slots': time_slots}
            
            # Find all time slot rows
            time_rows = booking_table.find_all('tr')
            
            for row in time_rows:
                # Extract time slot from th.time cell
                time_cell = row.find('th', class_='time')
                if not time_cell:
                    continue
                
                time_text = time_cell.get_text(strip=True)
                
                # Match time patterns like "5pm", "6pm", etc.
                time_match = re.search(r'(\d+)(am|pm)', time_text.lower())
                if not time_match:
                    continue
                
                time_slot = time_text  # Keep original format
                time_slots.append(time_slot)
                
                # Find the courts cell (td.courts)
                courts_cell = row.find('td', class_='courts')
                if not courts_cell:
                    continue
                
                # Find all court labels within this cell
                court_labels = courts_cell.find_all('label', class_='court')
                
                for i, court_label in enumerate(court_labels):
                    if i >= len(expected_courts):
                        break  # Don't exceed expected court count
                    
                    court_info = expected_courts[i]
                    court_id = court_info['id']
                    court_name = court_info['name']
                    
                    # Get the button element to check status
                    button = court_label.find('span', class_='button')
                    if not button:
                        continue
                    
                    button_text = button.get_text(strip=True)
                    button_classes = button.get('class', [])
                    
                    # Determine availability status based on button content and classes
                    session_info = None
                    
                    if 'booked' in button_classes:
                        status = 'booked'
                        available = False
                    elif 'session' in button_classes or 'Reserved' in button_text:
                        status = 'reserved'
                        available = False
                        # Extract session info from the link
                        session_link = button.find('a')
                        if session_link:
                            session_info = session_link.get_text(strip=True)
                        else:
                            session_info = 'Group coaching'
                    else:
                        # If no special classes, check if input is disabled
                        input_elem = court_label.find('input')
                        if input_elem and input_elem.get('disabled'):
                            status = 'unavailable'
                            available = False
                        else:
                            status = 'available'
                            available = True
                    
                    # Create court availability entry
                    court_data = {
                        'court_id': court_id,
                        'court_name': court_name,
                        'time_slot': time_slot,
                        'available': available,
                        'status': status,
                        'session_info': session_info,
                        'raw_text': button_text  # For debugging
                    }
                    
                    courts_data.append(court_data)
            
            self.logger.debug(f"Parsed {len(courts_data)} court slots across {len(time_slots)} time slots")
            
        except Exception as e:
            self.logger.error(f"Error parsing availability table: {e}")
        
        return {
            'courts': courts_data,
            'time_slots': list(set(time_slots))  # Remove duplicates
        }
    
    def _create_error_result(self, venue_id: str, error_message: str) -> Dict[str, Any]:
        """
        Create standardized error result.
        
        Args:
            venue_id: Venue identifier
            error_message: Error description
            
        Returns:
            Error result dictionary
        """
        return {
            'venue_id': venue_id,
            'scrape_date': datetime.now().isoformat(),
            'status': 'error',
            'error': error_message,
            'court_availability': [],
            'time_slots': [],
            'metadata': {
                'scraper_type': 'courtsides'
            }
        }
    
    def get_supported_venues(self) -> List[str]:
        """
        Get list of venue types supported by this scraper.
        
        Returns:
            List of supported venue provider types
        """
        return ['courtsides', 'tower_hamlets']


def test_courtside_scraper():
    """Test function for the Courtside scraper."""
    logging.basicConfig(level=logging.DEBUG)
    
    # Test venue configuration
    test_venue = {
        '_id': 'ropemakers_field',
        'name': 'Ropemakers Field',
        'provider': 'courtsides',
        'url': 'https://tennistowerhamlets.com/book/courts/ropemakers-field#book',
        'courts': [
            {'id': 'court_1', 'name': 'Court 1'},
            {'id': 'court_2', 'name': 'Court 2'}
        ]
    }
    
    scraper = CourtsidesVenueScraper()
    result = scraper.scrape_venue_availability(test_venue)
    
    print(f"Scraping result: {result}")
    return result


if __name__ == "__main__":
    test_courtside_scraper() 