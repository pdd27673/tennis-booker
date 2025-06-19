# Court Data API Endpoints

This document provides comprehensive documentation for the Court Data API endpoints in the Tennis Booker application.

## Overview

The Court Data API provides access to tennis venue information and available court booking slots. All endpoints require JWT authentication and return JSON responses.

## Authentication

All endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

### Error Responses

- `401 Unauthorized`: Missing, invalid, or expired JWT token
- `400 Bad Request`: Invalid request parameters
- `405 Method Not Allowed`: Unsupported HTTP method
- `500 Internal Server Error`: Server-side error

## Endpoints

### 1. Get Venues

Retrieves a list of all active tennis venues.

**Endpoint:** `GET /api/venues`

**Authentication:** Required

**Parameters:** None

**Response:**

```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "name": "Central Park Tennis Center",
    "location": "New York, NY",
    "courts": 6,
    "available_slots": 15,
    "earliest_available": "2024-01-15T09:00:00Z",
    "booking_url": "https://centralparktennis.com/book",
    "scraping_log_id": "SCRAPING_LOG_ID_PLACEHOLDER"
  }
]
```

**Example Request:**

```bash
curl -X GET "https://api.tennisbooker.com/api/venues" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

### 2. Get Court Slots

Retrieves available court booking slots with optional filtering capabilities.

**Endpoint:** `GET /api/courts`

**Authentication:** Required

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `venueId` | string | Filter by specific venue ID (ObjectID format) | `507f1f77bcf86cd799439011` |
| `date` | string | Filter by specific date (YYYY-MM-DD format) | `2024-01-15` |
| `startTime` | string | Filter slots starting at/after time (HH:MM format) | `18:00` |
| `endTime` | string | Filter slots ending at/before time (HH:MM format) | `20:00` |
| `provider` | string | Filter by provider type | `lta`, `courtsides` |
| `minPrice` | float | Minimum price filter | `15.00` |
| `maxPrice` | float | Maximum price filter | `30.00` |
| `limit` | integer | Limit number of results (default: 100) | `50` |

**Response:**

```json
[
  {
    "id": "premium_court1_2024-01-15_18:00",
    "venue_id": "507f1f77bcf86cd799439011",
    "venue_name": "Premium Tennis Club",
    "court_id": "court_1",
    "court_name": "Court 1",
    "date": "2024-01-15",
    "start_time": "18:00",
    "end_time": "19:00",
    "price": 25.00,
    "currency": "GBP",
    "available": true,
    "booking_url": "https://premium-tennis.com/book/court1",
    "provider": "lta",
    "last_scraped": "2024-01-15T17:30:00Z",
    "scraping_log_id": "507f1f77bcf86cd799439012"
  }
]
```

**Example Requests:**

1. **Get all available court slots:**
```bash
curl -X GET "https://api.tennisbooker.com/api/courts" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

2. **Filter by venue:**
```bash
curl -X GET "https://api.tennisbooker.com/api/courts?venueId=507f1f77bcf86cd799439011" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

3. **Filter by date and time:**
```bash
curl -X GET "https://api.tennisbooker.com/api/courts?date=2024-01-15&startTime=18:00" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

4. **Filter by price range:**
```bash
curl -X GET "https://api.tennisbooker.com/api/courts?minPrice=20&maxPrice=30" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

5. **Combined filters with limit:**
```bash
curl -X GET "https://api.tennisbooker.com/api/courts?venueId=507f1f77bcf86cd799439011&date=2024-01-15&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

## Filtering Logic

### Time Range Filtering

The time filtering logic works as follows:

- **startTime**: Returns slots where the slot's start time is >= the specified time
- **endTime**: Returns slots where the slot's end time is <= the specified time
- **Combined**: When both are specified, returns slots that overlap with the specified time range

**Examples:**

- `startTime=18:00`: Returns slots starting at 18:00 or later
- `endTime=20:00`: Returns slots ending at 20:00 or earlier  
- `startTime=18:00&endTime=20:00`: Returns slots that overlap with 18:00-20:00 window

### Price Range Filtering

- **minPrice**: Returns slots with price >= specified value
- **maxPrice**: Returns slots with price <= specified value
- **Combined**: Returns slots within the specified price range (inclusive)

