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