# Tennis Booking System - Production Deployment Plan

## 🎯 Executive Summary

Transform the existing tennis court scraping system into a production-ready, secure, and scalable application with:
- **Modern React TypeScript frontend** with Aceternity + ShadCN UI
- **Enhanced security** with proper secrets management (no hardcoded passwords)
- **Smart database optimization** with intelligent cleanup and deduplication
- **Zero-cost OCI deployment** with CI/CD pipeline
- **GitHub-ready codebase** with best practices

## 📋 Current System Analysis

### ✅ Strengths
- Excellent Go backend with proper models, MongoDB integration, Redis pub/sub
- Sophisticated scraper architecture with platform abstraction (ClubSpark, Courtside)
- Production-ready notification service with Gmail SMTP integration
- Vault integration for secrets management (underutilized)
- Docker containerization with health checks
- Comprehensive data models with proper MongoDB collections

### ⚠️ Critical Security Issues
- **Hardcoded Gmail password** in notification service (`eswk jgaw zbet wgxo`)
- **Default passwords** in Docker configs
- **No proper authentication** for API endpoints
- **Secrets in environment variables** instead of Vault

## 🏗️ Implementation Plan

### Phase 1: Foundation & Security (Week 1-2)
**Goal**: Secure monorepo structure with proper secrets management

#### 1.1 Monorepo Restructure
```
tennis-booking-system/
├── apps/
│   ├── backend/              # Go services (existing code reorganized)
│   ├── frontend/             # New React TypeScript Vite app
│   └── scraper/              # Python scraping service (existing)
├── packages/                 # Shared packages
├── infrastructure/           # Deployment & Infrastructure
├── .github/                  # CI/CD workflows
├── docs/                     # Documentation
└── tools/                    # Development tools
```

#### 1.2 Security Overhaul
- Remove all hardcoded credentials
- Enhance Vault integration for all secrets
- Implement environment-based configuration
- Create secure Docker setup

### Phase 2: Modern Frontend (Week 3-4)
**Goal**: Beautiful React TypeScript dashboard with authentication

#### 2.1 Tech Stack
- React 18 + TypeScript + Vite
- ShadCN UI components + Aceternity effects
- Zustand for state management
- React Query for API calls
- JWT authentication

#### 2.2 Key Features
- Dashboard with real-time court monitoring
- User preference management
- System control interface (pause/resume scraping)
- Mobile-responsive PWA

### Phase 3: Enhanced Backend API (Week 5-6)
**Goal**: Secure REST API with proper authentication

#### 3.1 Authentication System
- JWT-based authentication with refresh tokens
- User management endpoints
- Protected routes with middleware

#### 3.2 API Endpoints
- Auth: login, refresh, logout
- Preferences: get, update user preferences
- System: pause/resume scraping, health status
- Courts: get available slots, venues

### Phase 4: Database Optimization (Week 5-6)
**Goal**: Smart cleanup and efficient data management

#### 4.1 Intelligent Data Retention
- Only store slots matching user preferences OR already notified
- Automatic cleanup of irrelevant data
- Enhanced deduplication logic

#### 4.2 Queue Management
- Error handling with exponential backoff
- Dead letter queue for failed messages
- Redis queue optimization

### Phase 5: OCI Infrastructure (Week 7)
**Goal**: Zero-cost production deployment

#### 5.1 Infrastructure
- Terraform for OCI Always Free resources
- ARM Ampere A1 instance (2 OCPUs, 12GB RAM)
- Docker Compose with Traefik for SSL
- DuckDNS for free domain

#### 5.2 Services
- Frontend (React app)
- Backend API (Go)
- Notification service (Go)
- Scraper service (Python)
- MongoDB + Redis + Vault

### Phase 6: CI/CD Pipeline (Week 8)
**Goal**: Automated testing and deployment

#### 6.1 GitHub Actions
- Automated testing (Go, TypeScript, Python)
- Security scanning (Trivy, TruffleHog, Semgrep)
- Docker image building and pushing
- Automated deployment to OCI

