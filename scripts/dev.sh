#!/bin/bash

# 🎾 Tennis Booker - Development Utility Script
# Unified script for common development tasks

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
success() { echo -e "${GREEN}✅ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠️  $1${NC}"; }
error() { echo -e "${RED}❌ $1${NC}"; }

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/apps/backend"
SCRAPER_DIR="$PROJECT_ROOT/apps/scraper"
FRONTEND_DIR="$PROJECT_ROOT/apps/frontend"

show_help() {
    cat << EOF
🎾 Tennis Booker Development Utility

Usage: $0 <command> [options]

CORE COMMANDS:
  start                 Start all services (alias: run_local.sh start)
  stop                  Stop all services  
  restart               Restart all services
  status                Show service status
  logs [service]        Show logs (backend|frontend|scraper|notification)

DEVELOPMENT:
  test [app]            Run tests (backend|frontend|scraper|all)
  lint [app]            Run linting (backend|frontend|scraper|all)  
  build [app]           Build applications (backend|frontend|scraper|all)
  clean                 Clean build artifacts and temp files

CI/TESTING:
  ci                    Run full CI checks locally
  integration           Run integration tests
  
UTILITIES:
  setup                 Initial project setup
  check                 Check prerequisites
  db                    Database operations (seed|clear|backup)

EXAMPLES:
  $0 start              # Start development environment
  $0 test backend       # Run backend tests
  $0 logs scraper       # Show scraper logs
  $0 ci                 # Run all CI checks
  $0 db seed            # Seed database with test data

EOF
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."
    
    local missing=()
    
    command_exists docker || missing+=("docker")
    command_exists docker-compose || missing+=("docker-compose")
    command_exists go || missing+=("go")
    command_exists python3 || missing+=("python3")
    command_exists npm || missing+=("npm")
    command_exists jq || missing+=("jq")
    
    if [ ${#missing[@]} -ne 0 ]; then
        error "Missing required tools: ${missing[*]}"
        return 1
    fi
    
    success "All prerequisites installed"
}

# Core service operations
start_services() {
    info "Starting Tennis Booker development environment..."
    "$PROJECT_ROOT/scripts/run_local.sh" start
}

stop_services() {
    info "Stopping all services..."
    "$PROJECT_ROOT/scripts/run_local.sh" stop
}

restart_services() {
    info "Restarting all services..."
    "$PROJECT_ROOT/scripts/run_local.sh" restart
}

show_status() {
    "$PROJECT_ROOT/scripts/run_local.sh" status
}

show_logs() {
    local service="${1:-all}"
    
    case "$service" in
        backend)
            tail -f "$PROJECT_ROOT/logs/backend-server.log" 2>/dev/null || echo "No backend logs found"
            ;;
        frontend)
            tail -f "$PROJECT_ROOT/logs/frontend.log" 2>/dev/null || echo "No frontend logs found"
            ;;
        scraper)
            tail -f "$PROJECT_ROOT/logs/scraper.log" 2>/dev/null || echo "No scraper logs found"
            ;;
        notification)
            tail -f "$PROJECT_ROOT/logs/notification-service.log" 2>/dev/null || echo "No notification logs found"
            ;;
        all|*)
            "$PROJECT_ROOT/scripts/run_local.sh" logs
            ;;
    esac
}

# Test operations
run_tests() {
    local target="${1:-all}"
    
    case "$target" in
        backend)
            info "Running backend tests..."
            cd "$BACKEND_DIR" && make test
            ;;
        frontend)
            info "Running frontend tests..."
            cd "$FRONTEND_DIR" && npm test
            ;;
        scraper)
            info "Running scraper tests..."
            cd "$SCRAPER_DIR" && source venv/bin/activate && python -m pytest tests/ -v
            ;;
        all)
            run_tests backend
            run_tests frontend
            run_tests scraper
            ;;
        *)
            error "Unknown test target: $target"
            return 1
            ;;
    esac
}

