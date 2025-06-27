# 🎾 Tennis Booker

A full-stack tennis court booking and monitoring system with real-time availability tracking, user authentication, and automated court scraping.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.19+-blue.svg)](https://golang.org)
[![React](https://img.shields.io/badge/React-18+-blue.svg)](https://reactjs.org)
[![Python](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://python.org)

## 🏗️ Architecture Overview

This project consists of multiple integrated applications:

- **Frontend** (`apps/frontend/`) - React/TypeScript SPA with modern UI components
- **Backend** (`apps/backend/`) - Go REST API with JWT authentication and court monitoring
- **Scraper** (`apps/scraper/`) - Court availability monitoring service

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose** (for MongoDB, Redis, Vault)
- **Node.js** (v18 or higher)
- **Go** (v1.19 or higher)
- **Python 3** (for scraper)
- **npm** or **yarn**

### Option 1: One-Command Start (Recommended)

```bash
git clone <repository-url>
cd tennis-booker

# Start all services with one command
./scripts/run_local.sh

# Or use available commands:
./scripts/run_local.sh start    # Start all services (default)
./scripts/run_local.sh status   # Check service status
./scripts/run_local.sh logs     # View recent logs
./scripts/run_local.sh stop     # Stop all services
./scripts/run_local.sh restart  # Restart all services
```

This will automatically:
- Start MongoDB, Redis, and Vault in Docker
- Build and start the integrated backend server
- Build and start the notification service
- Set up and start the Python scraper
- Start the React frontend
- Configure all environment variables
- Seed the database with test data

### Option 2: Manual Setup

If you prefer to run services individually:

#### 1. Clone & Install

```bash
git clone <repository-url>
cd tennis-booker

# Install frontend dependencies
cd apps/frontend
npm install
cd ../..

# Install backend dependencies
cd apps/backend
go mod download
make build
cd ../..

# Setup scraper
cd apps/scraper
make setup
cd ../..
```

#### 2. Start Infrastructure Services

```bash
# Start MongoDB, Redis, and Vault
docker-compose up -d mongodb redis vault
```

#### 3. Start Application Services

```bash
# Terminal 1 - Integrated Backend Server
cd apps/backend
./bin/server

# Terminal 2 - Frontend
cd apps/frontend
npm run dev

# Terminal 3 - Scraper (optional)
cd apps/scraper
make run

# Terminal 4 - Notification Service (optional)
cd apps/backend
./bin/notification-service
```

### Access the Application

- **Frontend**: http://localhost:5173 (or as shown in terminal output)
- **Backend API**: http://localhost:8080
- **API Health**: http://localhost:8080/api/health
- **Vault UI**: http://localhost:8200 (token: set `VAULT_TOKEN` environment variable)

### Default Login Credentials

After running the setup, create your account using the registration flow at http://localhost:5173

## 🔧 Backend-Frontend Integration

### ✅ Completed Integration Features

1. **🔐 Authentication Flow**
   - Real JWT-based login/register
   - Token refresh mechanism
   - Secure logout with token cleanup

2. **📊 Dashboard Data**
   - Live court availability from real API
   - Real-time system status monitoring
   - Dashboard statistics (active courts, available slots, venues)

3. **⚙️ User Settings**
   - User preferences connected to backend
   - Profile management with real API calls

4. **🎛️ System Control**
   - Pause/Resume scraping system
   - Real-time system status updates
   - Visual system state indicators

5. **🌐 Environment Configuration**
   - Configurable API base URL
   - Feature flags for mock/real API switching
   - Development vs production settings

### API Endpoints Integration

| Frontend Service | Backend Endpoint | Status |
|------------------|------------------|---------|
| `authApi.login()` | `POST /api/auth/login` | ✅ |
| `authApi.register()` | `POST /api/auth/register` | ✅ |
| `authApi.refreshToken()` | `POST /api/auth/refresh` | ✅ |
| `authApi.getMe()` | `GET /api/auth/me` | ✅ |
| `courtApi.getVenues()` | `GET /api/venues` | ✅ |
| `courtApi.getCourtSlots()` | `GET /api/courts` | ✅ |
| `systemApi.getSystemStatus()` | `GET /api/system/status` | ✅ |
| `systemApi.pauseScraping()` | `POST /api/system/pause` | ✅ |
| `systemApi.resumeScraping()` | `POST /api/system/resume` | ✅ |
| `systemApi.restartScraping()` | `POST /api/system/restart` | ✅ |
| `userApi.getUserPreferences()` | `GET /api/users/preferences` | ✅ |
| `userApi.updateUserPreferences()` | `PUT /api/users/preferences` | ✅ |

## 🧪 Testing the Integration

### 1. Authentication Test

```bash
# Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test@example.com","email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test@example.com","password":"password123"}'
```

### 2. System Control Test

   ```bash
# Get system status
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/system/status

# Pause system
curl -X POST -H "Authorization: Bearer <token>" http://localhost:8080/api/system/pause

# Resume system
curl -X POST -H "Authorization: Bearer <token>" http://localhost:8080/api/system/resume
```

### 3. Frontend Integration Test

1. Navigate to http://localhost:5173
2. Register/Login with credentials
3. Verify Dashboard shows real data
4. Test pause/resume system controls
5. Check Settings page functionality

## 🔄 Development Workflow

### Frontend Development

   ```bash
cd apps/frontend

# Start development server
npm run dev

# Run type checking
npm run type-check

# Build for production
npm run build
```

### Backend Development

   ```bash
cd apps/backend

# Run in development mode
make run

# Run tests
make test

# Build for production
make build
```

### Switching Between Mock and Real APIs

Set in `apps/frontend/.env.local`:

```bash
# Use real backend APIs (default)
VITE_MOCK_API_ENABLED=false

# Use mock data for development
VITE_MOCK_API_ENABLED=true
```

## 🚨 Troubleshooting

### Common Issues

1. **CORS Errors**
   - Ensure backend is running on correct port (8080)
   - Check CORS configuration in backend

2. **Authentication Failures**
   - Verify JWT secret is set in backend `.env`
   - Check token expiration settings

3. **API Connection Issues**
   - Confirm `VITE_API_URL` points to correct backend URL
   - Ensure backend health endpoint responds: `curl http://localhost:8080/api/health`

4. **Mock vs Real API Confusion**
   - Check `VITE_MOCK_API_ENABLED` flag in frontend `.env.local`
   - Use browser dev tools to inspect network requests

### Debug Mode

Enable debug logging:

```bash
# Frontend
VITE_DEBUG_MODE=true
VITE_LOG_LEVEL=debug

# Backend  
LOG_LEVEL=debug
```

## 📁 Project Structure

```
tennis-booker/
├── apps/
│   ├── frontend/          # React TypeScript SPA
│   │   ├── src/
│   │   │   ├── services/  # API integration layer
│   │   │   ├── hooks/     # Custom React hooks
│   │   │   ├── pages/     # Page components
│   │   │   └── config/    # Environment configuration
│   │   └── env.local.example
│   ├── backend/           # Go REST API
│   │   ├── internal/
│   │   │   ├── handlers/  # HTTP handlers
│   │   │   ├── auth/      # Authentication logic
│   │   │   └── models/    # Data models
│   │   └── env.example
│   └── scraper/           # Court monitoring service
└── README.md
```

## 🚀 Deployment

For detailed deployment instructions, see the [deployment documentation](docs/DEPLOYMENT.md).

**Quick summary:**
- **Backend**: Oracle Cloud Infrastructure (OCI) with Docker Compose
- **Frontend**: Vercel with global CDN
- **Cost**: Free tier eligible (~$0/month for small usage)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test both frontend and backend integration
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For issues and questions:

1. Check the troubleshooting section above
2. Review the individual app READMEs:
   - [Frontend README](apps/frontend/README.md)
   - [Backend README](apps/backend/README.md)
3. Open an issue with full error details and reproduction steps 