### Provider Filtering

Supported provider values:
- `lta`: LTA (Lawn Tennis Association) venues
- `courtsides`: Courtsides booking platform venues

### Venue Filtering

- Must be a valid MongoDB ObjectID format (24-character hexadecimal string)
- Returns only slots for the specified venue

### Date Filtering

- Must be in YYYY-MM-DD format
- Returns only slots for the specified date

## Error Handling

### Validation Errors (400 Bad Request)

```json
{
  "error": "Invalid venue ID format"
}
```

Common validation errors:
- Invalid venue ID format (not a valid ObjectID)
- Invalid date format (not YYYY-MM-DD)
- Invalid time format (not HH:MM)
- Invalid price values (not valid numbers)
- Invalid limit value (not a positive integer)

### Authentication Errors (401 Unauthorized)

```json
{
  "error": "Authorization header required"
}
```

```json
{
  "error": "Invalid or expired token"
}
```

### Server Errors (500 Internal Server Error)

```json
{
  "error": "Internal server error"
}
```

## Performance Considerations

### Optimization Features

1. **MongoDB-level filtering**: Most filters are applied at the database level for optimal performance
2. **Application-level filtering**: Complex time range logic is handled in the application layer
3. **Result limiting**: Default limit of 100 results to prevent large response payloads
4. **Indexed queries**: Database queries utilize appropriate indexes for fast retrieval

### Response Times

- **Venues endpoint**: Typically < 50ms
- **Courts endpoint (no filters)**: Typically < 100ms
- **Courts endpoint (with filters)**: Typically < 150ms

### Rate Limiting

While not currently implemented, consider implementing rate limiting for production use:
- Recommended: 100 requests per minute per user
- Burst allowance: 20 requests per 10 seconds

## Data Freshness

- Court slot data is updated through periodic scraping
- Scraping intervals vary by venue (typically 30-60 minutes)
- `last_scraped` timestamp indicates data freshness
- Venues may have different `scraping_interval` configurations

## Integration Examples

### JavaScript/Node.js

```javascript
const axios = require('axios');

const apiClient = axios.create({
  baseURL: 'https://api.tennisbooker.com',
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
    'Content-Type': 'application/json'
  }
});

// Get all venues
const venues = await apiClient.get('/api/venues');

// Get court slots for today with price filter
const today = new Date().toISOString().split('T')[0];
const courts = await apiClient.get('/api/courts', {
  params: {
    date: today,
    minPrice: 20,
    maxPrice: 40,
    limit: 20
  }
});
```

### Python

```python
import requests
from datetime import date

class TennisBookerAPI:
    def __init__(self, jwt_token):
        self.base_url = 'https://api.tennisbooker.com'
        self.headers = {
            'Authorization': f'Bearer {jwt_token}',
            'Content-Type': 'application/json'
        }
    
    def get_venues(self):
        response = requests.get(f'{self.base_url}/api/venues', headers=self.headers)
        return response.json()
    
    def get_courts(self, **filters):
        response = requests.get(f'{self.base_url}/api/courts', 
                              headers=self.headers, params=filters)
        return response.json()

# Usage
api = TennisBookerAPI('your_jwt_token')
venues = api.get_venues()
courts = api.get_courts(date=str(date.today()), provider='lta', limit=10)
```

## Testing

The API includes comprehensive test coverage:

- **Unit tests**: Individual handler function testing
- **Integration tests**: End-to-end request/response testing
- **Performance tests**: Response time validation
- **Concurrent request tests**: Multi-user scenario testing

Test coverage includes:
- Authentication scenarios (valid, invalid, expired tokens)
- All filtering combinations
- Error handling and validation
- Edge cases and boundary conditions
- Performance benchmarks

## Changelog

### Version 1.0.0 (Current)
- Initial implementation of venues and courts endpoints
- JWT authentication integration
- Comprehensive filtering capabilities
- Full test coverage
- Performance optimization

## Support

For API support or questions:
- Documentation: This file
- Test examples: See `internal/handlers/court_integration_test.go`
- Implementation: See `internal/handlers/court.go` 