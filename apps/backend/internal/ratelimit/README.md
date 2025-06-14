# Rate Limiting Package

This package provides comprehensive rate limiting functionality for the Tennis Booker API using Redis as the distributed backend store.

## Features

- **Multi-layered Rate Limiting**: IP-based, user-based, and endpoint-specific limits
- **Redis Backend**: Distributed rate limiting across multiple server instances
- **Comprehensive Middleware**: Easy integration with HTTP handlers
- **Rate Limit Headers**: Standard HTTP headers for client guidance
- **Monitoring & Logging**: Detailed logging of rate limit events
- **Load Testing**: Built-in tools for testing rate limiting behavior

## Quick Start

### 1. Initialize Rate Limiter

```go
import "tennis-booker/internal/ratelimit"

// Create configuration
config := ratelimit.DefaultConfig()
config.RedisHost = "localhost"
config.RedisPort = 6379

// Initialize limiter
limiter, err := ratelimit.NewLimiter(config)
if err != nil {
    log.Fatal(err)
}
defer limiter.Close()
```

### 2. Apply Middleware

```go
import "net/http"

mux := http.NewServeMux()

// IP-based rate limiting for public endpoints
mux.Handle("/api/health", 
    ratelimit.IPRateLimitMiddleware(limiter)(healthHandler))

// User-based rate limiting for protected endpoints
mux.Handle("/api/users/me", 
    ratelimit.UserRateLimitMiddleware(limiter)(
        auth.JWTMiddleware(jwtService)(userHandler)))

// Strict rate limiting for auth endpoints
mux.Handle("/auth/login", 
    ratelimit.AuthRateLimitMiddleware(limiter)(loginHandler))
```

## Middleware Types

### IPRateLimitMiddleware
- **Purpose**: General IP-based rate limiting
- **Default**: 100 requests/minute per IP
- **Use Case**: Public endpoints, general traffic control

### AuthRateLimitMiddleware
- **Purpose**: Strict rate limiting for authentication endpoints
- **Default**: 10 requests/minute per IP
- **Use Case**: Login, register, password reset endpoints

### UserRateLimitMiddleware
- **Purpose**: User-based rate limiting for authenticated requests
- **Default**: 500 requests/minute per user
- **Use Case**: Protected endpoints requiring authentication

### DataRateLimitMiddleware
- **Purpose**: Moderate rate limiting for data-intensive endpoints
- **Default**: 200 requests/minute per IP
- **Use Case**: API endpoints that return large datasets

### SensitiveRateLimitMiddleware
- **Purpose**: Very strict rate limiting for sensitive operations
- **Default**: 5 requests/minute per IP
- **Use Case**: Administrative endpoints, system operations

### CombinedRateLimitMiddleware
- **Purpose**: Defense-in-depth with both IP and user limits
- **Behavior**: Both limits must be respected
- **Use Case**: High-security endpoints

### CustomRateLimitMiddleware
- **Purpose**: Flexible rate limiting with custom configuration
- **Configuration**: Specify custom limits and windows
- **Use Case**: Endpoints with specific requirements

## Configuration

### Default Configuration

```go
config := ratelimit.DefaultConfig()
// Redis settings
config.RedisHost = "localhost"
config.RedisPort = 6379
config.RedisPassword = ""
config.RedisDB = 0

// Rate limits
config.DefaultIPLimit = ratelimit.RateLimit{
    Requests: 100,
    Window:   time.Minute,
}
config.DefaultUserLimit = ratelimit.RateLimit{
    Requests: 500,
    Window:   time.Minute,
}
config.AuthEndpointLimit = ratelimit.RateLimit{
    Requests: 10,
    Window:   time.Minute,
}

// Features
config.IncludeHeaders = true
config.TrustedProxies = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
```

### Custom Rate Limits

```go
// Custom rate limit: 5 requests per hour
customLimit := ratelimit.RateLimit{
    Requests: 5,
    Window:   time.Hour,
}

middleware := ratelimit.CustomRateLimitMiddleware(limiter, customLimit, "expensive-op")
mux.Handle("/api/expensive-operation", middleware(handler))
```

## Rate Limit Headers

The middleware automatically adds standard rate limit headers:

- `X-RateLimit-Limit`: Maximum requests allowed in the window
- `X-RateLimit-Remaining`: Requests remaining in current window
- `X-RateLimit-Reset`: Unix timestamp when window resets
- `X-RateLimit-Reset-Time`: Human-readable reset time (RFC3339)
- `Retry-After`: Seconds until client can retry (when rate limited)

## Monitoring & Logging

### Log Events

The system logs rate limiting events with different levels:

```
[RATE_LIMIT_DEBUG] - Successful requests (debug level)
[RATE_LIMIT_BLOCKED] - Rate limited requests (warning level)
[RATE_LIMIT_ERROR] - System errors (error level)
```

### Example Log Output

