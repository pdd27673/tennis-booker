.PHONY: help build clean test coverage lint run-api run-notification run-scraper docker-build docker-run setup-db seed-db deps install

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build all binaries
	@echo "Building all binaries..."
	@mkdir -p bin
	go build -o bin/server cmd/server/main.go
	go build -o bin/notification-service cmd/notification-service/main.go
	go build -o bin/seed-db cmd/seed-db/main.go
	go build -o bin/seed-user cmd/seed-user/main.go
	@echo "✅ All binaries built successfully"

build-server: ## Build the main API server
	@echo "Building main API server..."
	@mkdir -p bin
	go build -o bin/server cmd/server/main.go
	@echo "✅ Server built successfully"

build-notification: ## Build the notification service
	@echo "Building notification service..."
	@mkdir -p bin
	go build -o bin/notification-service cmd/notification-service/main.go
	@echo "✅ Notification service built successfully"

build-scraper: ## Build the scraper service
	@echo "Building scraper service..."
	@mkdir -p bin
	go build -o bin/scraper cmd/scraper/main.go
	@echo "✅ Scraper built successfully"

build-tools: ## Build utility tools
	@echo "Building utility tools..."
	@mkdir -p bin
	go build -o bin/seed-db cmd/seed-db/main.go
	go build -o bin/seed-user cmd/seed-user/main.go
	@echo "✅ Tools built successfully"

# Development targets
run-server: build-server ## Run the main API server
	@echo "Starting API server..."
	./bin/server

run-notification: build-notification ## Run the notification service
	@echo "Starting notification service..."
	./bin/notification-service

run-scraper: build-scraper ## Run the scraper service
	@echo "Starting scraper service..."
	./bin/scraper

# Database targets
setup-db: ## Setup database collections and indexes
	@echo "Setting up database..."
	go run cmd/setup-db/main.go
	@echo "✅ Database setup complete"

seed-db: build-tools ## Seed database with initial data
	@echo "Seeding database..."
	./bin/seed-db
	@echo "✅ Database seeded successfully"

seed-user: build-tools ## Seed user preferences
	@echo "Seeding user preferences..."
	./bin/seed-user
	@echo "✅ User preferences seeded successfully"

# Testing targets
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Quality targets
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...
	@echo "✅ Go vet completed"

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	@echo "✅ Dependencies downloaded"

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "✅ Dependencies tidied"

vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	go mod vendor
	@echo "✅ Dependencies vendored"

# Docker targets
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker build -t tennis-booker-api -f Dockerfile.api .
	docker build -t tennis-booker-notification -f Dockerfile.notification .
	docker build -t tennis-booker-scraper -f Dockerfile.scraper .
	@echo "✅ Docker images built"

docker-run: ## Run services with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d
	@echo "✅ Services started"

docker-stop: ## Stop Docker Compose services
	@echo "Stopping services..."
	docker-compose down
	@echo "✅ Services stopped"

# Cleanup targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "✅ Clean completed"

clean-all: clean ## Clean everything including vendor
	@echo "Cleaning everything..."
	rm -rf vendor/
	@echo "✅ Deep clean completed"

# Development environment
dev: ## Start development environment
	@echo "Starting development environment..."
	make build
	@echo "✅ Development environment ready"

# Install tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

# All-in-one targets
all: clean deps build test ## Clean, download deps, build, and test
	@echo "✅ All tasks completed successfully"

local: build seed-db ## Build and seed for local development
	@echo "✅ Local development setup complete" 