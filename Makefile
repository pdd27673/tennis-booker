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
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/api ./cmd/api
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/scheduler ./cmd/scheduler
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

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test ./...

.PHONY: test-short
test-short:
	@echo "Running tests (short mode)..."
	$(GO) test -short ./...

# Seed targets
.PHONY: seed-venues
seed-venues:
	@echo "Seeding venues collection..."
	./scripts/seed_venues.sh

# Help target
.PHONY: help
help:
	@echo "Tennis Booking Bot Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build                 Build all binaries"
	@echo "  make test                  Run all tests"
	@echo "  make test-short            Run tests in short mode (skip MongoDB tests)"
	@echo "  make db-ensure-indexes     Create all database indexes"
	@echo "  make db-verify-indexes     Verify database indexes"
	@echo "  make db-verify-indexes-verbose  Verify database indexes with details"
	@echo "  make seed-venues           Seed venues collection"
	@echo "  make help                  Show this help message"

# Default target
.DEFAULT_GOAL := help 