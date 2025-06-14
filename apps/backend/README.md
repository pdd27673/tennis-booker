# Tennis Booker Backend

A comprehensive Go backend service for the Tennis Booker application, featuring authentication, rate limiting, database management, and notification services.

## 🏗️ Architecture Overview

The backend is organized into several key packages:

- **`cmd/`** - Entry points and executable services
- **`internal/`** - Private application packages
- **`docs/`** - Documentation and API specifications
- **`scripts/`** - Utility scripts and tools

## 📁 Project Structure

```
apps/backend/
├── cmd/                          # Entry points and services
│   ├── config-test/             # Configuration testing utility
│   ├── db-tools/                # Database management tools
│   ├── notification-service/    # Notification service daemon
│   ├── seed-db/                 # Database seeding utility
│   ├── seed-user/               # User seeding utility
│   └── test-auth-server/        # Test authentication server
├── internal/                     # Private application packages
│   ├── auth/                    # Authentication and JWT services
│   ├── config/                  # Configuration management
│   ├── database/                # Database connections and repositories
│   ├── handlers/                # HTTP handlers and routing
│   ├── models/                  # Data models and business logic
│   ├── ratelimit/               # Rate limiting middleware
│   ├── redis/                   # Redis event publishing
│   └── secrets/                 # Secret management (Vault integration)
├── docs/                        # Documentation
│   ├── api/                     # API documentation
│   └── rate-limiting.md         # Rate limiting guide
├── scripts/                     # Utility scripts
│   ├── load-test/               # Load testing tools
│   └── test-rate-limiting.sh    # Rate limiting test script
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── Makefile                     # Build and development commands
├── test.sh                      # Test runner script
└── SECURITY_NOTES.md           # Security audit results
```

## 🚀 Features

### Authentication & Security
- **JWT Authentication** - Secure token-based authentication
- **Password Hashing** - bcrypt-based password security
- **Vault Integration** - HashiCorp Vault for secret management
- **Rate Limiting** - Multi-layered API rate limiting
- **Security Audit** - Comprehensive security scanning and hardening

### Database Support
- **MongoDB** - Primary database with connection pooling
- **Redis** - Caching and rate limiting backend
- **Repository Pattern** - Clean data access layer
- **Database Tools** - Seeding, indexing, and management utilities

### Rate Limiting
- **Multi-layered Protection** - IP, user, and endpoint-specific limits
- **Redis Backend** - Distributed rate limiting
- **Configurable Limits** - Environment-based configuration
- **Monitoring** - Comprehensive logging and metrics

### Services
- **Notification Service** - Event-driven notifications
- **Event Publishing** - Redis-based event system
- **Configuration Management** - Environment-aware configuration
- **Health Checks** - Service health monitoring

## 🛠️ Development Setup

### Prerequisites
- Go 1.23+
- MongoDB
- Redis
- HashiCorp Vault (optional, for production)

### Installation

1. **Clone and navigate to backend:**
   ```bash
   cd apps/backend
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start required services:**
   ```bash
   # MongoDB (example with Docker)
   docker run -d -p 27017:27017 --name mongodb mongo:latest
   
   # Redis (example with Docker)
   docker run -d -p 6379:6379 --name redis redis:latest
   ```

### Running Services

#### Development Server
```bash
# Run test authentication server
make run-notification
# or
go run cmd/test-auth-server/main.go
```

#### Database Management
```bash
# Seed database with test data
make run-seed-db
# or
go run cmd/seed-db/main.go

# Create database indexes
go run cmd/db-tools/ensure_indexes.go
```

#### Testing
```bash
# Run all tests
make test
# or
go test ./...

# Run tests without MongoDB integration
./test.sh

# Test rate limiting
./scripts/test-rate-limiting.sh
```

## 📚 Package Documentation

### Authentication (`internal/auth/`)
- **JWT Service** - Token generation and validation
- **Password Service** - Secure password hashing
- **Vault Client** - HashiCorp Vault integration
- **Middleware** - HTTP authentication middleware

### Configuration (`internal/config/`)
- **Environment-aware** - Development, staging, production configs
- **Vault Integration** - Secure secret retrieval
- **Feature Flags** - Runtime feature toggling
- **Validation** - Configuration validation and defaults

### Database (`internal/database/`)
- **Connection Management** - MongoDB and Redis connections
- **Repository Pattern** - Clean data access abstractions
- **Models** - User, Venue, Booking, and Notification models
- **Migrations** - Database schema management

### Rate Limiting (`internal/ratelimit/`)
- **Multi-layered Strategy** - IP, user, and endpoint-specific limits
- **Redis Backend** - Distributed rate limiting
- **Middleware Suite** - 7 different middleware types
- **Monitoring** - Comprehensive logging and metrics

See [Rate Limiting Documentation](docs/rate-limiting.md) for detailed usage.

### Handlers (`internal/handlers/`)
- **REST API** - HTTP handlers for all endpoints
- **Authentication** - Login, register, token refresh
- **User Management** - Profile and preferences
- **System Management** - Health checks and admin operations
- **Court Management** - Venue and court operations

## 🔧 Configuration

### Environment Variables

#### Database Configuration
```bash
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=tennis_booking
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

