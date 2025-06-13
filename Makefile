.PHONY: help build test clean dev local setup backend-build backend-test scraper-setup scraper-run scraper-test vault-up vault-down vault-logs vault-status vault-clean vault-test vault-secrets

# Default target
help:
	@echo "🎾 Tennis Booking System - Monorepo Commands"
	@echo ""
	@echo "Setup Commands:"
	@echo "  setup           - Set up all applications"
	@echo "  dev             - Start development environment"
	@echo "  local           - Start complete local development (no Vault)"
	@echo ""
	@echo "Backend Commands:"
	@echo "  backend-build   - Build Go backend services"
	@echo "  backend-test    - Run Go backend tests"
	@echo "  backend-run     - Run notification service"
	@echo ""
	@echo "Scraper Commands:"
	@echo "  scraper-setup   - Set up Python scraper environment"
	@echo "  scraper-run     - Run the scraper"
	@echo "  scraper-test    - Run scraper tests"
	@echo ""
	@echo "Vault Integration Commands:"
	@echo "  vault-up        - Start all services with Vault integration"
	@echo "  vault-down      - Stop all Vault-integrated services"
	@echo "  vault-logs      - Show logs for all Vault services"
	@echo "  vault-status    - Show status of all Vault services"
	@echo "  vault-test      - Test Vault Agent integration"
	@echo "  vault-clean     - Clean up Vault volumes and containers"
	@echo ""
	@echo "General Commands:"
	@echo "  build           - Build all applications"
	@echo "  test            - Run all tests"
	@echo "  clean           - Clean all build artifacts"

# Setup all applications
setup: backend-setup scraper-setup
	@echo "✅ All applications set up successfully!"

# Start development environment
dev:
	@echo "🚀 Starting development environment..."
	docker-compose up -d
	@echo "✅ Development environment started!"

# Start complete local development
local:
	@echo "🎾 Starting complete local development environment..."
	./scripts/run_local.sh start

# Build all applications
build: backend-build
	@echo "✅ All applications built successfully!"

# Run all tests
test: backend-test scraper-test
	@echo "✅ All tests completed!"

# Clean all build artifacts
clean: backend-clean scraper-clean
	@echo "✅ All build artifacts cleaned!"

# Backend commands
backend-setup:
	@echo "🔧 Setting up Go backend..."
	cd apps/backend && go mod download

backend-build:
	@echo "🔨 Building Go backend..."
	cd apps/backend && make build

backend-test:
	@echo "🧪 Running Go backend tests..."
	cd apps/backend && make test

backend-run:
	@echo "🚀 Running notification service..."
	cd apps/backend && make run-notification

backend-clean:
	@echo "🧹 Cleaning Go backend..."
	cd apps/backend && make clean

# Scraper commands
scraper-setup:
	@echo "🔧 Setting up Python scraper..."
	cd apps/scraper && make setup

scraper-run:
	@echo "🕷️ Running scraper..."
	cd apps/scraper && make run

scraper-test:
	@echo "🧪 Running scraper tests..."
	cd apps/scraper && make test

scraper-clean:
	@echo "🧹 Cleaning scraper..."
	cd apps/scraper && make clean

# Docker commands
docker-up:
	@echo "🐳 Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "🐳 Stopping Docker services..."
	docker-compose down

docker-logs:
	@echo "📋 Showing Docker logs..."
	docker-compose logs -f

# Vault Integration Commands
vault-up:
	@echo "🚀 Starting Tennis Booker with Vault integration..."
	docker-compose -f docker-compose.yml -f docker-compose.vault-integrated.yml up -d
	@echo "✅ All services started!"
	@echo "🔍 Use 'make vault-status' to check service health"

vault-down:
	@echo "🛑 Stopping all Vault-integrated services..."
	docker-compose -f docker-compose.yml -f docker-compose.vault-integrated.yml down
	@echo "✅ All services stopped!"

vault-logs:
	@echo "📋 Showing logs for all Vault services..."
	docker-compose -f docker-compose.yml -f docker-compose.vault-integrated.yml logs -f

vault-status:
	@echo "📊 Service Status:"
	@docker-compose -f docker-compose.yml -f docker-compose.vault-integrated.yml ps
	@echo ""
	@echo "🔐 Vault Agent Status:"
	@docker ps --filter "name=tennis-vault-agent" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

vault-clean:
	@echo "🧹 Cleaning up Vault volumes and containers..."
	docker-compose -f docker-compose.yml -f docker-compose.vault-integrated.yml down -v
	docker volume prune -f
	@echo "✅ Cleanup complete!"

vault-test:
	@echo "🧪 Testing Vault Agent integration..."
	@echo "Checking if secrets are generated..."
	@docker exec tennis-vault-agent-backend ls -la /vault/secrets/ || echo "❌ Backend secrets not found"
	@docker exec tennis-vault-agent-scraper ls -la /vault/secrets/ || echo "❌ Scraper secrets not found"
	@docker exec tennis-vault-agent-notification ls -la /vault/secrets/ || echo "❌ Notification secrets not found"
	@echo "✅ Test complete!"

vault-secrets:
	@echo "🔍 Generated Secrets (Backend):"
	@docker exec tennis-vault-agent-backend cat /vault/secrets/backend.env 2>/dev/null || echo "❌ Backend secrets not available"
	@echo ""
	@echo "🔍 Generated Secrets (Scraper):"
	@docker exec tennis-vault-agent-scraper cat /vault/secrets/scraper.env 2>/dev/null || echo "❌ Scraper secrets not available"
	@echo ""
	@echo "🔍 Generated Secrets (Notification):"
	@docker exec tennis-vault-agent-notification cat /vault/secrets/notification.env 2>/dev/null || echo "❌ Notification secrets not available"
