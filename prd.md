# Tennis Court Booking Bot - Product Requirements Document

## 1. Executive Summary

**Product Vision**: An automated tennis court booking system that monitors London tennis venues 24/7 and secures preferred time slots the moment they become available.

**Problem Statement**: Tennis courts in London are extremely competitive to book, with slots opening exactly 7 days in advance and filling within minutes. Manual booking is unreliable and time-consuming.

**Solution**: A multi-component system combining intelligent scraping, automated booking, and smart scheduling to secure courts for preferred times (Sat/Sun 10am-3pm, Fri 7pm) at £10/hour or less.

---

## 2. Product Scope & Objectives

### Core Objectives
- **Primary**: Automatically book tennis courts at preferred times within budget
- **Secondary**: Alert user when new slots become available matching preferences
- **Tertiary**: Learn booking patterns to optimize success rates

### Success Metrics
- Booking success rate > 80% for preferred time slots
- Alert latency < 2 minutes from slot availability
- System uptime > 99.5%
- User confirmation response time < 24 hours

---

## 3. Functional Requirements

### 3.1 Core Features

#### F1: Automated Court Monitoring
- **Description**: Continuously scrape target venues for new slot availability
- **Acceptance Criteria**:
  - Monitor LTA/Clubspark and courtsides.com/tennistowerhamlets every 5 minutes
  - Detect new slots within 2 minutes of availability
  - Store slot data with timestamps for pattern analysis
  - Handle rate limiting and anti-bot measures

#### F2: Intelligent Booking Engine
- **Description**: Automatically book courts when slots match user preferences
- **Acceptance Criteria**:
  - Book slots exactly 7 days in advance at slot opening time
  - Handle payment processing during booking
  - Implement retry logic with exponential backoff
  - Prioritize time over venue (preferred time at any venue > alternative time at preferred venue)

#### F3: User Preference Management
- **Description**: Allow dynamic preference updates and booking confirmations
- **Acceptance Criteria**:
  - API endpoint for real-time preference updates
  - 24-hour advance confirmation system for weekly bookings
  - Default preferences: Sat/Sun 10am-3pm, Fri 7pm, max £10/hour
  - Support for venue-specific preferences and blacklists

#### F4: Notification System
- **Description**: Multi-channel alerts for booking events and slot availability
- **Acceptance Criteria**:
  - Email notifications for bookings, confirmations, failures
  - SMS alerts for urgent confirmations and successful bookings
  - Real-time alerts when matching slots become available
  - Booking summary reports with success/failure analytics

### 3.2 System Features

#### F5: Credential Management
- **Description**: Secure storage and rotation of booking platform credentials
- **Acceptance Criteria**:
  - Encrypted credential storage with HashiCorp Vault
  - Automatic credential validation and rotation
  - Multi-platform credential support
  - Audit logging for credential usage

#### F6: Data Analytics & Learning
- **Description**: Learn booking patterns to optimize success rates
- **Acceptance Criteria**:
  - Track booking success rates by time, venue, day of week
  - Identify optimal booking times and strategies
  - Generate weekly performance reports
  - Suggest preference optimizations based on historical data

---

## 4. Technical Architecture

### 4.1 Tech Stack

#### Backend Services
- **Scheduler**: Go (Golang) with cron-like scheduling
- **Scraper**: Python with Firecrawl MCP integration
- **Database**: MongoDB for document storage
- **Message Queue**: Redis for task queuing
- **API Server**: Go with Gin framework

#### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Credential Store**: HashiCorp Vault
- **Monitoring**: Prometheus + Grafana
- **Logging**: ELK Stack (Elasticsearch, Logstash, Kibana)

#### External Integrations
- **Notifications**: Twilio (SMS), SendGrid (Email)
- **Browser Automation**: Playwright (Python)
- **AI/LLM**: Firecrawl MCP for intelligent scraping
- **TaskManager AI**: Roo Code integration for orchestration

### 4.2 System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Go Scheduler  │────│  Redis Queue    │────│ Python Scraper  │
│   (Main Engine) │    │  (Task Queue)   │    │ (Firecrawl MCP) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MongoDB       │    │  HashiCorp      │    │  Notification   │
│   (Data Store)  │    │  Vault          │    │  Services       │
│                 │    │  (Credentials)  │    │  (SMS/Email)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 4.3 Scraping Strategy

