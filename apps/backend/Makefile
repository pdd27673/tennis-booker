.PHONY: build test clean run-notification

# Default target
all: build

# Build all services
build:
	@echo "Building all backend services..."
	go build -o bin/notification-service ./cmd/notification-service
	go build -o bin/seed-db ./cmd/seed-db
	go build -o bin/seed-user ./cmd/seed-user
	go build -o bin/db-tools ./cmd/db-tools

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Run notification service
run-notification:
	@echo "Running notification service..."
	go run ./cmd/notification-service/main.go

# Run seed-db
run-seed-db:
	@echo "Running seed-db..."
	go run ./cmd/seed-db/main.go

# Run seed-user
run-seed-user:
	@echo "Running seed-user..."
	go run ./cmd/seed-user/main.go

# Run db-tools
run-db-tools:
	@echo "Running db-tools..."
	go run ./cmd/db-tools/main.go 