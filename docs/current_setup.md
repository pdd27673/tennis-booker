# Tennis Booker - Current Setup & Architecture

## 🏗️ Project Overview

**Tennis Booker** is a full-stack tennis court booking automation system with real-time availability tracking, user authentication, and automated court scraping. The project follows a microservices architecture with clear separation of concerns across frontend, backend, and scraping services.

## 📦 Service Architecture

### Core Services

| Service | Technology | Purpose | Status |
|---------|------------|---------|---------|
| **Frontend** | React 19 + TypeScript + Vite | User interface and dashboard | ✅ Production Ready |
| **Backend** | Go 1.23 + Gorilla Mux | REST API, authentication, business logic | ✅ Production Ready |
| **Scraper** | Python 3.11+ + Playwright | Court availability monitoring | ✅ Production Ready |
| **Notification Service** | Go (separate binary) | Email notifications for bookings | ✅ Production Ready |

### Infrastructure Services

| Service | Technology | Purpose | Production Status |
|---------|------------|---------|-------------------|
| **MongoDB** | MongoDB 6.0 | Primary database | ✅ Critical |
| **Redis** | Redis 7.0 Alpine | Caching, pub/sub, rate limiting | ✅ Critical |
| **HashiCorp Vault** | Vault 1.13/1.16 | Secret management | ⚠️ **OPTIONAL** |
| **Traefik** | Traefik v3.0 | Reverse proxy, SSL termination | ✅ Production only |

## 🔍 Detailed Component Analysis

### Frontend (`apps/frontend/`)
- **Framework**: React 19 with TypeScript
- **Build Tool**: Vite 6.x
- **UI Framework**: Tailwind CSS + Radix UI + Aceternity UI
- **State Management**: Zustand + TanStack Query
- **Authentication**: JWT tokens with automatic refresh
- **Deployment**: Vercel-ready with PWA support
- **Testing**: Vitest + Testing Library

**Key Dependencies**:
- `@tanstack/react-query` - Server state management
- `react-router-dom` - Client-side routing  
- `axios` - HTTP client
- `zod` - Schema validation
- `framer-motion` - Animations

### Backend (`apps/backend/`)
- **Language**: Go 1.23
- **Framework**: Gorilla Mux for HTTP routing
- **Architecture**: Clean architecture with internal packages
- **Authentication**: JWT-based with bcrypt password hashing
- **Rate Limiting**: Multi-layer protection with Redis
- **Database**: MongoDB with connection pooling
- **Secret Management**: HashiCorp Vault integration (optional)

**Key Dependencies**:
- `github.com/gorilla/mux` - HTTP router
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `go.mongodb.org/mongo-driver` - MongoDB driver
- `github.com/redis/go-redis/v9` - Redis client
- `github.com/hashicorp/vault/api` - Vault integration
- `github.com/ulule/limiter/v3` - Rate limiting

**Service Binaries**:
- `server` - Main API server
- `notification-service` - Email notification daemon
- `retention-service` - Data cleanup service
- `seed-db` - Database seeding utility
- `seed-user` - User seeding utility

### Scraper (`apps/scraper/`)
- **Language**: Python 3.11+
- **Framework**: Playwright for browser automation
- **Architecture**: Modular scraper with orchestrator pattern
- **Scheduling**: Built-in scheduler with configurable intervals
- **Deduplication**: Redis-based duplicate detection
- **Court Platforms**: ClubSpark, Courtside (extensible)

**Key Dependencies**:
- `playwright>=1.48.0` - Browser automation
- `pymongo>=4.6.2` - MongoDB client
- `redis>=5.0.1` - Redis client
- `beautifulsoup4>=4.12.3` - HTML parsing

## 🚀 Deployment Configurations

### Development Setup (`docker-compose.yml`)
**Services Running**:
- MongoDB (port 27017)
- Redis (port 6379) 
- Vault DEV mode (port 8200) - **OPTIONAL**
- Scraper service with 5-minute intervals

**Not Included in Dev**:
- Frontend (runs via `npm run dev` on port 5173)
- Backend (runs via Go binary on port 8080)
- Traefik (not needed for local development)

### Production Setup (`docker-compose.prod.yml`)
**Additional Services**:
- Traefik reverse proxy with SSL termination
- Frontend served via Nginx
- All services containerized
- Health checks for all critical services
- Proper volume persistence
- Production Vault configuration (file storage)

**Production URLs**:
- Frontend: `https://yourdomain.com`
- API: `https://api.yourdomain.com`
- Traefik Dashboard: `https://traefik.yourdomain.com`

## 🔐 Vault Usage Analysis

### ⚠️ **IMPORTANT: Vault is OPTIONAL**

**Current Vault Integration**:
- Used for storing database credentials
- Platform-specific credentials for scrapers
- JWT secrets and API tokens
- Email service credentials

