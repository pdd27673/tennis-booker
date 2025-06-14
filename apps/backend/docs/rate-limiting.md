# API Rate Limiting Documentation

## Overview

The Tennis Booker API implements comprehensive rate limiting to protect against abuse, ensure fair usage, and maintain system stability. Rate limiting is applied at multiple levels using Redis as the distributed backend store.

## Rate Limiting Strategy

### Multi-Layered Approach

1. **IP-Based Rate Limiting**: Applied to all requests based on client IP address
2. **User-Based Rate Limiting**: Applied to authenticated requests based on user ID
3. **Endpoint-Specific Rate Limiting**: Custom limits for sensitive or resource-intensive endpoints
4. **Combined Rate Limiting**: Defense-in-depth approach applying both IP and user limits

### Rate Limiting Hierarchy

```
Request → IP Rate Limit → Authentication → User Rate Limit → Endpoint Handler
```

## Default Rate Limits

### IP-Based Limits (Per IP Address)

| Endpoint Category | Requests | Window | Description |
|------------------|----------|---------|-------------|
| General Traffic | 100 | 1 minute | Default limit for all IP addresses |
| Authentication | 10 | 1 minute | Login, register, refresh, logout |
| Public Health | 20 | 1 minute | Health checks and status endpoints |

### User-Based Limits (Per Authenticated User)

| Endpoint Category | Requests | Window | Description |
|------------------|----------|---------|-------------|
| General Authenticated | 500 | 1 minute | Default limit for authenticated users |
| Data Access | 200 | 1 minute | Venues, courts, booking data |
| Sensitive Operations | 5 | 1 minute | System management, admin functions |

### Custom Endpoint Limits

| Endpoint | Requests | Window | Rationale |
|----------|----------|---------|-----------|
| Password Reset | 2 | 1 hour | Prevent abuse of password reset |
| Data Export | 10 | 1 hour | Resource-intensive operations |
| Bulk Operations | 5 | 1 minute | High-impact database operations |

## Protected Endpoints

### Public Endpoints (IP Rate Limited)

- `GET /api/health` - Health check endpoint
- `POST /auth/register` - User registration
- `POST /auth/login` - User authentication
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout

### Protected Endpoints (Combined IP + User Rate Limited)

- `GET /api/users/me` - User profile
- `PUT /api/users/preferences` - Update user preferences
- `GET /api/venues` - List venues
- `GET /api/courts` - List courts

### Sensitive Endpoints (Strict Rate Limited)

- `GET /api/system/status` - System status
- `POST /api/system/pause` - Pause system operations
- `POST /api/system/resume` - Resume system operations

## HTTP Headers

### Rate Limit Information Headers

The API includes standard rate limit headers in all responses:

| Header | Description | Example |
|--------|-------------|---------|
| `X-RateLimit-Limit` | Maximum requests allowed in the window | `100` |
| `X-RateLimit-Remaining` | Requests remaining in current window | `95` |
| `X-RateLimit-Reset` | Unix timestamp when window resets | `1640995200` |
| `X-RateLimit-Reset-Time` | Human-readable reset time (RFC3339) | `2021-12-31T12:00:00Z` |

### Rate Limit Exceeded Response

When rate limits are exceeded, the API returns:

- **Status Code**: `429 Too Many Requests`
- **Retry-After Header**: Seconds until the client can retry
- **Error Message**: Descriptive message indicating the limit type

Example response:
```http
HTTP/1.1 429 Too Many Requests
Retry-After: 45
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1640995245
X-RateLimit-Reset-Time: 2021-12-31T12:00:45Z
Content-Type: application/json

{
  "error": "Too many requests from IP 192.168.1.100. Try again in 45 seconds."
}
```

## Client Implementation Guidelines

### Handling Rate Limits

1. **Check Headers**: Always check rate limit headers in responses
2. **Respect Retry-After**: Wait for the specified time before retrying
3. **Implement Backoff**: Use exponential backoff for repeated failures
4. **Cache Responses**: Cache API responses to reduce request frequency

### Example Client Code (JavaScript)

