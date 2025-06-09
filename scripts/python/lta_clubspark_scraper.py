#!/usr/bin/env python3
"""
LTA Clubspark venue scraper for tennis court availability
"""

import re
import requests
import logging
from datetime import datetime, timedelta
from typing import Dict, Any, List, Optional
from urllib.parse import urljoin, urlparse
from bs4 import BeautifulSoup


class LTAClubsparkScraper:
    """Scraper for LTA Clubspark tennis booking websites."""
    
    def __init__(self, firecrawl_client=None):
        """
        Initialize the LTA Clubspark scraper.
        
        Args:
            firecrawl_client: Optional Firecrawl client for JavaScript-heavy sites
        """
        self.firecrawl_client = firecrawl_client
        self.logger = logging.getLogger(__name__)
        
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
        Scrape court availability for an LTA Clubspark venue.
        
        Args:
            venue_config: Venue configuration dictionary
            target_date: Date to check (defaults to today)
            
        Returns:
            Dictionary with scraping results
        """
        if target_date is None:
            target_date = datetime.now()
        
        venue_id = venue_config.get('_id')
        venue_name = venue_config.get('name')
        base_url = venue_config.get('url')
        
        self.logger.info(f"Scraping LTA Clubspark venue: {venue_name} for {target_date.date()}")
        
        try:
            # Build date-specific URL
            scrape_url = self._build_date_specific_url(base_url, target_date)
            self.logger.debug(f"Scraping URL: {scrape_url}")
            
            # LTA Clubspark sites require JavaScript, so use Firecrawl primarily
            if self.firecrawl_client:
                self.logger.info("Using Firecrawl for JavaScript-heavy LTA site")
                content = self._scrape_with_firecrawl(scrape_url)
            else:
                self.logger.warning("No Firecrawl client available, trying requests (may not work for LTA sites)")
                content = self._scrape_with_requests(scrape_url)
            
            if not content:
                return {
                    'venue_id': venue_id,
                    'status': 'error',
                    'error': 'Failed to retrieve page content',
                    'scrape_date': datetime.now().isoformat()
                }
            
            # Parse the content
            court_availability = self._parse_lta_clubspark_content(content, venue_config)
            
            # Extract time slots
            time_slots = list(set([slot['time_slot'] for slot in court_availability]))
            time_slots.sort()
            
            result = {
                'venue_id': venue_id,
                'venue_name': venue_name,
                'status': 'success',
                'scrape_date': datetime.now().isoformat(),
                'target_date': target_date.date().isoformat(),
                'time_slots': time_slots,
                'court_availability': court_availability,
                'total_slots': len(court_availability)
            }
            
            self.logger.info(f"Successfully scraped {len(court_availability)} court slots from {venue_name}")
            return result
            
        except Exception as e:
            self.logger.error(f"Error scraping LTA Clubspark venue {venue_name}: {e}")
            return {
                'venue_id': venue_id,
                'status': 'error',
                'error': str(e),
                'scrape_date': datetime.now().isoformat()
            }
    
    def _build_date_specific_url(self, base_url: str, target_date: datetime) -> str:
        """Build URL with specific date parameter for LTA Clubspark."""
        try:
            # LTA Clubspark URLs typically use #?date=YYYY-MM-DD format
            date_str = target_date.strftime('%Y-%m-%d')
            
            # Remove existing date parameters and add new one
            if '#' in base_url:
                base_part = base_url.split('#')[0]
            else:
                base_part = base_url
            
            # Add date parameter
            url = f"{base_part}#?date={date_str}&role=guest"
            
            self.logger.debug(f"Built date-specific URL: {url}")
            return url
            
        except Exception as e:
            self.logger.warning(f"Failed to build date-specific URL: {e}, using base URL")
            return base_url
    
    def _scrape_with_requests(self, url: str) -> Optional[str]:
        """Scrape content using requests and BeautifulSoup."""
        try:
            self.logger.debug(f"Scraping with requests: {url}")
            
            response = self.session.get(url, timeout=30)
            response.raise_for_status()
            
            if response.text:
                self.logger.debug(f"Successfully retrieved {len(response.text)} characters")
                return response.text
            else:
                self.logger.warning("Empty response from requests")
                return None
                
        except Exception as e:
            self.logger.error(f"Requests scraping failed: {e}")
            return None
    
    def _scrape_with_firecrawl(self, url: str) -> Optional[str]:
        """Scrape content using Firecrawl as fallback."""
        try:
            self.logger.debug(f"Scraping with Firecrawl: {url}")
            
            params = {
                'formats': ['html'],
                'waitFor': 3000,  # Wait for JavaScript
                'onlyMainContent': False  # Need full page for LTA sites
            }
            
            result = self.firecrawl_client.scrape_url(url, params=params)
            
            if result and 'html' in result:
                content = result['html']
                self.logger.debug(f"Firecrawl retrieved {len(content)} characters")
                return content
            else:
                self.logger.warning("No HTML content from Firecrawl")
                return None
                
        except Exception as e:
            self.logger.error(f"Firecrawl scraping failed: {e}")
            return None
    
    def _parse_lta_clubspark_content(self, html_content: str, venue_config: Dict[str, Any]) -> List[Dict[str, Any]]:
        """Parse LTA Clubspark HTML content to extract court availability."""
        court_availability = []
        
        try:
            soup = BeautifulSoup(html_content, 'html.parser')
            
            # Find all resource sessions (both available and booked)
            sessions = soup.find_all('div', class_='resource-session')
            
            self.logger.debug(f"Found {len(sessions)} resource sessions")
            
            # Look for court headers in the booking area
            court_headers = soup.find_all(['h3', 'h4', 'div'], class_=lambda x: x and any(term in str(x).lower() for term in ['court', 'resource']))
            court_names = []
            
            # Try different selectors for court names
            selectors = ['h3', 'h4', '.court-header', '.resource-header']
            for selector in selectors:
                headers = soup.select(selector)
                for header in headers:
                    text = header.get_text(strip=True)
                    if 'Court' in text and text not in court_names:
                        court_names.append(text)
            
            # If no court names found, extract from data structure or generate
            if not court_names:
                # Try to count unique courts from session data
                unique_slots = set()
                for session in sessions:
                    slot_key = session.get('data-slot-key', '')
                    if slot_key:
                        # Extract court identifier from slot key
                        parts = slot_key.split('-')
                        if len(parts) > 0:
                            unique_slots.add(parts[0])
                
                # Generate court names based on unique slots
                for i, slot_id in enumerate(sorted(unique_slots)):
                    court_names.append(f"Court {i+1}")
            
            if not court_names:
                court_names = ["Court 1", "Court 2"]  # Default fallback
            
            self.logger.debug(f"Found court names: {court_names}")
            
            # Build court mapping from sessions
            court_session_map = {}
            sessions_per_court = len(sessions) // len(court_names) if court_names else 1
            
            # Map sessions to courts (simplified approach)
            current_court_index = 0
            for i, session in enumerate(sessions):
                try:
                    # Determine which court this session belongs to
                    court_index = i // sessions_per_court if sessions_per_court > 0 else 0
                    if court_index >= len(court_names):
                        court_index = len(court_names) - 1
                    current_court_name = court_names[court_index]
                    
                    # Extract session data attributes
                    availability = session.get('data-availability', 'false') == 'true'
                    start_time = session.get('data-start-time')
                    end_time = session.get('data-end-time')
                    session_cost = session.get('data-session-cost', '0')
                    
                    # Convert time to readable format
                    time_slot = self._convert_time_minutes(start_time, end_time)
                    
                    if availability:
                        # Available slot - look for booking link
                        book_link = session.find('a', class_='book-interval')
                        if book_link:
                            cost_span = book_link.find('span', class_='cost')
                            slot_span = book_link.find('span', class_='available-booking-slot')
                            
                            price = '0'
                            if cost_span:
                                price_text = cost_span.get_text(strip=True)
                                price_match = re.search(r'£([\d.]+)', price_text)
                                if price_match:
                                    price = price_match.group(1)
                            
                            court_availability.append({
                                'court_name': current_court_name,
                                'time_slot': time_slot,
                                'status': 'available',
                                'available': True,
                                'price': price,
                                'session_info': None
                            })
                    else:
                        # Booked or unavailable slot
                        session_div = session.find('div', class_='full-session')
                        if session_div:
                            title = session_div.get('title', '')
                            status = 'booked'
                            session_info = None
                            
                            # Extract session info from title or content
                            if 'Booked' in title:
                                status = 'booked'
                            elif any(keyword in title.lower() for keyword in ['coaching', 'lesson', 'session']):
                                status = 'reserved'
                                session_info = title.replace('Category:', '').strip()
                            else:
                                # Look for text content
                                text_content = session_div.get_text(strip=True)
                                if any(keyword in text_content.lower() for keyword in ['coaching', 'lesson', 'beginner', 'advanced', 'intermediate']):
                                    status = 'reserved'
                                    session_info = text_content
                            
                            court_availability.append({
                                'court_name': current_court_name,
                                'time_slot': time_slot,
                                'status': status,
                                'available': False,
                                'price': '0',
                                'session_info': session_info
                            })
                    
                except Exception as e:
                    self.logger.warning(f"Error parsing session {i}: {e}")
                    continue
            
            self.logger.info(f"Parsed {len(court_availability)} court slots")
            return court_availability
            
        except Exception as e:
            self.logger.error(f"Error parsing LTA Clubspark content: {e}")
            return []
    
    def _convert_time_minutes(self, start_minutes: str, end_minutes: str) -> str:
        """Convert time in minutes since midnight to readable format."""
        try:
            if not start_minutes or not end_minutes:
                return "Unknown"
            
            start_min = int(start_minutes)
            end_min = int(end_minutes)
            
            start_hour = start_min // 60
            start_minute = start_min % 60
            end_hour = end_min // 60
            end_minute = end_min % 60
            
            start_time = f"{start_hour:02d}:{start_minute:02d}"
            end_time = f"{end_hour:02d}:{end_minute:02d}"
            
            return f"{start_time}-{end_time}"
            
        except (ValueError, TypeError):
            return "Unknown"


def main():
    """Test the LTA Clubspark scraper."""
    logging.basicConfig(level=logging.INFO)
    
    # Test venue configuration
    test_venue = {
        '_id': 'stratford_park_test',
        'name': 'Stratford Park (Test)',
        'provider': 'lta_clubspark',
        'url': 'https://stratford.newhamparkstennis.org.uk/Booking/BookByDate#?date=2025-06-09&role=guest',
        'courts': [
            {'id': 'court_1', 'name': 'Court 1'},
            {'id': 'court_2', 'name': 'Court 2'}
        ]
    }
    
    scraper = LTAClubsparkScraper()
    result = scraper.scrape_venue_availability(test_venue)
    
    print(f"Status: {result['status']}")
    if result['status'] == 'success':
        print(f"Courts found: {len(result['court_availability'])}")
        print(f"Time slots: {result['time_slots']}")
        for slot in result['court_availability'][:5]:  # Show first 5
            print(f"  - {slot['court_name']} at {slot['time_slot']}: {slot['status']}")


if __name__ == "__main__":
    main() 