#### Authentication
```bash
JWT_SECRET=your-jwt-secret
JWT_ISSUER=tennis-booker
JWT_EXPIRY=24h
```

#### Vault (Production)
```bash
VAULT_ADDR=https://vault.example.com
VAULT_TOKEN=your-vault-token
VAULT_ROLE_ID=your-role-id
VAULT_SECRET_ID=your-secret-id
```

#### Rate Limiting
```bash
RATE_LIMIT_IP_REQUESTS=100
RATE_LIMIT_IP_WINDOW=60
RATE_LIMIT_USER_REQUESTS=500
RATE_LIMIT_USER_WINDOW=60
```

### Configuration Files
- **Development** - Uses environment variables and defaults
- **Production** - Integrates with HashiCorp Vault for secrets
- **Testing** - Isolated test configurations

## 🧪 Testing

### Test Categories
- **Unit Tests** - Individual package testing
- **Integration Tests** - Cross-package functionality
- **Load Tests** - Performance and rate limiting
- **Security Tests** - Authentication and authorization

### Running Tests
```bash
# All tests
go test ./...

# Specific package
go test ./internal/auth -v

# With coverage
go test ./... -cover

# Load testing
go run scripts/load-test/main.go -endpoint=/api/health -requests=100
```

## 🚀 Deployment

### Build
```bash
# Build all services
make build

# Build specific service
go build -o bin/notification-service ./cmd/notification-service
```

### Docker
```bash
# Build Docker image
docker build -t tennis-booker-backend .

# Run with Docker Compose
docker-compose up -d
```

### Production Considerations
- **Vault Integration** - Use Vault for all secrets in production
- **Rate Limiting** - Configure appropriate limits for your traffic
- **Monitoring** - Set up logging and metrics collection
- **Health Checks** - Configure load balancer health checks
- **Database Indexes** - Ensure all required indexes are created

## 📊 Monitoring & Observability

### Logging
- **Structured Logging** - JSON format for production
- **Rate Limit Events** - Detailed rate limiting logs
- **Error Tracking** - Comprehensive error logging
- **Performance Metrics** - Request timing and throughput

### Health Checks
- **Service Health** - `/api/health` endpoint
- **Database Health** - MongoDB and Redis connectivity
- **Vault Health** - Secret service availability

### Metrics
- **Request Metrics** - Request count, latency, errors
- **Rate Limit Metrics** - Limit hits, remaining capacity
- **Database Metrics** - Connection pool, query performance

## 🔒 Security

### Security Features
- **Authentication** - JWT-based secure authentication
- **Rate Limiting** - Multi-layered DDoS protection
- **Secret Management** - Vault integration for production secrets
- **Input Validation** - Request validation and sanitization
- **CORS** - Cross-origin request security

### Security Audit
The codebase has undergone comprehensive security scanning:
- ✅ No hardcoded secrets
- ✅ Proper secret management
- ✅ Secure authentication flow
- ✅ Rate limiting protection

See [SECURITY_NOTES.md](SECURITY_NOTES.md) for detailed audit results.

## 🤝 Contributing

### Development Workflow
1. **Format Code** - `gofmt -w .`
2. **Run Tests** - `go test ./...`
3. **Check Security** - Review for hardcoded secrets
4. **Update Documentation** - Keep docs current

### Code Standards
- **Go Conventions** - Follow standard Go practices
- **Error Handling** - Comprehensive error handling
- **Testing** - Maintain test coverage
- **Documentation** - Document public APIs

## 📝 API Documentation

API documentation is available in the `docs/api/` directory. Key endpoints:

### Authentication
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Token refresh
- `POST /auth/logout` - User logout

### User Management
- `GET /api/users/me` - Get current user
- `PUT /api/users/preferences` - Update preferences

### System
- `GET /api/health` - Health check
- `GET /api/system/status` - System status

### Courts & Venues
- `GET /api/venues` - List venues
- `GET /api/courts` - List courts

## 📞 Support

For questions or issues:
1. Check the documentation in `docs/`
2. Review test files for usage examples
3. Check the issue tracker
4. Consult the security notes for security-related questions

## 📄 License

This project is part of the Tennis Booker application.

## Database Optimization

### MongoDB Index Strategy

The application uses optimized MongoDB indexes for high-performance queries:

#### Court Slots Collection
```javascript
// Primary query optimization (venue + date + time)
db.court_slots.createIndex({venue_id: 1, slot_date: 1, start_time: 1}, {name: "idx_venue_date_time"})

// Availability queries
db.court_slots.createIndex({is_available: 1, last_scraped: 1}, {name: "idx_availability_freshness"})

// Court-specific queries
db.court_slots.createIndex({venue_id: 1, court_id: 1, slot_date: 1}, {name: "idx_venue_court_date"})

// Automatic cleanup (7 days)
db.court_slots.createIndex({slot_date: 1}, {name: "idx_slot_date_ttl", expireAfterSeconds: 604800})
```

