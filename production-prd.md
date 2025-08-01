# Tennis Booker Production Deployment Strategy & Product Requirements

## Executive Summary

Based on comprehensive research across eight critical technical domains, this deployment strategy provides a practical roadmap for launching a cost-effective Tennis Booker app in London. **Key finding**: No major tennis booking APIs exist, requiring alternative integration approaches, but modern authentication, notification, and deployment solutions offer excellent startup-friendly options with clear scaling paths.

## Product Requirements Document (PRD)

### Product Vision
Tennis Booker is a **real-time court availability monitoring and booking platform** that eliminates the frustration of finding tennis courts in London by providing instant notifications when preferred slots become available, combined with a seamless booking experience.

### Success Metrics
- **User Acquisition**: 1,000 MAU within 3 months, 10,000 MAU within 12 months
- **Booking Success Rate**: >85% of notifications result in successful bookings
- **Notification Latency**: <2 minutes from court availability to user notification
- **System Uptime**: 99.5% availability for critical booking flows
- **Conversion Rate**: 5% free-to-paid conversion by month 6
- **User Retention**: 60% monthly retention rate for active users

### Core User Stories

#### Free Tier Users
1. **As a casual player**, I want to receive email alerts when courts become available on weekends so I can book occasional games
2. **As a new user**, I want to easily set my preferences (location, time, price) without complex onboarding
3. **As a mobile user**, I want to install the PWA and receive push notifications about court availability

#### Premium Users
1. **As a regular player**, I want SMS notifications for immediate alerts so I never miss premium slots
2. **As a group organizer**, I want to coordinate bookings for multiple players and split payments
3. **As a coach**, I want to manage recurring bookings and student schedules efficiently

### Technical Requirements

#### Frontend Requirements

**Technology Stack**:
- React 19 with TypeScript for type safety
- Vite for build optimization
- TanStack Query for server state management
- Zustand for client state
- Tailwind CSS + Radix UI for consistent design system
- Framer Motion for micro-interactions

**Key Features**:
1. **Authentication Flow**
   - Clerk.dev integration with social login buttons
   - Magic link email authentication option
   - Biometric authentication for mobile PWA
   - Session persistence with automatic refresh

2. **Dashboard Interface**
   - Real-time availability grid showing 7-day forecast
   - Filter system: location (max distance), time slots, price range, court surface
   - Saved preferences with quick-select presets
   - Booking history with rebooking shortcuts

3. **Notification Management**
   - Granular notification preferences per saved search
   - Quiet hours configuration
   - Notification frequency limits with visual indicators
   - Test notification system for verification

4. **PWA Implementation**
   - Service worker for offline functionality
   - Background sync for preference updates
   - Push notification permission flow with value proposition
   - App-like navigation with bottom tab bar

5. **Ad Integration**
   - Native ad components between content sections
   - Lazy-loaded ad units for performance
   - Ad-free experience toggle for premium users
   - GDPR-compliant consent management

**Performance Requirements**:
- First Contentful Paint: <1.5s
- Time to Interactive: <3s
- Lighthouse score: >90
- Bundle size: <200KB initial load

#### Backend Requirements

**Technology Stack**:
- Go 1.23 with Gorilla Mux
- MongoDB for persistent storage
- Redis for caching and pub/sub
- BullMQ for job queuing
- Docker containerization

**API Endpoints**:

1. **Authentication Service**
   ```
   POST /api/auth/register - User registration with Clerk webhook
   POST /api/auth/login - Session creation
   POST /api/auth/refresh - Token refresh
   DELETE /api/auth/logout - Session termination
   ```

2. **User Preference Service**
   ```
   GET /api/preferences - Retrieve user preferences
   POST /api/preferences - Create/update preferences
   DELETE /api/preferences/:id - Remove preference set
   POST /api/preferences/test - Trigger test notification
   ```

3. **Booking Service**
   ```
   GET /api/courts/availability - Real-time availability
   POST /api/bookings/reserve - Reserve court slot
   GET /api/bookings/history - User booking history
   POST /api/bookings/share - Generate shareable booking link
   ```

4. **Notification Service**
   ```
   POST /api/notifications/subscribe - PWA push subscription
   PUT /api/notifications/settings - Update notification preferences
   GET /api/notifications/history - Notification log
   POST /api/notifications/batch - Admin batch notifications
   ```

5. **Analytics Service**
   ```
   POST /api/analytics/event - Track user events
   GET /api/analytics/usage - User usage statistics
   GET /api/analytics/performance - System performance metrics
   ```

**Database Schema**:

