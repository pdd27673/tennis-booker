#!/bin/bash

# CI Checks Script for Tennis Booker
# This script runs the same checks as the GitHub CI workflow locally
# Based on .github/workflows/ci.yml

# CI checks script - error handling is managed per command

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_ROOT/apps/backend"
FRONTEND_DIR="$PROJECT_ROOT/apps/frontend"
SCRAPER_DIR="$PROJECT_ROOT/apps/scraper"

# Flags
RUN_BACKEND=true
RUN_FRONTEND=true
RUN_SCRAPER=true
RUN_SECURITY=true
VERBOSE=false
FAIL_FAST=true

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    echo ""
    print_status "$BLUE" "===================================================="
    print_status "$BLUE" "$1"
    print_status "$BLUE" "===================================================="
}

print_step() {
    echo ""
    print_status "$YELLOW" "▶ $1"
}

print_success() {
    print_status "$GREEN" "✅ $1"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
}

print_error() {
    print_status "$RED" "❌ $1"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
}

# Function to run a command and handle errors
run_check() {
    local description=$1
    local command=$2
    local directory=${3:-$PROJECT_ROOT}
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    print_step "$description"
    
    if [ "$VERBOSE" = true ]; then
        echo "Running: $command"
        echo "In directory: $directory"
    fi
    
    if (cd "$directory" && eval "$command" &>/dev/null); then
        print_success "$description"
        return 0
    else
        print_error "$description"
        if [ "$VERBOSE" = true ] || [ "$FAIL_FAST" = false ]; then
            echo "Command failed: $command"
            echo "In directory: $directory"
            if [ "$VERBOSE" = true ]; then
                echo "Error output:"
                (cd "$directory" && eval "$command") || true
            fi
        fi
        if [ "$FAIL_FAST" = true ]; then
            exit 1
        fi
        return 1
    fi
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if directory exists
dir_exists() {
    [ -d "$1" ]
}

# Usage function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Run CI checks locally for the Tennis Booker project (matches GitHub CI workflow).

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    --no-backend            Skip backend checks
    --no-frontend           Skip frontend checks
    --no-scraper            Skip scraper checks
    --no-security           Skip security checks
    --no-fail-fast          Continue running checks even if some fail
    --backend-only          Run only backend checks
    --frontend-only         Run only frontend checks
    --scraper-only          Run only scraper checks

EXAMPLES:
    $0                      # Run all checks (same as GitHub CI)
    $0 --verbose            # Run all checks with verbose output
    $0 --backend-only       # Run only backend checks
    $0 --no-fail-fast       # Continue on failures

EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --no-backend)
            RUN_BACKEND=false
            shift
            ;;
        --no-frontend)
            RUN_FRONTEND=false
            shift
            ;;
        --no-scraper)
            RUN_SCRAPER=false
            shift
            ;;
        --no-security)
            RUN_SECURITY=false
            shift
            ;;
        --no-fail-fast)
            FAIL_FAST=false
            shift
            ;;
        --backend-only)
            RUN_BACKEND=true
            RUN_FRONTEND=false
            RUN_SCRAPER=false
            RUN_SECURITY=false
            shift
            ;;
        --frontend-only)
            RUN_BACKEND=false
            RUN_FRONTEND=true
            RUN_SCRAPER=false
            RUN_SECURITY=false
            shift
            ;;
        --scraper-only)
            RUN_BACKEND=false
            RUN_FRONTEND=false
            RUN_SCRAPER=true
            RUN_SECURITY=false
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_header "Tennis Booker - Local CI Checks (GitHub CI Compatible)"
    
    echo "Project Root: $PROJECT_ROOT"
    echo "Components to check:"
    [ "$RUN_BACKEND" = true ] && echo "  ✓ Backend Tests (Go)"
    [ "$RUN_FRONTEND" = true ] && echo "  ✓ Frontend Tests (TypeScript/React)"
    [ "$RUN_SCRAPER" = true ] && echo "  ✓ Scraper Tests (Python)"
    [ "$RUN_SECURITY" = true ] && echo "  ✓ Security Scan"
    echo ""

    # Pre-flight checks
    print_header "Pre-flight Checks"
    
    if [ "$RUN_BACKEND" = true ]; then
        if ! command_exists go; then
            print_error "Go is not installed or not in PATH"
            exit 1
        fi
        print_success "Go is available ($(go version | cut -d' ' -f3))"
        
        if ! dir_exists "$BACKEND_DIR"; then
            print_error "Backend directory not found: $BACKEND_DIR"
            exit 1
        fi
        print_success "Backend directory found"
    fi
    
    if [ "$RUN_FRONTEND" = true ]; then
        if ! command_exists node; then
            print_error "Node.js is not installed or not in PATH"
            exit 1
        fi
        print_success "Node.js is available ($(node --version))"
        
        if ! command_exists npm; then
            print_error "npm is not installed or not in PATH"
            exit 1
        fi
        print_success "npm is available ($(npm --version))"
        
        if ! dir_exists "$FRONTEND_DIR"; then
            print_error "Frontend directory not found: $FRONTEND_DIR"
            exit 1
        fi
        print_success "Frontend directory found"
    fi
    
    if [ "$RUN_SCRAPER" = true ]; then
        if ! command_exists python3; then
            print_error "Python 3 is not installed or not in PATH"
            exit 1
        fi
        print_success "Python 3 is available ($(python3 --version))"
        
        if ! dir_exists "$SCRAPER_DIR"; then
            print_error "Scraper directory not found: $SCRAPER_DIR"
            exit 1
        fi
        print_success "Scraper directory found"
    fi

    # Backend Tests (matches GitHub CI backend-tests job)
    if [ "$RUN_BACKEND" = true ]; then
        print_header "Backend Tests (Go)"
        
        # Install dependencies
        run_check "Install Go dependencies" "go mod download" "$BACKEND_DIR"
        
        # Run tests (same as GitHub CI)
        run_check "Run Go tests" "go test ./... -v" "$BACKEND_DIR"
        
        # Run tests with coverage
        run_check "Run Go tests with coverage" "go test -coverprofile=coverage.out -covermode=atomic ./..." "$BACKEND_DIR"
        
        # Generate coverage HTML (optional for local)
        if [ "$VERBOSE" = true ]; then
            run_check "Generate coverage HTML" "go tool cover -html=coverage.out -o coverage.html" "$BACKEND_DIR"
        fi
    fi

    # Frontend Tests (matches GitHub CI frontend-tests job)
    if [ "$RUN_FRONTEND" = true ]; then
        print_header "Frontend Tests (TypeScript/React)"
        
        # Install dependencies if node_modules doesn't exist
        if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
            print_step "Installing frontend dependencies"
            (cd "$FRONTEND_DIR" && npm ci)
        fi
        
        # Run linting (same as GitHub CI)
        run_check "Run ESLint" "npm run lint" "$FRONTEND_DIR"
        
        # Build application (same as GitHub CI)
        run_check "Build application" "npm run build" "$FRONTEND_DIR"
        
        # Note: Tests are commented out in GitHub CI, so we don't run them here either
        if [ "$VERBOSE" = true ]; then
            echo "Note: Frontend tests are not yet implemented in the CI workflow"
        fi
    fi

    # Scraper Tests (matches GitHub CI scraper-tests job)
    if [ "$RUN_SCRAPER" = true ]; then
        print_header "Scraper Tests (Python)"
        
        # Check if virtual environment exists and create it if not
        if [ ! -d "$SCRAPER_DIR/venv" ]; then
            print_step "Creating Python virtual environment"
            (cd "$SCRAPER_DIR" && python3 -m venv venv)
        fi
        
        # Install dependencies (same as GitHub CI)
        run_check "Install Python dependencies" "source venv/bin/activate && python -m pip install --upgrade pip && pip install -r requirements.txt" "$SCRAPER_DIR"
        
        # Install Playwright browsers (same as GitHub CI)
        run_check "Install Playwright browsers" "source venv/bin/activate && python -m playwright install --with-deps chromium" "$SCRAPER_DIR"
        
        # Run tests (same as GitHub CI)
        run_check "Run Python tests" "source venv/bin/activate && python -m pytest tests/" "$SCRAPER_DIR"
    fi

    # Security Scan (matches GitHub CI security-scan job)
    if [ "$RUN_SECURITY" = true ]; then
        print_header "Security Scan"
        
        # Check for Trivy (filesystem scan like in GitHub CI)
        if command_exists trivy; then
            run_check "Run Trivy vulnerability scanner" "trivy fs --format table ." "$PROJECT_ROOT"
        else
            print_step "Trivy vulnerability scanner"
            print_status "$YELLOW" "⚠️  Trivy not installed - install with: brew install trivy"
        fi
        
        # Check for TruffleHog (secret scan like in GitHub CI)
        if command_exists trufflehog; then
            run_check "Run TruffleHog secret scan" "trufflehog filesystem . --only-verified" "$PROJECT_ROOT"
        else
            print_step "TruffleHog secret scan"
            print_status "$YELLOW" "⚠️  TruffleHog not installed - install with: brew install trufflehog"
        fi
    fi

    # Final summary
    print_header "CI Checks Summary"
    
    echo "Total checks run: $TOTAL_CHECKS"
    print_status "$GREEN" "Passed: $PASSED_CHECKS"
    
    if [ $FAILED_CHECKS -gt 0 ]; then
        print_status "$RED" "Failed: $FAILED_CHECKS"
        echo ""
        print_status "$RED" "❌ CI checks failed! Please fix the issues above."
        echo ""
        echo "This matches the same checks that run in GitHub Actions CI."
        echo "Fix these issues to ensure your PR will pass CI."
        exit 1
    else
        echo ""
        print_status "$GREEN" "🎉 All CI checks passed! Your code is ready for GitHub CI."
        echo ""
        echo "These checks match the GitHub Actions CI workflow."
        echo "Your changes should pass the CI pipeline."
        exit 0
    fi
}

# Run main function
main "$@" 