#### User Preferences Collection
```javascript
// Unique user lookup
db.user_preferences.createIndex({user_id: 1}, {name: "idx_user_id_unique", unique: true})

// Notification settings
db.user_preferences.createIndex({user_id: 1, "notification_settings.enabled": 1}, {name: "idx_user_notifications"})

// Sparse indexes for optional fields
db.user_preferences.createIndex({preferred_venues: 1}, {name: "idx_preferred_venues", sparse: true})
db.user_preferences.createIndex({preferred_sports: 1}, {name: "idx_preferred_sports", sparse: true})
```

#### Scraping Logs Collection
```javascript
// Monitoring queries
db.scraping_logs.createIndex({venue_id: 1, scrape_timestamp: 1}, {name: "idx_venue_timestamp"})

// Status-based queries
db.scraping_logs.createIndex({status: 1, scrape_timestamp: 1}, {name: "idx_status_timestamp"})

// Automatic cleanup (30 days)
db.scraping_logs.createIndex({scrape_timestamp: 1}, {name: "idx_scrape_timestamp_ttl", expireAfterSeconds: 2592000})
```

### Index Management Tools

Use the provided MongoDB optimization tools in `scripts/mongodb/`:

```bash
# Analyze current database performance
cd scripts/mongodb
go run analyze_indexes.go > analysis.json

# Apply optimizations
go run optimize_indexes.go analysis.json

# Or use the automated script
./optimize_mongodb.sh --analyze --optimize
```

### Query Performance Verification

Verify index effectiveness using MongoDB's explain:

```javascript
// Check if queries use indexes (should show IXSCAN, not COLLSCAN)
db.court_slots.find({venue_id: "venue123", slot_date: {$gte: new Date()}}).explain("executionStats")

// Performance targets:
// - Execution time: <100ms
// - Stage: IXSCAN (not COLLSCAN)
// - Document efficiency: totalDocsExamined/nReturned ≤ 1.5
```

## Redis Integration

### Deduplication Strategy

The scraper uses Redis for efficient slot deduplication:

```go
// Key format: dedupe:slot:<venueId>:<date>:<startTime>:<courtId>
key := fmt.Sprintf("dedupe:slot:%s:%s:%s:%s", venueId, date, startTime, courtId)

// Check for duplicates with 48-hour expiry
result := redisClient.Set(ctx, key, slotData, 48*time.Hour).Val()
if result == "OK" {
    // New slot - process and store
    processSlot(slot)
} else {
    // Duplicate - skip processing
    skipSlot(slot)
}
```

### Performance Targets

- **Redis deduplication hit rate**: >90%
- **MongoDB query execution time**: <100ms
- **System throughput**: >100 slots/second
- **Error rate**: <1%

## Configuration

### Environment Variables

```bash
# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=tennis_booking
REDIS_URL=redis://localhost:6379

# Authentication
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# API Configuration
PORT=8080
GIN_MODE=release

# Scraper Integration
SCRAPER_REDIS_DB=1
DEDUP_EXPIRY_HOURS=48
```

### Production Deployment

```bash
# Build Docker image
docker build -t tennis-booking-backend .

# Run with Docker Compose
docker-compose up -d

# Health check
curl http://localhost:8080/health
```

## Monitoring

### Key Metrics

Monitor these performance indicators:

- **API Response Times**: <200ms for 95th percentile
- **Database Query Performance**: <100ms average
- **Redis Hit Rate**: >90% for deduplication
- **Error Rate**: <1% of requests
- **Memory Usage**: Stable, no leaks

### Database Monitoring

```javascript
// Check index usage
db.court_slots.aggregate([{$indexStats: {}}])

// Monitor slow queries (>100ms)
db.setProfilingLevel(2, {slowms: 100})
db.system.profile.find().sort({ts: -1}).limit(10)
```

## Development

### Code Structure

```
cmd/
├── api/           # API server
├── retention/     # Data retention service
└── notification/  # Notification service

internal/
├── handlers/      # HTTP handlers
├── models/        # Data models
├── services/      # Business logic
├── middleware/    # HTTP middleware
└── database/      # Database layer

scripts/
├── mongodb/       # Database optimization tools
└── load-test/     # Performance testing
```

### Testing

```bash
# Unit tests
go test ./...

# Integration tests
make test-integration

# Load testing
cd scripts/load-test
./run-load-test.sh
```

### Database Migrations

```bash
# Apply indexes
cd scripts/mongodb
./optimize_mongodb.sh --optimize

# Verify performance
./optimize_mongodb.sh --verify
```

## Troubleshooting

### Common Issues

**Slow Queries**
```javascript
// Check execution plan
db.collection.find(query).explain("executionStats")

// Look for COLLSCAN instead of IXSCAN
// Add appropriate indexes if needed
```

**Redis Connection Issues**
```bash
# Test Redis connectivity
redis-cli ping

# Check Redis memory usage
redis-cli info memory
```

**High Memory Usage**
```bash
# Check MongoDB memory usage
db.serverStatus().mem

# Monitor Go memory usage
curl http://localhost:8080/debug/pprof/heap
```

## Contributing

1. Follow Go conventions and best practices
2. Add tests for new features
3. Update documentation
4. Verify performance impact
5. Use conventional commits

## License

MIT License - see LICENSE file for details. 