#### 6.2 Quality Gates
- 95%+ test coverage
- Zero security vulnerabilities
- Performance benchmarks
- Code quality checks

## 🔧 Technical Implementation Details

### Frontend Architecture
```typescript
// State Management with Zustand
interface AuthStore {
  user: User | null;
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  updatePreferences: (preferences: Partial<UserPreferences>) => Promise<void>;
}

// API Layer with React Query
const { data: courts, isLoading } = useQuery({
  queryKey: ['courts'],
  queryFn: () => api.get('/courts'),
});
```

### Backend Security
```go
// JWT Authentication Middleware
func (a *AuthMiddleware) JWTAuth() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        token := extractToken(c)
        claims, err := a.validateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        c.Set("user_id", claims.UserID)
        c.Next()
    })
}

// Secrets Management
func (s *SecretManager) GetEmailCredentials() (*EmailCredentials, error) {
    secret, err := s.getSecret("secret/tennis-booking/email")
    return &EmailCredentials{
        Email:    secret["email"].(string),
        Password: secret["password"].(string),
    }, err
}
```

### Database Optimization
```go
// Smart Cleanup Service
func (s *CleanupService) cleanupIrrelevantSlots() error {
    // Only keep slots that:
    // 1. Match at least one user's preferences
    // 2. Have been notified about
    // 3. Are within the next 7 days
    
    filter := bson.M{
        "$and": []bson.M{
            {"notified": false},
            {"matches_preferences": false},
            {"date": bson.M{"$lt": time.Now().AddDate(0, 0, 7)}},
        },
    }
    
    result, err := s.db.Collection("slots").DeleteMany(context.Background(), filter)
    s.logger.Printf("Cleaned up %d irrelevant slots", result.DeletedCount)
    return err
}
```

### Infrastructure as Code
```hcl
# Terraform OCI Configuration
resource "oci_core_instance" "tennis_booking" {
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  compartment_id      = var.compartment_id
  display_name        = "tennis-booking-system"
  shape               = "VM.Standard.A1.Flex"
  
  shape_config {
    ocpus         = 2
    memory_in_gbs = 12
  }
}
```

## 💰 Cost Analysis

### Infrastructure: $0/month
- OCI Always Free Tier (ARM instance, storage, networking)
- DuckDNS free subdomain
- Let's Encrypt free SSL
- GitHub free CI/CD

### Optional Services: $0-45/month
- Firecrawl API: $10-20/month (enhanced scraping)
- SendGrid: $0-15/month (email notifications)
- Twilio: $0-10/month (SMS notifications)

## 🎯 Success Metrics

### Technical
- 99.9% uptime
- <2 second page load times
- <1 minute notification delivery
- Zero security vulnerabilities
- 95%+ test coverage

### Security
- No hardcoded secrets
- Encrypted data at rest
- JWT-based authentication
- Regular security scans
- Audit logging

### User Experience
- Intuitive dashboard interface
- Real-time preference updates
- Mobile-responsive design
- PWA installation capability
- Offline functionality

## 📅 Timeline Summary

| Week | Phase | Focus | Deliverables |
|------|-------|-------|--------------|
| 1-2  | Foundation | Security & Structure | Monorepo, Vault integration, secure configs |
| 3-4  | Frontend | React Dashboard | Authentication, preferences UI, court monitoring |
| 5-6  | Backend | API & Database | REST API, JWT auth, smart cleanup |
| 7    | Infrastructure | OCI Deployment | Terraform, Docker, SSL, domain |
| 8    | CI/CD | Automation | GitHub Actions, testing, deployment |

## 🚀 Next Steps

1. **Initialize Task Master** with this plan
2. **Start with Phase 1**: Monorepo restructure and security
3. **Implement incrementally** following the task breakdown
4. **Test thoroughly** at each phase
5. **Deploy to production** with zero downtime

This plan transforms the existing excellent system into a production-ready, secure, and scalable application while maintaining zero hosting costs through OCI Always Free resources.