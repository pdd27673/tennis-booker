# 🎾 Tennis Booking System

A comprehensive system for monitoring tennis court availability, sending notifications for available slots, and managing booking preferences.

## 📋 Project Overview

This system monitors tennis court availability across multiple platforms (ClubSpark, Courtside), notifies users based on their preferences, and provides a modern dashboard for managing preferences and viewing availability.

## 🏗️ Monorepo Structure

This project is organized as a monorepo with the following structure:

```
tennis-booking-system/
├── apps/                      # Application code
│   ├── backend/               # Go backend services (includes notification service)
│   └── scraper/               # Python scraping services
├── infrastructure/            # Deployment & infrastructure
│   ├── vault/                 # HashiCorp Vault configuration & integration
│   └── terraform/             # OCI infrastructure as code
├── scripts/                   # Utility scripts
├── .github/                   # CI/CD workflows
└── .taskmaster/               # Project management and tasks
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
   # Complete local development (recommended for development)
   make local
   
   # Or basic Docker services only
   make dev
   
   # Or with Vault integration (production-like setup)
   make vault-up
   ```

## 📦 Applications

### Backend (Go)

The backend provides:
- REST API for court data and user preferences
- Authentication and user management
- Notification service
- Database management

### Scraper (Python)

The scraper service:
- Monitors tennis court availability
- Supports multiple platforms (ClubSpark, Courtside)
- Publishes availability data to Redis

## 🔐 Security & Vault Integration

This project uses HashiCorp Vault for secure secret management:

- **No hardcoded secrets**: All sensitive data managed by Vault
- **Vault Agent integration**: Automatic secret injection for all services
- **Non-root containers**: Enhanced security posture
- **Production-ready**: Enterprise-grade security controls

See [infrastructure/vault/README.md](infrastructure/vault/README.md) for detailed setup and usage.

## 🛠️ Available Commands

### Development Commands
- `make setup` - Set up all applications
- `make local` - Start complete local development (MongoDB, Redis, notification, scraper)
- `make dev` - Start basic Docker services only
- `make build` - Build all applications
- `make test` - Run all tests

### Vault Integration Commands
- `make vault-up` - Start all services with Vault integration
- `make vault-down` - Stop all Vault-integrated services
- `make vault-status` - Show status of all services
- `make vault-test` - Test Vault Agent integration
- `make vault-logs` - Show logs for all services
- `make vault-clean` - Clean up volumes and containers

### Service-Specific Commands
- `make backend-build` - Build Go backend services
- `make backend-test` - Run Go backend tests
- `make scraper-setup` - Set up Python scraper environment
- `make scraper-run` - Run the scraper
- `make scraper-test` - Run scraper tests

## 📚 Documentation

- [Vault Integration Guide](infrastructure/vault/README.md)
- [Backend Security Notes](apps/backend/SECURITY_NOTES.md)
- [Scraper Documentation](apps/scraper/README.md)

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details. 