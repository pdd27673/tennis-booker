.PHONY: help build test clean dev setup backend-build backend-test scraper-setup scraper-run scraper-test

# Default target
help:
	@echo "🎾 Tennis Booking System - Monorepo Commands"
	@echo ""
	@echo "Setup Commands:"
	@echo "  setup           - Set up all applications"
	@echo "  dev             - Start development environment"
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
