.PHONY: build test clean run-notification run-auth-server run-seed-db run-seed-user run-db-tools format lint tidy check-deps

# Default target
all: format tidy test build

# Build all services
build:
	@echo "Building all backend services..."
	@mkdir -p bin
	go build -o bin/notification-service ./cmd/notification-service
	go build -o bin/seed-db ./cmd/seed-db
	go build -o bin/seed-user ./cmd/seed-user
	go build -o bin/db-tools ./cmd/db-tools
	go build -o bin/test-auth-server ./cmd/test-auth-server
	go build -o bin/config-test ./cmd/config-test

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v

# Run tests in short mode (skip integration tests)
test-short:
	@echo "Running tests in short mode..."
	go test ./... -short -v

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Format code
format:
	@echo "Formatting code..."
	gofmt -w .

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Check for security issues
security-check:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping security check"; \
	fi

# Check dependencies for vulnerabilities
check-deps:
	@echo "Checking dependencies for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed, skipping vulnerability check"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run notification service
run-notification:
	@echo "Running notification service..."
	go run ./cmd/notification-service/main.go

# Run test auth server
run-auth-server:
	@echo "Running test auth server..."
	go run ./cmd/test-auth-server/main.go

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
	go run ./cmd/db-tools/ensure_indexes.go

# Run config test
run-config-test:
	@echo "Running config test..."
	go run ./cmd/config-test/main.go

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	@echo "Development environment setup complete!"

# Full check (format, lint, test, build)
check: format tidy lint test build
	@echo "All checks passed!"

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build all services"
	@echo "  test           - Run all tests"
	@echo "  test-short     - Run tests in short mode"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  format         - Format code with gofmt"
	@echo "  lint           - Lint code with golangci-lint"
	@echo "  tidy           - Tidy go modules"
	@echo "  security-check - Run security checks with gosec"
	@echo "  check-deps     - Check dependencies for vulnerabilities"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev-setup      - Set up development environment"
	@echo "  check          - Run all checks (format, lint, test, build)"
	@echo "  run-*          - Run specific services"
	@echo "  help           - Show this help message" 