**Vault Usage Status**:
- ✅ **Fully Integrated**: Backend has complete Vault client
- ✅ **Fallback Available**: Environment variables work without Vault
- ✅ **Development Ready**: DEV mode with root token
- ⚠️ **Production Optional**: Can use environment variables instead

**Files Using Vault**:
- `apps/backend/internal/auth/vault_client.go` - Vault client implementation
- `apps/backend/internal/config/vault.go` - Vault configuration
- `apps/backend/internal/secrets/manager.go` - Secret management
- `apps/backend/internal/database/connection.go` - DB connection with Vault

**Vault Alternatives for Production**:
- **Railway**: Use Railway's built-in environment variables
- **Docker Secrets**: Use Docker swarm secrets
- **Cloud Provider Secrets**: AWS Secrets Manager, GCP Secret Manager
- **Environment Variables**: Traditional env vars with secure storage

## 🏭 Production Considerations

### Critical Components for Railway Deployment

**✅ REQUIRED Services**:
1. **MongoDB** - Primary database (Railway has MongoDB addon)
2. **Redis** - Caching and pub/sub (Railway has Redis addon)
3. **Backend API** - Core business logic
4. **Frontend** - User interface
5. **Scraper Service** - Court monitoring

**⚠️ OPTIONAL Services**:
1. **Vault** - Can be replaced with Railway environment variables
2. **Notification Service** - Can be integrated into main backend
3. **Traefik** - Railway handles reverse proxy and SSL

**🔄 DEVELOPMENT-ONLY Services**:
1. **Seed services** - Run once during setup
2. **DB tools** - Utility services for maintenance

### Railway Migration Strategy

**Phase 1 - Core Services**:
- Backend API service
- Frontend service  
- MongoDB addon
- Redis addon

**Phase 2 - Background Services**:
- Scraper service (as separate Railway service)
- Notification service (or integrate into backend)

**Phase 3 - Optimization**:
- CDN for frontend assets
- Database optimization
- Monitoring and logging

### Configuration Management

**Current Environment Variables**:
```bash
# Database
MONGO_URI=mongodb://...
MONGO_ROOT_USERNAME=admin
MONGO_ROOT_PASSWORD=...
DB_NAME=tennis_booking

# Cache
REDIS_ADDR=redis:6379
REDIS_PASSWORD=...

# Authentication
JWT_SECRET=...

# Email (for notifications)
GMAIL_EMAIL=...
GMAIL_PASSWORD=...

# Scraper
SCRAPER_INTERVAL=300
```

**Railway Environment Variables Strategy**:
- Use Railway's secure environment variable storage
- No Vault needed - Railway provides secure secret management
- Use Railway's database connection strings
- Leverage Railway's internal DNS for service-to-service communication

## 📊 Resource Requirements

### Minimum Production Resources
- **Backend**: 512MB RAM, 0.5 CPU
- **Frontend**: Static hosting (CDN)
- **Scraper**: 1GB RAM, 0.5 CPU (for Playwright)
- **MongoDB**: 1GB RAM, 1GB storage minimum
- **Redis**: 256MB RAM

### Scaling Considerations
- **Backend**: Stateless, can scale horizontally
- **Scraper**: Single instance recommended (avoid duplicate scraping)
- **Database**: MongoDB replica set for production
- **Cache**: Redis cluster for high availability

## 🔧 Development Workflow

### Local Development
```bash
# One-command start
./scripts/run_local.sh start

# Or manual setup
make setup          # Install dependencies
make dev           # Start infrastructure
make run-local     # Start all services
```

### Testing
```bash
make test          # Run all tests
make backend-test  # Go tests only
make scraper-test  # Python tests only
```

### Production Build
```bash
make build         # Build all services
docker-compose -f docker-compose.prod.yml up -d
```

## 🚨 Migration Recommendations

### For Railway Deployment:

1. **Remove Vault Dependency**: Replace with Railway environment variables
2. **Consolidate Services**: Consider merging notification service into backend
3. **Use Railway Addons**: MongoDB and Redis addons for managed infrastructure
4. **Environment Configuration**: Set up Railway-specific environment variables
5. **Health Checks**: Ensure all services have proper health endpoints
6. **Monitoring**: Add logging and monitoring for production visibility

### Critical Files for Another LLM:
- `docker-compose.yml` - Development setup
- `docker-compose.prod.yml` - Production configuration  
- `apps/backend/go.mod` - Backend dependencies
- `apps/frontend/package.json` - Frontend dependencies
- `apps/scraper/pyproject.toml` - Scraper dependencies
- `Makefile` - Build and deployment commands
- Configuration files in `config/` directory

This project is **production-ready** with optional Vault integration. The architecture is well-suited for cloud deployment platforms like Railway with minimal modifications needed.