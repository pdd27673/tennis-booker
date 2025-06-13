#!/bin/bash

# 🎾 Tennis Booker - Local Development Script
# Starts MongoDB, Redis, notification service, and scraper for local development

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

# Default environment variables for local development
export MONGO_ROOT_USERNAME="admin"
export MONGO_ROOT_PASSWORD="password"
export REDIS_PASSWORD="password"
export MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin"
export REDIS_ADDR="localhost:6379"
export DB_NAME="tennis_booking"

# Email configuration for notifications
export GMAIL_EMAIL="mvgnum@gmail.com"
export GMAIL_PASSWORD="eswk jgaw zbet wgxo"

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker is required but not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is required but not installed"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        error "Go is required but not installed"
        exit 1
    fi
    
    if ! command -v python3 &> /dev/null; then
        error "Python 3 is required but not installed"
        exit 1
    fi
    
    success "Prerequisites check passed"
}

# Start Docker services (MongoDB and Redis only)
start_docker_services() {
    info "Starting Docker services (MongoDB, Redis)..."
    cd "$PROJECT_ROOT"
    
    # Start only MongoDB and Redis
    docker-compose up -d mongodb redis
    
    # Wait for services to be ready
    info "Waiting for services to be ready..."
    
    # Wait for MongoDB
    for i in {1..30}; do
        if docker exec tennis-mongodb mongosh --quiet --eval "db.adminCommand('ping')" &>/dev/null; then
            success "MongoDB is ready"
            break
        fi
        echo -n "."
        sleep 2
    done
    
    # Wait for Redis
    for i in {1..30}; do
        if docker exec tennis-redis redis-cli -a password ping &>/dev/null; then
            success "Redis is ready"
            break
        fi
        echo -n "."
        sleep 2
    done
}

# Build backend services
build_backend() {
    info "Building backend services..."
    cd "$BACKEND_DIR"
    make build
    success "Backend services built"
}

# Seed database
seed_database() {
    info "Seeding database with venues..."
    cd "$BACKEND_DIR"
    ./bin/seed-db
    success "Database seeded with venues"
    
    info "Seeding user preferences..."
    export USER_EMAIL="mvgnum@gmail.com"
    ./bin/seed-user
    success "User preferences seeded"
}

# Setup scraper environment
setup_scraper() {
    info "Setting up scraper environment..."
    cd "$SCRAPER_DIR"
    
    if [ ! -d "venv" ]; then
        make setup
        success "Scraper environment set up"
    else
        success "Scraper environment already exists"
    fi
}

# Start notification service in background
start_notification_service() {
    info "Starting notification service..."
    cd "$BACKEND_DIR"
    
    # Kill existing notification service if running
    pkill -f "bin/notification-service" || true
    
    # Create logs directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/logs"
    
    # Start notification service in background
    nohup ./bin/notification-service > "$PROJECT_ROOT/logs/notification-service.log" 2>&1 &
    NOTIFICATION_PID=$!
    echo $NOTIFICATION_PID > "$PROJECT_ROOT/logs/notification.pid"
    
    sleep 2
    if kill -0 $NOTIFICATION_PID 2>/dev/null; then
        success "Notification service started (PID: $NOTIFICATION_PID)"
    else
        error "Notification service failed to start"
        cat "$PROJECT_ROOT/logs/notification-service.log"
        exit 1
    fi
}

# Start scraper in background
start_scraper() {
    info "Starting scraper..."
    cd "$SCRAPER_DIR"
    
    # Kill existing scraper if running
    pkill -f "scraper_orchestrator.py" || true
    
    # Create logs directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/logs"
    
    # Start scraper in background (set PYTHONPATH to fix imports)
    export PYTHONPATH="$SCRAPER_DIR:$PYTHONPATH"
    nohup venv/bin/python src/scrapers/scraper_orchestrator.py > "$PROJECT_ROOT/logs/scraper.log" 2>&1 &
    SCRAPER_PID=$!
    echo $SCRAPER_PID > "$PROJECT_ROOT/logs/scraper.pid"
    
    sleep 2
    if kill -0 $SCRAPER_PID 2>/dev/null; then
        success "Scraper started (PID: $SCRAPER_PID)"
    else
        error "Scraper failed to start"
        cat "$PROJECT_ROOT/logs/scraper.log"
        exit 1
    fi
}