```javascript
// Users Collection
{
  _id: ObjectId,
  clerkId: String,
  email: String,
  phone: String,
  tier: Enum["free", "player", "club", "tournament"],
  preferences: [{
    name: String,
    locations: [{ lat, lng, radius }],
    timeSlots: [{ day, startTime, endTime }],
    maxPrice: Number,
    courtTypes: Array,
    notificationChannels: Array
  }],
  usage: {
    bookingsThisMonth: Number,
    notificationsSent: Number,
    lastActiveAt: Date
  }
}

// Courts Collection
{
  _id: ObjectId,
  venueId: String,
  venueName: String,
  location: { type: "Point", coordinates: [lng, lat] },
  address: String,
  courtNumber: String,
  surface: String,
  amenities: Array,
  pricePerHour: Number
}

// Availability Collection
{
  _id: ObjectId,
  courtId: ObjectId,
  datetime: Date,
  duration: Number,
  status: Enum["available", "booked", "maintenance"],
  price: Number,
  lastChecked: Date,
  checksumHash: String
}
```

**Rate Limiting**:
- API calls: 100/minute per user (free), 500/minute (premium)
- Notification triggers: 20/day (free), 100/day (premium)
- Booking attempts: 10/hour per user

#### Scraper Service Requirements

**Architecture**:
- Python 3.11+ with Playwright
- Modular scraper design for multiple platforms
- Redis-based deduplication
- Distributed task queue with BullMQ

**Scraping Strategy**:
1. **Intelligent Scheduling**
   - Peak hours (Fri 5-8pm, Sat/Sun 9am-4pm): 2-minute intervals
   - Standard hours: 10-minute intervals
   - Off-peak: 30-minute intervals
   - Adaptive based on historical cancellation patterns

2. **Data Collection**
   - Court availability for next 7 days
   - Price variations by time slot
   - Cancellation detection within 2 minutes
   - Checksum-based change detection

3. **Anti-Detection**
   - Rotating residential proxies
   - Random delays (15-45 seconds)
   - Browser fingerprint randomization
   - Distributed scraping across multiple IPs

**Performance Targets**:
- Scraping latency: <30s per venue
- Detection accuracy: >99%
- Resource usage: <1GB RAM per scraper instance

#### Infrastructure Requirements

**Railway Deployment**:
```yaml
services:
  api:
    source: ./apps/backend
    builder: DOCKERFILE
    replicas: 2
    healthcheck:
      path: /health
      interval: 30s
    env:
      - MONGO_URI
      - REDIS_URL
      - CLERK_SECRET

  scraper:
    source: ./apps/scraper
    builder: DOCKERFILE
    replicas: 1
    cron: "*/5 * * * *"
    env:
      - SCRAPER_MODE=smart
      - PROXY_ENABLED=true

  frontend:
    source: ./apps/frontend
    builder: DOCKERFILE
    domains:
      - tennisbooker.app
    env:
      - VITE_API_URL
      - VITE_CLERK_KEY
```

**MongoDB Atlas Configuration**:
- M10 cluster for production
- 3-node replica set
- Automated daily backups
- Performance advisor enabled

**Redis Configuration**:
- 512MB memory allocation
- Persistence enabled (AOF)
- Separate databases for cache/queue/pub-sub

#### Security Requirements

1. **Authentication & Authorization**
   - Clerk.dev managed authentication
   - JWT tokens with 15-minute expiry
   - Refresh token rotation
   - Role-based access control

2. **Data Protection**
   - TLS 1.3 for all connections
   - Encryption at rest for sensitive data
   - PII data minimization
   - GDPR-compliant data retention (90 days)

3. **API Security**
   - Rate limiting per endpoint
   - Request signing for webhooks
   - CORS configuration for frontend domain
   - Input validation and sanitization

#### Monitoring & Observability

**Required Integrations**:
1. **Sentry** - Error tracking and performance monitoring
2. **Logtail** - Centralized logging with search
3. **UptimeRobot** - Endpoint monitoring
4. **Custom Metrics** - Booking success rate, notification delivery

**Alerting Thresholds**:
- API response time >2s
- Error rate >1%
- Scraper failure rate >10%
- Database connection pool exhaustion

### User Experience Requirements

#### Onboarding Flow
1. **Welcome Screen** - Value proposition with social proof
2. **Location Permission** - Optional but recommended
3. **Preference Setup** - Visual time/location selector
4. **Notification Permission** - Clear benefit explanation
5. **First Match** - Immediate value demonstration

#### Core Interactions
1. **Search Creation** - <30 seconds to set preferences
2. **Notification Receipt** - Deep link to booking page
3. **Booking Completion** - 2-click booking process
4. **Payment Integration** - Saved cards for premium users

#### Accessibility
- WCAG 2.1 AA compliance
- Screen reader optimization
- Keyboard navigation support
- High contrast mode option

### Beta Testing Requirements