**Option 1: Python + Firecrawl MCP (Recommended)**
- Leverage Firecrawl's AI-powered content extraction
- Handle JavaScript-heavy booking platforms
- Built-in anti-detection measures
- LLM integration for content understanding

**Option 2: Go + Colly/Goquery**
- Native Go scraping with excellent performance
- Custom anti-detection implementation required
- More complex JavaScript handling

**Option 3: LLM-Powered Scraping**
- Use TaskManager AI with browser automation MCP
- Natural language understanding of booking interfaces
- Adaptive to UI changes
- Higher cost per scrape

**Recommended**: Python + Firecrawl MCP for reliability and AI integration

---

## 5. Project Setup Commands

### 5.1 Initial Project Structure
```bash
# Create project directory
mkdir tennis-booking-bot
cd tennis-booking-bot

# Initialize Go module
go mod init tennis-booking-bot

# Create directory structure
mkdir -p {cmd,internal,pkg,api,configs,scripts,deployments,docs}
mkdir -p {internal/{scheduler,booking,scraper,auth,notifications},pkg/{database,utils}}
mkdir -p {configs/{development,production},scripts/{setup,deploy}}

# Create main application files
touch cmd/scheduler/main.go
touch cmd/api/main.go
touch internal/scheduler/scheduler.go
touch internal/booking/engine.go
touch internal/scraper/client.go
touch pkg/database/mongodb.go
touch configs/config.yaml
touch docker-compose.yml
touch Dockerfile
touch .env.example
touch README.md
```

### 5.2 Go Dependencies
```bash
# Core dependencies
go get github.com/gin-gonic/gin
go get github.com/robfig/cron/v3
go get go.mongodb.org/mongo-driver/mongo
go get github.com/go-redis/redis/v8
go get github.com/spf13/viper
go get github.com/hashicorp/vault/api

# Additional utilities
go get github.com/sirupsen/logrus
go get github.com/stretchr/testify
go get github.com/golang-jwt/jwt/v4
```

### 5.3 Python Environment Setup
```bash
# Create Python virtual environment for scraper
python -m venv scraper-env
source scraper-env/bin/activate  # On Windows: scraper-env\Scripts\activate

# Install Python dependencies
pip install firecrawl-py playwright beautifulsoup4 requests redis pymongo python-dotenv
pip install playwright
playwright install
```

### 5.4 TaskManager AI Integration
```bash
# Install TaskManager AI (assuming npm/node setup)
npm install -g taskmanager-ai

# Initialize TaskManager configuration
taskmanager init --project tennis-booking-bot
taskmanager config set --llm claude-sonnet-4 --mode orchestrator

# Create TaskManager workflow
mkdir .taskmanager
touch .taskmanager/workflows.yml
touch .taskmanager/prompts/booking-orchestrator.md
```

### 5.5 Infrastructure Setup
```bash
# Start development infrastructure
docker-compose up -d mongodb redis vault

# Initialize Vault
vault operator init
vault operator unseal
vault auth -method=userpass username=admin

# Create initial secrets
vault write secret/tennis-bot/lta username="your-username" password="your-password"
vault write secret/tennis-bot/courtsides username="your-username" password="your-password"
```

---

## 6. Data Models

### 6.1 MongoDB Collections

