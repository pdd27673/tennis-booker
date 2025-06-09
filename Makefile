# Tennis Booking Bot Makefile

# Variables
GO=go
GOFLAGS=-v
BUILD_DIR=./bin
DB_TOOLS_DIR=./cmd/db-tools
ENSURE_INDEXES=$(DB_TOOLS_DIR)/ensure_indexes.go

# Build targets
.PHONY: build
build:
	@echo "Building application..."
	mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/notification-service ./cmd/notification-service
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/seed-db ./cmd/seed-db
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/seed-user ./cmd/seed-user
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/ensure-indexes $(ENSURE_INDEXES)

# Database targets
.PHONY: db-ensure-indexes
db-ensure-indexes:
	@echo "Creating database indexes..."
	$(GO) run $(GOFLAGS) $(ENSURE_INDEXES)

.PHONY: db-verify-indexes
db-verify-indexes:
	@echo "Verifying database indexes..."
	$(GO) run $(GOFLAGS) $(ENSURE_INDEXES) --verify

.PHONY: db-verify-indexes-verbose
db-verify-indexes-verbose:
	@echo "Verifying database indexes (verbose)..."
	$(GO) run $(GOFLAGS) $(ENSURE_INDEXES) --verify --verbose

# Seed targets
.PHONY: seed-db
seed-db:
	@echo "Seeding database with venues..."
	$(GO) run ./cmd/seed-db

.PHONY: seed-user
seed-user:
	@echo "Seeding user preferences..."
	$(GO) run ./cmd/seed-user

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test ./...

.PHONY: test-short
test-short:
	@echo "Running tests (short mode)..."
	$(GO) test -short ./...

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

# Help target
.PHONY: help
help:
	@echo "Tennis Court Availability Alert System Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build                 Build all binaries (notification-service, seed tools)"
	@echo "  make test                  Run all tests"
	@echo "  make test-short            Run tests in short mode"
	@echo "  make db-ensure-indexes     Create all database indexes"
	@echo "  make db-verify-indexes     Verify database indexes"
	@echo "  make db-verify-indexes-verbose  Verify database indexes with details"
	@echo "  make seed-db               Seed venues collection"
	@echo "  make seed-user             Seed user preferences"
	@echo "  make clean                 Clean build artifacts"
	@echo "  make help                  Show this help message"

# Default target
.DEFAULT_GOAL := help 