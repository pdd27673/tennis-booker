#!/bin/bash

# 🎾 Tennis Booker - Local Development Script
# Starts MongoDB, Redis, Vault, notification service, test-auth-server, and frontend for local development

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

# Default environment variables for local development
# Load environment variables from .env file if it exists
if [ -f ".env" ]; then
    echo "📄 Loading environment variables from .env file..."
    set -a  # automatically export all variables
    source .env
    set +a  # disable automatic export
else
    echo "⚠️  No .env file found. Please copy .env-example to .env and configure your settings."
    echo "   You can run: cp .env-example .env"
    echo ""
fi

# Set default values for required variables if not already set
export MONGO_ROOT_USERNAME="${MONGO_ROOT_USERNAME:-admin}"
export MONGO_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD:-}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-}"
export MONGO_URI="${MONGO_URI:-mongodb://${MONGO_ROOT_USERNAME}:${MONGO_ROOT_PASSWORD}@localhost:27017/tennis_booking?authSource=admin}"
export REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"
export DB_NAME="${DB_NAME:-tennis_booking}"

# Vault configuration
export VAULT_ADDR="${VAULT_ADDR:-http://localhost:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-dev-token}"
export VAULT_DEV_ROOT_TOKEN_ID="${VAULT_DEV_ROOT_TOKEN_ID:-dev-token}"

# Email configuration for notifications (loaded from .env)
export GMAIL_EMAIL="${GMAIL_EMAIL:-}"
export GMAIL_PASSWORD="${GMAIL_PASSWORD:-}"

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
    
    if ! command -v npm &> /dev/null; then
        error "npm is required but not installed"
        exit 1
    fi
    
    success "Prerequisites check passed"
}