#### Users Collection
```json
{
  "_id": "ObjectId",
  "email": "user@example.com",
  "phone": "+44XXXXXXXXX",
  "preferences": {
    "times": ["Saturday 10:00-15:00", "Sunday 10:00-15:00", "Friday 19:00"],
    "max_price": 10,
    "preferred_venues": ["venue_id_1", "venue_id_2"],
    "excluded_venues": []
  },
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

#### Venues Collection
```json
{
  "_id": "ObjectId",
  "name": "Tennis Venue Name",
  "platform": "lta_clubspark", // or "courtsides"
  "url": "https://venue-booking-url.com",
  "location": {
    "address": "123 Tennis St, London",
    "coordinates": [51.5074, -0.1278]
  },
  "courts": [
    {
      "id": "court_1",
      "name": "Court 1",
      "surface": "hard",
      "hourly_rate": 8.50
    }
  ],
  "booking_window_days": 7,
  "booking_opens_at": "00:00", // Time when new slots open
  "active": true
}
```

#### Bookings Collection
```json
{
  "_id": "ObjectId",
  "user_id": "ObjectId",
  "venue_id": "ObjectId",
  "court_id": "court_1",
  "date": "2025-01-15",
  "start_time": "10:00",
  "end_time": "11:00",
  "price": 8.50,
  "status": "confirmed", // pending, confirmed, cancelled, failed
  "booking_reference": "BK123456",
  "booked_at": "2025-01-08T00:01:00Z",
  "confirmation_required_by": "2025-01-14T00:00:00Z",
  "confirmed_at": "2025-01-13T14:30:00Z"
}
```

#### Scraping Logs Collection
```json
{
  "_id": "ObjectId",
  "venue_id": "ObjectId",
  "scrape_timestamp": "2025-01-01T12:00:00Z",
  "slots_found": [
    {
      "date": "2025-01-08",
      "time": "10:00-11:00",
      "court": "Court 1",
      "price": 8.50,
      "available": true
    }
  ],
  "scrape_duration_ms": 2500,
  "errors": [],
  "success": true
}
```

---

## 7. API Endpoints

### 7.1 Core Endpoints

#### Booking Management
- `GET /api/v1/bookings` - List user bookings
- `POST /api/v1/bookings` - Create manual booking
- `PUT /api/v1/bookings/{id}/confirm` - Confirm pending booking
- `DELETE /api/v1/bookings/{id}` - Cancel booking

#### Preferences
- `GET /api/v1/preferences` - Get user preferences
- `PUT /api/v1/preferences` - Update preferences
- `POST /api/v1/preferences/venues` - Add preferred venue

#### Monitoring
- `GET /api/v1/venues` - List monitored venues
- `GET /api/v1/venues/{id}/slots` - Get available slots
- `POST /api/v1/venues/{id}/monitor` - Start monitoring venue

#### System
- `GET /api/health` - Health check
- `GET /api/metrics` - System metrics
- `GET /api/v1/logs` - System logs

---

## 8. Development Phases

### Phase 1: Foundation (Week 1-2)
- Set up project structure and dependencies
- Implement basic Go scheduler with MongoDB
- Create Python scraper with Firecrawl integration
- Basic notification system

### Phase 2: Core Functionality (Week 3-4)
- Implement booking engine with retry logic
- Add credential management with Vault
- Build preference management API
- Integrate TaskManager AI orchestration

### Phase 3: Intelligence & Optimization (Week 5-6)
- Add booking pattern analytics
- Implement smart scheduling optimization
- Build comprehensive monitoring and alerting
- Performance tuning and reliability improvements

### Phase 4: Production Readiness (Week 7-8)
- Comprehensive testing and error handling
- Security audit and penetration testing
- Production deployment and monitoring setup
- Documentation and user training

---

## 9. Risk Assessment

### Technical Risks
- **Anti-bot measures**: Venues may implement CAPTCHAs or IP blocking
- **Platform changes**: Booking sites may change UI/API without notice
- **Rate limiting**: Too frequent scraping may result in temporary bans
- **Payment failures**: Booking may fail at payment step

### Mitigation Strategies
- Use residential proxies and user-agent rotation
- Implement adaptive scraping with Firecrawl's AI capabilities
- Implement exponential backoff and circuit breakers
- Test payment flows regularly and implement fallback notifications

### Business Risks
- **Terms of service violations**: Automated booking may violate platform ToS
- **Competition**: Other bots may compete for same slots
- **Booking policy changes**: Venues may change booking windows or policies

---

## 10. Success Criteria

### Minimum Viable Product (MVP)
- ✅ Successfully book at least 3 courts per week for preferred times
- ✅ Alert system working with <5 minute latency
- ✅ 95% system uptime during peak booking hours
- ✅ User confirmation system functional

### Full Product Success
- ✅ 80%+ booking success rate for preferred slots
- ✅ Adaptive learning improving success rates over time
- ✅ Support for 5+ venues across London
- ✅ Sub-2-minute response time for new slot alerts
- ✅ Comprehensive analytics and reporting dashboard