**Beta Program Structure**:
1. **Cohort Size**: 100 users (diverse player levels)
2. **Duration**: 4 weeks with weekly feedback cycles
3. **Access**: Special beta tier with all premium features
4. **Feedback Tools**: In-app feedback widget + weekly surveys

**Beta Success Criteria**:
- 80% of testers successfully book a court
- <5 critical bugs discovered
- Average app rating >4.5/5
- 50% would recommend to friends

### Launch Requirements

#### Marketing Site
- Landing page with live availability counter
- Social proof (testimonials, booking stats)
- SEO-optimized for "tennis court booking London"
- Waitlist with referral incentives

#### Legal Compliance
- Terms of Service with booking disclaimers
- Privacy Policy (GDPR/UK GDPR compliant)
- Cookie consent banner
- Age verification (16+)

#### Support Infrastructure
- Help center with FAQs
- In-app chat (Intercom free tier)
- Email support with 24h SLA
- Community Discord server

## Tennis Court Integration Strategy

### API Reality Check
**Major platforms lack public APIs** - neither LTA ClubSpark nor Courtside offer booking APIs. ClubSpark's API only covers educational courses, not court reservations. However, **viable alternatives exist**:

**Recommended Approach**: **Planyo API Integration**
- Full-featured RESTful API with reservation management
- Rate limits: 12,000 calls/day with authentication via API keys
- Supports tennis courts specifically with pricing calculations
- Contact required for pricing but proven tennis venue support

**Alternative Strategies**:
1. **Partnership Program**: Direct data access agreements with venues
2. **Widget Integration**: Bookteq and similar platforms offer embeddable booking widgets
3. **Hybrid Model**: Combine limited APIs with venue partnerships for comprehensive coverage

**Implementation Priority**: Start with Planyo for proof of concept, develop venue partnership program for scale.

## Authentication Architecture

### Clerk.dev Integration (Strongly Recommended)
**Optimal choice** for startup growth trajectory with superior cost structure:

**Pricing Advantage**:
- **Free tier**: 10,000 Monthly Active Users (zero cost until scale)
- **Pro scaling**: $25/month + $0.02/MAU after 10K users
- **Cost projection**: 20K users = $225/month (vs Auth0 at $150+ for 500 users)

**Technical Benefits**:
- **Zero-configuration social logins** (Google, Apple, Facebook)
- **Built-in organization support** essential for tennis clubs/courts
- **Enterprise-grade security** without complexity
- **Fastest implementation**: 2-3 hours for complete setup vs days for alternatives

**Implementation Timeline**:
- Week 1: Basic authentication (2-3 hours)
- Week 2: Social logins (4-6 hours)  
- Month 2: Organizations for tennis clubs (6-8 hours)

## Notification Infrastructure

### Recommended Service Stack
**Cost-effective combination** for startup launch:

**Email Service**: **SMTP2GO** (Best value)
- Free: 1,000 emails/month
- Paid: $15/month for 40,000 emails
- **95.5% deliverability** (highest tested)
- Simple SMTP setup with excellent startup support

**SMS Service**: **MessageBird** (Cost advantage)
- **Significantly cheaper** than Twilio at $0.005/SMS
- 250+ carrier connections globally
- UK coverage with fast delivery

**Cost Analysis for 1,000 Users**:
- **Option A** (Budget): SMTP2GO + MessageBird = ~$23/month
- **Option B** (Premium): Resend + Twilio = ~$89/month
- **Option C** (Unified): Brevo all-in-one = ~$90/month

**Architecture Recommendation**:
- Redis + BullMQ for queue management
- Template engine (Handlebars) for personalization
- Provider fallback strategy for reliability

## Revenue Generation Strategy

### Ad Monetization (Citymapper Model)
**Non-invasive contextual advertising** approach:

**Primary Network**: **Media.net** (Yahoo/Bing)
- **£3-9 CPM** in UK market
- Superior to Google AdSense alternatives
- 1-2 day approval process

**Implementation Strategy**:
- **Native ad formats** between booking sections
- **Contextual targeting** based on location/booking patterns
- **Progressive enhancement**: Start post-user acquisition

**Revenue Projections**:
- 10K monthly users: £200-500/month
- 50K monthly users: £1,000-2,500/month
- 100K+ monthly users: £3,000-8,000/month

### PWA Push Notifications
**Firebase Cloud Messaging** implementation:
- **Cross-platform support** (iOS 16.4+, Android full support)
- **Contextual permission requests** with clear value proposition
- **Frequency limit**: Maximum 5 notifications/week
- **A/B testing** for timing and messaging optimization

## Deployment Infrastructure

### Railway Platform Strategy
**Usage-based pricing** ideal for startup scaling:

**MongoDB Strategy**: **MongoDB Atlas** (Recommended)
- M0 free tier for development
- M10 ($57/month) for small production
- **Fully managed** with automated backups and scaling
- Better than Railway's unmanaged MongoDB templates

**Cost Projections**:
- **Development**: $0-10/month
- **Small production** (1K users): $97-117/month total
- **Scaled production** (5K+ users): $290-400/month total

**Docker Optimization**:
- Multi-stage builds for 60-80% size reduction
- Alpine base images for performance
- Pre-built image strategy for 30-second deployments

**Alternative Consideration**: Fly.io for global distribution at similar costs with better performance characteristics.

## Data Collection Approach

### Smart Scraping Implementation
**Playwright-based solution** for tennis booking platforms:

**Technical Stack**:
- **Playwright**: Superior JavaScript handling and cross-browser support
- **BullMQ**: Modern queue management with built-in rate limiting  
- **Residential proxies**: Essential for booking platform success

**Ethical Framework**:
- **15-30 second delays** minimum for tennis platforms
- **Robots.txt compliance** mandatory
- **Public data focus**: Availability and pricing only, no personal information
- **GDPR compliance** for UK market operations

**Architecture Pattern**:
```
Tennis App → BullMQ Queue → Playwright Workers → Data Validation → Database
```

**Rate Limiting**: Token bucket algorithm with 100 tokens/minute maximum, adaptive throttling based on response times.

## Freemium Business Model

### Four-Tier Structure
**Optimized for UK tennis market**:

**Free Tier** ("Court Finder"):
- 2 bookings/month, 3-day advance booking
- Email notifications only, single venue access
- **Value delivered**: Full booking functionality with calendar integration

**Player Pro** (£9/month):
- 15 bookings/month, 7-day advance booking
- SMS notifications (50/month), multi-venue access
- **Target**: Regular recreational players

**Club Champion** (£24/month):
- Unlimited bookings, 14-day advance booking
- Group management, payment collection tools
- **Target**: Serious players and coaches

**Tournament Master** (£59/month):
- Event management, white-label options
- Full API access, priority support
- **Target**: Tennis academies and tournament organizers

### Conversion Strategy
**Industry benchmarks**: 3.7% average freemium conversion rate
- **Focus on network effects**: Group bookings and friend invitations
- **Usage-based triggers**: "80% of monthly bookings used"
- **Value demonstration**: Personal tennis statistics and booking success rates

## Implementation Roadmap

### Phase 1: Foundation (Month 1-2)
1. **Authentication setup** with Clerk.dev social logins
2. **Basic booking functionality** using Planyo API integration
3. **Railway deployment** with MongoDB Atlas
4. **Email notifications** via SMTP2GO

### Phase 2: Growth Features (Month 2-4)
1. **SMS notifications** integration with MessageBird
2. **PWA implementation** with push notifications
3. **Freemium tier** rollout starting with generous free tier
4. **Venue partnership program** development

### Phase 3: Monetization (Month 4-6)
1. **Ad integration** with Media.net contextual advertising
2. **Smart scraping implementation** for comprehensive court data
3. **Premium feature rollout** across paid tiers
4. **Analytics dashboard** for user engagement optimization

## Cost Summary and Projections

### Development Phase (Month 1-3)
- **Authentication**: Free (Clerk.dev)
- **Deployment**: $10-30/month (Railway + MongoDB Atlas M0)
- **Notifications**: Free tier limits adequate
- **Total**: $10-30/month

### Launch Phase (Month 4-12, 1K users)
- **Authentication**: Free (under 10K MAUs)
- **Infrastructure**: $97-117/month (Railway + MongoDB Atlas M10)
- **Notifications**: $25-40/month (email + SMS)
- **Total**: $122-157/month

### Growth Phase (Year 2, 10K+ users)  
- **Authentication**: $25-50/month (Clerk.dev scaling)
- **Infrastructure**: $200-400/month (scaled Railway + Atlas)
- **Notifications**: $100-200/month (higher volume)
- **Revenue offset**: £1,000-2,500/month (ads + subscriptions)
- **Net position**: Revenue positive

## Risk Mitigation

### Technical Risks
- **API dependencies**: Develop multi-provider integration strategy
- **Scraping reliability**: Implement robust error handling and fallbacks
- **Scaling bottlenecks**: Choose platforms with clear scaling paths

### Business Risks
- **Competition**: Focus on superior UX and local partnership network
- **Venue resistance**: Emphasize value-add rather than replacement
- **Regulatory changes**: Maintain GDPR compliance and data protection standards

This deployment strategy provides a **balanced approach to cost, scalability, and technical risk** while addressing the specific needs of the London tennis market. The modular architecture allows for iterative development and scaling based on user adoption and revenue growth.