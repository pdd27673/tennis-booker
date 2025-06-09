# 🎾 Tennis Court Booking System - Cost-Effective Edition

A **production-ready tennis court notification system** that monitors court availability across multiple venues and sends real-time alerts to users. Designed for **cost-effectiveness** with strategic use of premium services and lightweight daily operations.

## 🌟 Key Features

### Cost-Effective Scraping Strategy
- **💡 Daily Lightweight Scraping**: Uses direct HTTP requests (no premium API costs)
- **🔥 Strategic Firecrawl Usage**: Only for initial site analysis (minimal credits)
- **⚡ Smart Fallback**: Firecrawl only when lightweight scraping fails
- **📅 7-Day Ahead Monitoring**: Tennis courts release slots weekly

### Complete Notification System
- **📧 Email Notifications**: Instant alerts for court availability
- **🔄 Redis Pub/Sub**: Real-time message processing
- **👤 User Preferences**: Customizable venue, time, and price filters
- **📊 Alert History**: Track notification history and performance
- **🏥 Health Monitoring**: System health checks and metrics

### Production Features
- **🚀 Single Command Deployment**: Start entire system with one script
- **📋 OpenAPI Specification**: Auto-generated API documentation
- **🧪 Comprehensive Testing**: Full system integration tests
- **📈 Continuous Monitoring**: Real-time system health monitoring
- **🧹 Automated Cleanup**: Log rotation and maintenance

## 🏃‍♂️ Quick Start

### 1. Initial Setup
```bash
# Clone and setup
git clone [repository]
cd tennis-booker

# One-command setup (installs everything)
./tennis-system setup
```

### 2. Start the System
```bash
# Start all services (Docker + API + Scheduler)
./tennis-system start

# Check system status
./tennis-system status
```

### 3. Run Cost-Effective Scraping
```bash
# Daily scraping (no premium costs)
./tennis-system scrape-daily

# Full 7-day scraping
./tennis-system scrape-full

# One-time site analysis (uses Firecrawl credits)
./tennis-system scrape-analyze --force
```

## 📖 Master Control Script

The `./tennis-system` command provides unified control:

```bash
🎾 Tennis Court Booking System - Master Control Script

COMMANDS:
    start           Start all services (Docker + API + Scheduler)
    stop            Stop all services
    restart         Restart all services
    status          Check status of all services
    
    scrape-daily    Run daily lightweight scraping (no Firecrawl credits)
    scrape-full     Run full 7-day scraping  
    scrape-analyze  Use Firecrawl to analyze booking sites (costs credits)
    
    test-system     Run comprehensive system tests
    setup           Initialize project and seed data
    monitor         Monitor system health continuously
    openapi         Generate OpenAPI specification

COST-EFFECTIVE STRATEGY:
    - Use 'scrape-daily' for everyday monitoring (no Firecrawl costs)
    - Use 'scrape-analyze' ONCE to understand sites (minimal credits)
    - Use 'scrape-full' for comprehensive 7-day scanning
    - Firecrawl fallback only when lightweight scraping fails
```

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                 COST-EFFECTIVE SCRAPING FLOW               │
└─────────────────────────────────────────────────────────────┘

Real Tennis Sites → Lightweight HTTP → Redis Pub/Sub → Email Alerts
       ↓                    ↓               ↓              ↑
   Firecrawl         Pattern Analysis    Notification    User
   (Fallback)        (One-time)         Service      Preferences
                                                         ↓
    MongoDB ← API Endpoints ← Processed Data ← Alert History
```

### Core Components

1. **Unified Scraper** (`src/tennis_court_scraper.py`)
   - Daily lightweight scraping (no costs)
   - Strategic Firecrawl usage for analysis
   - 7-day ahead monitoring
   - Redis integration

2. **Master Control** (`scripts/run_tennis_system.sh`)
   - Single-command operations
   - Service management
   - Health monitoring
   - Testing and maintenance

3. **API System** (Go - `cmd/` and `internal/`)
   - RESTful endpoints
   - Health checks and metrics
   - Venue and preference management
   - Court availability queries

4. **Notification Engine** (Go + Redis - `utils/redis_worker.py`)
   - Real-time message processing
   - Email notifications
   - User preference filtering
   - Alert history tracking

5. **Utilities** (`utils/` directory)
   - Database seeding (`seed_venues.py`)
   - Data verification tools
   - Redis notification worker
   - Python dependencies

## 📁 Project Structure

```
tennis-booker/
├── 🎾 Core System
│   ├── src/
│   │   └── tennis_court_scraper.py # Unified scraper (main)
│   ├── scripts/
│   │   ├── run_tennis_system.sh    # Master control script
│   │   └── seed_venues.sh          # Venue seeding wrapper
│   ├── tennis-system               # Convenience wrapper
│   └── docker-compose.yml          # Docker services
│
├── 🏗️ Backend (Go)
│   ├── cmd/                        # Go applications
│   ├── internal/                   # Go internal packages
│   ├── bin/                        # Compiled binaries
│   └── Makefile                    # Build configuration
│
├── 🛠️ Utilities
│   └── utils/
│       ├── seed_venues.py          # Database seeding
│       ├── redis_worker.py         # Notification worker
│       ├── verify_*.py             # Verification tools
│       └── requirements.txt        # Python dependencies
│
├── 🧪 Testing
│   ├── tests/                      # Test files
│   └── test_*.py                   # Integration tests
│
├── 📊 Data & Logs
│   ├── logs/                       # System logs
│   ├── data/                       # Data storage
│   └── scraper-env/                # Python virtual environment
│
└── 📋 Documentation
    ├── README.md                   # This file
    ├── prd.md                      # Product requirements
    └── api-spec.yaml               # OpenAPI specification
