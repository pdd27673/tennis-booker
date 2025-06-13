# 🎾 Tennis Booking System

A comprehensive system for monitoring tennis court availability, sending notifications for available slots, and managing booking preferences.

## 📋 Project Overview

This system monitors tennis court availability across multiple platforms (ClubSpark, Courtside), notifies users based on their preferences, and provides a modern dashboard for managing preferences and viewing availability.

## 🏗️ Monorepo Structure

This project is organized as a monorepo with the following structure:

```
tennis-booking-system/
├── apps/                      # Application code
│   ├── backend/               # Go backend services
│   ├── frontend/              # React TypeScript frontend
│   └── scraper/               # Python scraping services
├── packages/                  # Shared libraries and types
├── infrastructure/            # Deployment & infrastructure
│   ├── terraform/             # OCI infrastructure as code
│   └── docker/                # Docker configurations
├── .github/                   # CI/CD workflows
└── docs/                      # Documentation
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Python 3.10+
- Node.js 18+
- Docker & Docker Compose
- MongoDB
- Redis

### Development Setup

1. Clone the repository
   ```bash
   git clone https://github.com/yourusername/tennis-booking-system.git
   cd tennis-booking-system
   ```

2. Set up environment variables
   ```bash
   cp .env-example .env
   # Edit .env with your configuration
   ```

3. Start development services
   ```bash
   make dev
   ```

## 📦 Applications

### Backend (Go)

The backend provides:
- REST API for court data and user preferences
- Authentication and user management
- Notification service
- Database management

### Frontend (React/TypeScript)

The frontend provides:
- User authentication
- Court availability dashboard
- Preference management
- System control

### Scraper (Python)

The scraper service:
- Monitors tennis court availability
- Supports multiple platforms (ClubSpark, Courtside)
- Publishes availability data to Redis

## 📚 Documentation

- [API Documentation](docs/api.md)
- [Deployment Guide](docs/deployment.md)
- [Development Guide](docs/development.md)

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details. 