```javascript
async function makeAPIRequest(url, options = {}) {
  try {
    const response = await fetch(url, options);
    
    // Check rate limit headers
    const limit = response.headers.get('X-RateLimit-Limit');
    const remaining = response.headers.get('X-RateLimit-Remaining');
    const resetTime = response.headers.get('X-RateLimit-Reset-Time');
    
    console.log(`Rate limit: ${remaining}/${limit}, resets at ${resetTime}`);
    
    if (response.status === 429) {
      const retryAfter = response.headers.get('Retry-After');
      console.log(`Rate limited. Retry after ${retryAfter} seconds`);
      
      // Wait and retry
      await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
      return makeAPIRequest(url, options);
    }
    
    return response;
  } catch (error) {
    console.error('API request failed:', error);
    throw error;
  }
}
```

## Configuration

### Environment Variables

Rate limiting can be configured via environment variables:

```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limit Configuration (handled via .taskmaster/config.json)
# Use task-master models command to configure
```

### Custom Rate Limits

Developers can apply custom rate limits to specific endpoints:

```go
// Custom rate limit: 5 requests per hour
customLimit := ratelimit.CustomRateLimitMiddleware(rateLimiter, ratelimit.RateLimit{
    Requests: 5,
    Window:   time.Hour,
})

mux.Handle("/api/expensive-operation", customLimit(handler))
```

## Monitoring and Logging

### Rate Limit Events

The system logs rate limit events for monitoring and analysis:

```json
{
  "timestamp": "2021-12-31T12:00:00Z",
  "level": "WARN",
  "message": "Rate limit exceeded",
  "ip": "192.168.1.100",
  "user_id": "user123",
  "endpoint": "/api/venues",
  "limit_type": "user",
  "requests_made": 501,
  "limit": 500,
  "window": "1m"
}
```

### Metrics

Key metrics to monitor:

- Rate limit hit rate by endpoint
- Top rate-limited IP addresses
- User rate limit patterns
- Redis performance and availability

## Troubleshooting

### Common Issues

1. **Redis Connection Failures**
   - Check Redis server availability
   - Verify connection parameters
   - Monitor Redis memory usage

2. **Unexpected Rate Limiting**
   - Check for shared IP addresses (NAT, proxies)
   - Verify user authentication status
   - Review custom rate limit configurations

3. **Performance Issues**
   - Monitor Redis latency
   - Check network connectivity
   - Review rate limit window sizes

### Debug Mode

Enable debug logging to troubleshoot rate limiting issues:

```bash
export TASKMASTER_LOG_LEVEL=debug
```

## Security Considerations

### IP Spoofing Protection

The system validates trusted proxy headers:

- Configurable trusted proxy CIDR ranges
- Header validation for X-Forwarded-For
- Fallback to direct connection IP

### Rate Limit Bypass Prevention

- Redis-based distributed counting prevents bypass
- Multiple validation layers (IP + User)
- Secure header processing

## Performance

### Redis Backend

- Distributed rate limiting across multiple instances
- Efficient Redis operations with minimal latency
- Automatic cleanup of expired rate limit keys

### Middleware Performance

- Minimal overhead per request (~1-2ms)
- Efficient IP extraction and validation
- Optimized Redis operations

## Testing

### Load Testing

Use the provided load testing scripts:

```bash
# Run basic load test
go run scripts/load-test/main.go -endpoint=/api/health -requests=1000 -concurrent=10

# Run rate limit test
go run scripts/load-test/main.go -endpoint=/auth/login -requests=100 -concurrent=5 -test-rate-limit
```

### Integration Tests

Run the comprehensive test suite:

```bash
go test ./internal/ratelimit -v
```

## Migration and Deployment

### Zero-Downtime Deployment

Rate limiting supports zero-downtime deployment:

1. Deploy new instances with rate limiting enabled
2. Gradually shift traffic to new instances
3. Monitor rate limit effectiveness
4. Complete migration when stable

### Rollback Procedure

If issues arise:

1. Disable rate limiting middleware
2. Monitor system performance
3. Investigate and fix issues
4. Re-enable with adjusted limits

## Support

For questions or issues with rate limiting:

1. Check the logs for rate limit events
2. Review this documentation
3. Contact the development team
4. File an issue in the project repository

---

*Last updated: December 2024*
*Version: 1.0.0* 