```

## 🔧 Configuration

### Environment Variables
```bash
# API Keys (only for providers you want to use)
FIRECRAWL_API_KEY=fc-...        # For fallback scraping
SENDGRID_API_KEY=SG....         # For email notifications

# System Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password
MONGODB_URI=mongodb://localhost:27017
API_BASE=http://localhost:8080
```

### Supported Tennis Venues
- **Victoria Park** (4 courts) - Courtsides system
- **Ropemakers Field** (2 courts) - Courtsides system  
- **Stratford Park Tennis** - LTA ClubSpark system

## 📊 API Endpoints

### Health & Monitoring
- `GET /api/health` - System health check
- `GET /api/metrics` - System performance metrics

### Venues & Courts
- `GET /api/v1/venues` - List all tennis venues
- `GET /api/v1/courts/available` - Get available court slots
  - Filters: `venue_ids`, `date_from`, `date_to`, `price_min`, `price_max`

### User Management
- `GET /api/v1/preferences?user_id=...` - Get user preferences
- `PUT /api/v1/preferences?user_id=...` - Update preferences
- `GET /api/v1/alerts/history?user_id=...` - Get alert history

## 🧪 Testing

### Quick Tests
```bash
# Test entire system
./tennis-system test-system

# Test scraping only
./tennis-system test-scraping

# Test API endpoints
./tennis-system test-api
```

### Manual Testing
```bash
# Test individual scraper
python src/tennis_court_scraper.py --mode=test --log-level=DEBUG

# Test specific venues
python src/tennis_court_scraper.py --mode=daily --venues=victoria,stratford
```

## 💰 Cost Optimization

### Firecrawl Usage Strategy
1. **One-time Analysis**: Run `scrape-analyze` once per venue to understand structure
2. **Daily Operations**: Use `scrape-daily` for ongoing monitoring (no API costs)
3. **Emergency Fallback**: Firecrawl only when lightweight scraping fails
4. **Estimated Costs**: ~$3-5/month with normal usage vs $50-100/month with daily Firecrawl

### Resource Optimization
- **Lightweight Scraping**: Direct HTTP requests with smart parsing
- **Rate Limiting**: 2-second delays between requests
- **Efficient Caching**: Store analysis results locally
- **Minimal Dependencies**: Only essential libraries

## 📈 Monitoring & Maintenance

### Continuous Monitoring
```bash
# Real-time system monitor
./tennis-system monitor

# View recent logs
./tennis-system logs

# System status check
./tennis-system status
```

### Automated Cleanup
```bash
# Clean old logs and cache
./tennis-system cleanup
```

### Production Deployment

1. **Setup Production Environment**
   ```bash
   # Production setup
   ./tennis-system setup
   ```

2. **Configure Monitoring**
   - Set up log monitoring
   - Configure alert thresholds
   - Schedule daily scraping

3. **Schedule Daily Operations**
   ```bash
   # Add to crontab for daily scraping
   0 9 * * * cd /path/to/tennis-booker && ./tennis-system scrape-daily
   ```

## 🔄 Typical Daily Workflow

```bash
# Morning: Check system health
./tennis-system status

# Run daily scraping (no costs)
./tennis-system scrape-daily

# Monitor results
./tennis-system monitor

# Evening: Cleanup if needed
./tennis-system cleanup
```

## 🛠️ Development

### Project Structure
```
tennis-booker/
├── tennis-system                 # Convenience wrapper script
├── src/
│   └── tennis_court_scraper.py   # Unified cost-effective scraper
├── scripts/
│   ├── run_tennis_system.sh      # Master control script
│   └── seed_venues.sh            # Venue data seeding
├── cmd/
│   ├── api/                       # API server
│   ├── scheduler/                 # Background scheduler
│   └── notification-service/      # Email notification service
├── internal/
│   ├── handlers/                  # API route handlers
│   ├── models/                    # Data models
│   └── services/                  # Business logic
├── utils/                         # Python utilities
├── tests/                         # Test files
├── docker-compose.yml            # Infrastructure services
└── api-spec.yaml                 # OpenAPI specification
```

### Adding New Venues
1. Update venue configuration in `utils/seed_venues.py`
2. Add scraping patterns for the booking system
3. Test with `./tennis-system test-scraping`
4. Run analysis if needed: `./tennis-system scrape-analyze`

## 📝 OpenAPI Documentation

Generate comprehensive API documentation:
```bash
./tennis-system openapi
# Creates api-spec.yaml with full endpoint documentation
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Test changes: `./tennis-system test-system`
4. Commit changes: `git commit -m 'Add amazing feature'`
5. Push to branch: `git push origin feature/amazing-feature`
6. Create Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🎯 Next Steps

- [ ] Add more tennis venues
- [ ] Implement mobile app integration
- [ ] Add SMS notifications
- [ ] Create web dashboard
- [ ] Add court booking automation
- [ ] Implement ML-based availability prediction

---

**🎾 Ready to never miss a tennis court booking again? Get started with one command: `./tennis-system setup`** 