# Lint operations
run_lint() {
    local target="${1:-all}"
    
    case "$target" in
        backend)
            info "Running backend linting..."
            cd "$BACKEND_DIR" && make lint
            ;;
        frontend)
            info "Running frontend linting..."
            cd "$FRONTEND_DIR" && npm run lint
            ;;
        scraper)
            info "Running scraper linting..."
            cd "$SCRAPER_DIR" && source venv/bin/activate && python -m flake8 src/ --max-line-length=120
            ;;
        all)
            run_lint backend
            run_lint frontend  
            run_lint scraper
            ;;
        *)
            error "Unknown lint target: $target"
            return 1
            ;;
    esac
}

# Build operations
run_build() {
    local target="${1:-all}"
    
    case "$target" in
        backend)
            info "Building backend..."
            cd "$BACKEND_DIR" && make build
            ;;
        frontend)
            info "Building frontend..."
            cd "$FRONTEND_DIR" && npm run build
            ;;
        scraper)
            info "Setting up scraper..."
            cd "$SCRAPER_DIR" && make setup
            ;;
        all)
            run_build backend
            run_build frontend
            run_build scraper
            ;;
        *)
            error "Unknown build target: $target"
            return 1
            ;;
    esac
}

# Clean operations
clean_project() {
    info "Cleaning Tennis Booker project..."
    "$PROJECT_ROOT/scripts/cleanup.sh"
}

# CI operations
run_ci() {
    info "Running CI checks..."
    "$PROJECT_ROOT/scripts/ci-checks.sh"
}

# Integration tests
run_integration() {
    info "Running integration tests..."
    "$PROJECT_ROOT/scripts/test-integration.sh"
}

# Database operations
database_ops() {
    local operation="${1:-help}"
    
    case "$operation" in
        seed)
            info "Seeding database..."
            cd "$BACKEND_DIR"
            ./bin/seed-db
            export USER_EMAIL="${USER_EMAIL:-mvgnum@gmail.com}"
            ./bin/seed-user
            ;;
        clear)
            info "Clearing database..."
            # Add database clearing logic
            warn "Database clear not implemented yet"
            ;;
        backup)
            info "Backing up database..."
            # Add database backup logic
            warn "Database backup not implemented yet"
            ;;
        help|*)
            echo "Database operations:"
            echo "  seed   - Seed database with test data"
            echo "  clear  - Clear all database data"
            echo "  backup - Backup database"
            ;;
    esac
}

# Initial setup
setup_project() {
    info "Setting up Tennis Booker development environment..."
    
    check_prerequisites
    
    # Setup backend
    info "Setting up backend..."
    cd "$BACKEND_DIR" && make setup
    
    # Setup frontend
    info "Setting up frontend..."
    cd "$FRONTEND_DIR" && npm install
    
    # Setup scraper
    info "Setting up scraper..."
    cd "$SCRAPER_DIR" && make setup
    
    success "Project setup complete!"
    echo ""
    info "Next steps:"
    echo "  $0 start    # Start development environment"
    echo "  $0 test     # Run all tests"
}

# Main command handling
main() {
    case "${1:-help}" in
        # Core operations
        start|up)
            start_services
            ;;
        stop|down)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status)
            show_status
            ;;
        logs|log)
            show_logs "$2"
            ;;
        
        # Development
        test|tests)
            run_tests "$2"
            ;;
        lint)
            run_lint "$2"
            ;;
        build)
            run_build "$2"
            ;;
        clean|cleanup)
            clean_project
            ;;
        
        # CI/Testing
        ci)
            run_ci
            ;;
        integration|e2e)
            run_integration
            ;;
        
        # Utilities
        setup)
            setup_project
            ;;
        check)
            check_prerequisites
            ;;
        db|database)
            database_ops "$2"
            ;;
        
        # Help
        help|-h|--help)
            show_help
            ;;
        
        *)
            error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"