```
[RATE_LIMIT_BLOCKED] IP=192.168.1.100 UserID=user123 Endpoint=/api/venues Method=GET Type=user Requests=501/500 Window=1m UserAgent=curl/7.68.0
```

### Structured Logging

For structured logging systems, use the `RateLimitEvent` struct:

```go
type RateLimitEvent struct {
    Timestamp    time.Time `json:"timestamp"`
    IP           string    `json:"ip"`
    UserID       string    `json:"user_id,omitempty"`
    Endpoint     string    `json:"endpoint"`
    Method       string    `json:"method"`
    LimitType    string    `json:"limit_type"`
    RequestsMade int64     `json:"requests_made"`
    Limit        int64     `json:"limit"`
    Window       string    `json:"window"`
    Blocked      bool      `json:"blocked"`
    UserAgent    string    `json:"user_agent,omitempty"`
}
```

## Testing

### Unit Tests

```bash
# Run all rate limiting tests
go test ./internal/ratelimit -v

# Run specific test suites
go test ./internal/ratelimit -run "TestLimiter" -v
go test ./internal/ratelimit -run "TestMiddleware" -v
go test ./internal/ratelimit -run "TestIntegration" -v
```

### Load Testing

Use the built-in load testing tool:

```bash
# Basic load test
go run scripts/load-test/main.go -endpoint=/api/health -requests=100 -concurrent=10

# Rate limit test
go run scripts/load-test/main.go -endpoint=/auth/login -test-rate-limit=true

# Burst traffic test
go run scripts/load-test/main.go -endpoint=/api/health -requests=50 -concurrent=50

# Duration-based test
go run scripts/load-test/main.go -endpoint=/api/health -duration=60s -concurrent=5
```

### Automated Testing Script

```bash
# Run comprehensive rate limiting tests
./scripts/test-rate-limiting.sh
```

## Performance

### Benchmarks

- **Middleware Overhead**: ~1-2ms per request
- **Redis Operations**: ~0.5ms average latency
- **Memory Usage**: Minimal (stateless middleware)
- **Throughput**: Supports thousands of requests per second

### Optimization Tips

1. **Redis Connection Pooling**: Enabled by default
2. **Trusted Proxy Configuration**: Reduces IP extraction overhead
3. **Header Inclusion**: Disable if not needed to reduce response size
4. **Window Size**: Larger windows reduce Redis operations

## Security Considerations

### IP Spoofing Protection

- Validates trusted proxy headers
- Supports CIDR range validation
- Falls back to direct connection IP

### Rate Limit Bypass Prevention

- Redis-based distributed counting
- Multiple validation layers
- Secure header processing

### DDoS Protection

- IP-based rate limiting prevents single-source attacks
- Combined rate limiting provides defense in depth
- Configurable limits allow fine-tuning

## Troubleshooting

### Common Issues

1. **Redis Connection Errors**
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

Enable debug logging:

```bash
export TASKMASTER_LOG_LEVEL=debug
```

### Health Checks

```go
// Check Redis connectivity
err := limiter.HealthCheck(context.Background())
if err != nil {
    log.Printf("Rate limiter health check failed: %v", err)
}
```

## Examples

### Basic HTTP Server with Rate Limiting

```go
package main

import (
    "log"
    "net/http"
    "tennis-booker/internal/ratelimit"
)

func main() {
    // Initialize rate limiter
    config := ratelimit.DefaultConfig()
    limiter, err := ratelimit.NewLimiter(config)
    if err != nil {
        log.Fatal(err)
    }
    defer limiter.Close()

    // Create router with rate limiting
    mux := http.NewServeMux()
    
    // Public endpoint with IP rate limiting
    mux.Handle("/health", 
        ratelimit.IPRateLimitMiddleware(limiter)(
            http.HandlerFunc(healthHandler)))
    
    // Auth endpoint with strict rate limiting
    mux.Handle("/login", 
        ratelimit.AuthRateLimitMiddleware(limiter)(
            http.HandlerFunc(loginHandler)))

    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
    // Login logic here
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Login successful"))
}
```

### Custom Rate Limiting

```go
// Password reset: 2 requests per hour
passwordResetLimit := ratelimit.RateLimit{
    Requests: 2,
    Window:   time.Hour,
}

mux.Handle("/auth/password-reset", 
    ratelimit.CustomRateLimitMiddleware(limiter, passwordResetLimit, "password-reset")(
        passwordResetHandler))

// Data export: 10 requests per hour
dataExportLimit := ratelimit.RateLimit{
    Requests: 10,
    Window:   time.Hour,
}

mux.Handle("/api/export", 
    ratelimit.CustomRateLimitMiddleware(limiter, dataExportLimit, "data-export")(
        dataExportHandler))
```

## Contributing

1. Add tests for new functionality
2. Update documentation
3. Follow existing code patterns
4. Ensure backward compatibility

## License

This package is part of the Tennis Booker project. 