# Start Docker services (MongoDB, Redis, and Vault)
start_docker_services() {
    info "Starting Docker services (MongoDB, Redis, Vault)..."
    cd "$PROJECT_ROOT"
    
    # Create a docker-compose.override.yml file to add Vault if it doesn't exist in the original
    if ! grep -q "vault:" docker-compose.yml; then
        cat > docker-compose.override.yml << EOF
version: '3'

services:
  vault:
    image: vault:1.12.0
    container_name: tennis-vault
    ports:
      - "8200:8200"
    environment:
      - VAULT_DEV_ROOT_TOKEN_ID=dev-token
      - VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200
    cap_add:
      - IPC_LOCK
    command: server -dev -dev-root-token-id=dev-token
    restart: unless-stopped
EOF
        info "Created docker-compose.override.yml with Vault configuration"
    fi
    
    # Start MongoDB, Redis, and Vault
    docker-compose up -d mongodb redis vault
    
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
        if [ $i -eq 30 ]; then
            error "MongoDB failed to start in time"
            exit 1
        fi
    done
    
    # Wait for Redis
    for i in {1..30}; do
        if docker exec tennis-redis redis-cli -a password ping &>/dev/null; then
            success "Redis is ready"
            break
        fi
        echo -n "."
        sleep 2
        if [ $i -eq 30 ]; then
            error "Redis failed to start in time"
            exit 1
        fi
    done
    
    # Wait for Vault
    for i in {1..30}; do
        if curl -s -o /dev/null -w "%{http_code}" http://localhost:8200/v1/sys/health | grep -q "^200"; then
            success "Vault is ready"
            
            # Setup Vault secrets for development
            echo "🔐 Initializing Vault secrets..."
            curl -s -X POST -H "X-Vault-Token: ${VAULT_TOKEN:-dev-token}" \
                -d '{"data":{"secret":"super-secret-jwt-key-for-local-development"}}' \
                http://localhost:8200/v1/secret/data/tennisapp/prod/jwt > /dev/null
                
            success "Vault initialized with JWT secret"
            break
        fi
        echo -n "."
        sleep 2
        if [ $i -eq 30 ]; then
            error "Vault failed to start in time"
            exit 1
        fi
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
    export USER_EMAIL="${USER_EMAIL:-mvgnum@gmail.com}"
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

# Start integrated backend server
start_backend_server() {
    info "Starting integrated backend server..."
    cd "$BACKEND_DIR"
    
    # Kill existing backend server if running
    pkill -f "bin/server" || true
    
    # Create logs directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/logs"
    
    # Build the server first
    make build-server
    
    # Start integrated backend server in background
    nohup ./bin/server > "$PROJECT_ROOT/logs/backend-server.log" 2>&1 &
    BACKEND_SERVER_PID=$!
    echo $BACKEND_SERVER_PID > "$PROJECT_ROOT/logs/backend-server.pid"
    
    sleep 3
    if kill -0 $BACKEND_SERVER_PID 2>/dev/null; then
        success "Backend server started (PID: $BACKEND_SERVER_PID)"
    else
        error "Backend server failed to start"
        cat "$PROJECT_ROOT/logs/backend-server.log"
        exit 1
    fi
    
    # Check if the server is responding
    for i in {1..10}; do
        if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health | grep -q "^200"; then
            success "Backend server is responding"
            break
        fi
        echo -n "."
        sleep 2
        if [ $i -eq 10 ]; then
            warn "Backend server is not responding to health check (this could be normal during startup)"
            tail -n 10 "$PROJECT_ROOT/logs/backend-server.log"
        fi
    done
}

# Start frontend in background
start_frontend() {
    info "Starting frontend..."
    cd "$FRONTEND_DIR"
    
    # Create .env.local if it doesn't exist
    if [ ! -f ".env.local" ]; then
        cat > .env.local << EOF
VITE_API_URL=http://localhost:8080
VITE_MOCK_API_ENABLED=false
EOF
        success "Created .env.local with API configuration"
    fi
    
    # Kill existing frontend process if running
    pkill -f "vite.*$FRONTEND_DIR" || true
    
    # Create logs directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/logs"
    
    # Start frontend in background
    nohup npm run dev > "$PROJECT_ROOT/logs/frontend.log" 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > "$PROJECT_ROOT/logs/frontend.pid"
    
    sleep 5
    if kill -0 $FRONTEND_PID 2>/dev/null; then
        success "Frontend started (PID: $FRONTEND_PID)"
    else
        error "Frontend failed to start"
        cat "$PROJECT_ROOT/logs/frontend.log"
        exit 1
    fi
    
    # Extract the frontend URL from the logs
    FRONTEND_URL=$(grep -o 'http://[a-zA-Z0-9.:]*' "$PROJECT_ROOT/logs/frontend.log" | head -n 1)
    if [ ! -z "$FRONTEND_URL" ]; then
        success "Frontend available at $FRONTEND_URL"
    else
        warn "Frontend started but URL could not be detected"
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

# Start scraper scheduler in background
start_scraper() {
    info "Starting scraper scheduler..."
    cd "$SCRAPER_DIR"
    
    # Kill existing scraper processes if running
    pkill -f "scheduler.py" || true
    pkill -f "scraper_orchestrator.py" || true
    
    # Create logs directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/logs"
    
    # Set environment variables for scraper
    export MONGO_URI="${MONGO_URI}"
    export REDIS_HOST="localhost"
    export REDIS_PORT="6379"
    export REDIS_PASSWORD="${REDIS_PASSWORD}"
    export SCRAPER_INTERVAL_MINUTES="5"  # 5 minutes for local development
    export SCRAPER_DAYS_AHEAD="8"        # 8 days ahead as requested
    export LOG_LEVEL="INFO"
    
    # Start scraper scheduler in background (set PYTHONPATH to fix imports)
    export PYTHONPATH="$SCRAPER_DIR:$PYTHONPATH"
    nohup venv/bin/python src/scheduler.py > "$PROJECT_ROOT/logs/scraper.log" 2>&1 &
    SCRAPER_PID=$!
    echo $SCRAPER_PID > "$PROJECT_ROOT/logs/scraper.pid"
    
    sleep 3
    if kill -0 $SCRAPER_PID 2>/dev/null; then
        success "Scraper scheduler started (PID: $SCRAPER_PID, interval: 5 minutes)"
    else
        error "Scraper scheduler failed to start"
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
    
    # Also kill any remaining scraper processes
    pkill -f "scheduler.py" || true
    pkill -f "scraper_orchestrator.py" || true
    
    if [ -f "$PROJECT_ROOT/logs/backend-server.pid" ]; then
        BACKEND_SERVER_PID=$(cat "$PROJECT_ROOT/logs/backend-server.pid")
        kill $BACKEND_SERVER_PID 2>/dev/null || true
        rm -f "$PROJECT_ROOT/logs/backend-server.pid"
    fi
    
    if [ -f "$PROJECT_ROOT/logs/frontend.pid" ]; then
        FRONTEND_PID=$(cat "$PROJECT_ROOT/logs/frontend.pid")
        kill $FRONTEND_PID 2>/dev/null || true
        rm -f "$PROJECT_ROOT/logs/frontend.pid"
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
    
    # Check scraper scheduler
    if [ -f "$PROJECT_ROOT/logs/scraper.pid" ]; then
        SCRAPER_PID=$(cat "$PROJECT_ROOT/logs/scraper.pid")
        if kill -0 $SCRAPER_PID 2>/dev/null; then
            echo "  ✅ Scraper Scheduler (PID: $SCRAPER_PID)"
        else
            echo "  ❌ Scraper Scheduler (not running)"
        fi
    else
        echo "  ❌ Scraper Scheduler (not started)"
    fi
    
    # Check backend server
    if [ -f "$PROJECT_ROOT/logs/backend-server.pid" ]; then
        BACKEND_SERVER_PID=$(cat "$PROJECT_ROOT/logs/backend-server.pid")
        if kill -0 $BACKEND_SERVER_PID 2>/dev/null; then
            echo "  ✅ Backend Server (PID: $BACKEND_SERVER_PID)"
        else
            echo "  ❌ Backend Server (not running)"
        fi
    else
        echo "  ❌ Backend Server (not started)"
    fi
    
    # Check frontend
    if [ -f "$PROJECT_ROOT/logs/frontend.pid" ]; then
        FRONTEND_PID=$(cat "$PROJECT_ROOT/logs/frontend.pid")
        if kill -0 $FRONTEND_PID 2>/dev/null; then
            FRONTEND_URL=$(grep -o 'http://[a-zA-Z0-9.:]*' "$PROJECT_ROOT/logs/frontend.log" | head -n 1)
            echo "  ✅ Frontend (PID: $FRONTEND_PID, URL: $FRONTEND_URL)"
        else
            echo "  ❌ Frontend (not running)"
        fi
    else
        echo "  ❌ Frontend (not started)"
    fi
    
    # Check API health
    echo ""
    echo "API Health:"
    HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/health 2>/dev/null || echo "failed")
    if [ "$HEALTH_STATUS" == "200" ]; then
        echo "  ✅ API Health: OK"
    else
        echo "  ❌ API Health: Not responding (status: $HEALTH_STATUS)"
    fi
}

# Show logs
show_logs() {
    info "Recent logs:"
    
    echo ""
    echo "=== Backend Server Logs ==="
    tail -n 20 "$PROJECT_ROOT/logs/backend-server.log" 2>/dev/null || echo "No backend server logs found"
    
    echo ""
    echo "=== Frontend Logs ==="
    tail -n 20 "$PROJECT_ROOT/logs/frontend.log" 2>/dev/null || echo "No frontend logs found"
    
    echo ""
    echo "=== Notification Service Logs ==="
    tail -n 20 "$PROJECT_ROOT/logs/notification-service.log" 2>/dev/null || echo "No notification logs found"
    
    echo ""
    echo "=== Scraper Scheduler Logs ==="
    tail -n 20 "$PROJECT_ROOT/logs/scraper.log" 2>/dev/null || echo "No scraper scheduler logs found"
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
            start_backend_server
            start_notification_service
            start_scraper
            start_frontend
            echo ""
            success "🎾 Tennis Booker is running locally!"
            echo ""
            show_status
            echo ""
            info "📋 Useful commands:"
            echo "  $0 status    - Check service status"
            echo "  $0 logs      - Show recent logs"
            echo "  $0 stop      - Stop all services"
            echo ""
            info "📁 Log files:"
            echo "  Backend Server: logs/backend-server.log"
            echo "  Frontend: logs/frontend.log"
            echo "  Notification: logs/notification-service.log"
            echo "  Scraper Scheduler: logs/scraper.log"
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