# Stop all services
stop_services() {
    info "Stopping all services..."
    
    # Stop background processes
    if [ -f "$PROJECT_ROOT/logs/notification.pid" ]; then
        NOTIFICATION_PID=$(cat "$PROJECT_ROOT/logs/notification.pid")
        kill $NOTIFICATION_PID 2>/dev/null || true
        rm -f "$PROJECT_ROOT/logs/notification.pid"
    fi
    
    if [ -f "$PROJECT_ROOT/logs/scraper.pid" ]; then
        SCRAPER_PID=$(cat "$PROJECT_ROOT/logs/scraper.pid")
        kill $SCRAPER_PID 2>/dev/null || true
        rm -f "$PROJECT_ROOT/logs/scraper.pid"
    fi
    
    # Stop Docker services
    cd "$PROJECT_ROOT"
    docker-compose down
    
    success "All services stopped"
}

# Show status
show_status() {
    info "Service Status:"
    
    # Check Docker services
    echo "Docker Services:"
    docker-compose ps
    
    echo ""
    echo "Background Services:"
    
    # Check notification service
    if [ -f "$PROJECT_ROOT/logs/notification.pid" ]; then
        NOTIFICATION_PID=$(cat "$PROJECT_ROOT/logs/notification.pid")
        if kill -0 $NOTIFICATION_PID 2>/dev/null; then
            echo "  ✅ Notification Service (PID: $NOTIFICATION_PID)"
        else
            echo "  ❌ Notification Service (not running)"
        fi
    else
        echo "  ❌ Notification Service (not started)"
    fi
    
    # Check scraper
    if [ -f "$PROJECT_ROOT/logs/scraper.pid" ]; then
        SCRAPER_PID=$(cat "$PROJECT_ROOT/logs/scraper.pid")
        if kill -0 $SCRAPER_PID 2>/dev/null; then
            echo "  ✅ Scraper (PID: $SCRAPER_PID)"
        else
            echo "  ❌ Scraper (not running)"
        fi
    else
        echo "  ❌ Scraper (not started)"
    fi
}

# Show logs
show_logs() {
    info "Recent logs:"
    echo ""
    echo "=== Notification Service Logs ==="
    tail -n 20 "$PROJECT_ROOT/logs/notification-service.log" 2>/dev/null || echo "No notification logs found"
    echo ""
    echo "=== Scraper Logs ==="
    tail -n 20 "$PROJECT_ROOT/logs/scraper.log" 2>/dev/null || echo "No scraper logs found"
}

# Main function
main() {
    case "${1:-start}" in
        start)
            info "🚀 Starting Tennis Booker local development environment..."
            check_prerequisites
            start_docker_services
            build_backend
            seed_database
            setup_scraper
            start_notification_service
            start_scraper
            echo ""
            success "🎾 Tennis Booker is running locally!"
            echo ""
            info "📋 Useful commands:"
            echo "  $0 status    - Check service status"
            echo "  $0 logs      - Show recent logs"
            echo "  $0 stop      - Stop all services"
            echo ""
            info "📁 Log files:"
            echo "  Notification: logs/notification-service.log"
            echo "  Scraper: logs/scraper.log"
            ;;
        stop)
            stop_services
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        restart)
            stop_services
            sleep 2
            main start
            ;;
        *)
            echo "Usage: $0 {start|stop|status|logs|restart}"
            echo ""
            echo "Commands:"
            echo "  start    - Start all services (default)"
            echo "  stop     - Stop all services"
            echo "  status   - Show service status"
            echo "  logs     - Show recent logs"
            echo "  restart  - Restart all services"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 