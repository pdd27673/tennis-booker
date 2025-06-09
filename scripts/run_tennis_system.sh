#!/bin/bash

# 🎾 Tennis Court Booking System - Master Control Script
# 
# This script provides a unified interface to:
# 1. Start/stop all services
# 2. Run cost-effective scraping
# 3. Test the complete system
# 4. Monitor and maintain the system

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DOCKER_COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.yml"
API_PORT=8080
REDIS_PORT=6379
MONGO_PORT=27017
VENV_PATH="${PROJECT_ROOT}/scraper-env"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
🎾 Tennis Court Booking System - Master Control Script

USAGE:
    $0 [COMMAND] [OPTIONS]

COMMANDS:
    start           Start all services (Docker + API + Scheduler)
    stop            Stop all services
    restart         Restart all services
    status          Check status of all services
    
    scrape-daily    Run daily lightweight scraping (no Firecrawl credits)
    scrape-full     Run full 7-day scraping  
    scrape-analyze  Use Firecrawl to analyze booking sites (costs credits)
    
    test-system     Run comprehensive system tests
    test-scraping   Test scraping functionality only
    test-api        Test API endpoints only
    
    setup           Initialize project and seed data
    cleanup         Clean up logs and temporary files
    monitor         Monitor system health continuously
    
    openapi         Generate OpenAPI specification
    logs            Show recent logs from all services

OPTIONS:
    --venues NAMES  Comma-separated venue names for scraping
    --debug         Enable debug logging
    --force         Force operations without confirmation
    --port PORT     Override API port (default: 8080)

EXAMPLES:
    $0 start                        # Start all services
    $0 scrape-daily                 # Daily scraping (no credits used)
    $0 scrape-full --venues=victoria,stratford
    $0 test-system --debug          # Full system test with debug
    $0 monitor                      # Continuous health monitoring

COST-EFFECTIVE STRATEGY:
    - Use 'scrape-daily' for everyday monitoring (no Firecrawl costs)
    - Use 'scrape-analyze' ONCE to understand sites (minimal credits)
    - Use 'scrape-full' for comprehensive 7-day scanning
    - Firecrawl fallback only when lightweight scraping fails

EOF
}

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        error "Docker is required but not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is required but not installed"
        exit 1
    fi
    
    # Check Python virtual environment
    if [ ! -d "$VENV_PATH" ]; then
        warn "Python virtual environment not found at $VENV_PATH"
        warn "Run: $0 setup"
        exit 1
    fi
    
    # Check Go binary
    if [ ! -f "$PROJECT_ROOT/bin/api" ]; then
        warn "API binary not found. Building..."
        cd "$PROJECT_ROOT"
        make build
    fi
    
    success "Prerequisites check passed"
}

# Start Docker services
start_docker() {
    info "Starting Docker services (MongoDB, Redis)..."
    
    cd "$PROJECT_ROOT"
    docker-compose up -d mongodb redis
    
    # Wait for services to be ready
    info "Waiting for services to be ready..."
    
    # Wait for MongoDB
    for i in {1..30}; do
        if docker-compose exec -T mongodb mongosh --eval "db.runCommand('ping')" &>/dev/null; then
            success "MongoDB is ready"
            break
        fi
        echo -n "."
        sleep 2
    done
    
    # Wait for Redis
    for i in {1..30}; do
        if docker-compose exec -T redis redis-cli ping &>/dev/null; then
            success "Redis is ready"
            break
        fi
        echo -n "."
        sleep 2
    done
}

# Stop Docker services
stop_docker() {
    info "Stopping Docker services..."
    cd "$PROJECT_ROOT"
    docker-compose down
    success "Docker services stopped"
}

# Start API server
start_api() {
    info "Starting API server on port $API_PORT..."
    
    # Kill existing API if running
    pkill -f "bin/api" || true
    
    # Start API in background
    cd "$PROJECT_ROOT"
    nohup ./bin/api > logs/api.log 2>&1 &
    API_PID=$!
    echo $API_PID > logs/api.pid
    
    # Wait for API to be ready
    for i in {1..20}; do
        if curl -s "http://localhost:$API_PORT/api/health" &>/dev/null; then
            success "API server is ready (PID: $API_PID)"
            return 0
        fi
        echo -n "."
        sleep 2
    done
    
    error "API server failed to start"
    return 1
}

# Stop API server
stop_api() {
    info "Stopping API server..."
    
    if [ -f logs/api.pid ]; then
        PID=$(cat logs/api.pid)
        if kill $PID 2>/dev/null; then
            success "API server stopped (PID: $PID)"
        fi
        rm -f logs/api.pid
    else
        pkill -f "bin/api" || true
        success "API server stopped"
    fi
}

# Start scheduler
start_scheduler() {
    info "Starting scheduler service..."
    
    # Kill existing scheduler if running
    pkill -f "bin/scheduler" || true
    
    # Start scheduler in background
    cd "$PROJECT_ROOT"
    nohup ./bin/scheduler > logs/scheduler.log 2>&1 &
    SCHEDULER_PID=$!
    echo $SCHEDULER_PID > logs/scheduler.pid
    
    success "Scheduler started (PID: $SCHEDULER_PID)"
}

# Stop scheduler
stop_scheduler() {
    info "Stopping scheduler service..."
    
    if [ -f logs/scheduler.pid ]; then
        PID=$(cat logs/scheduler.pid)
        if kill $PID 2>/dev/null; then
            success "Scheduler stopped (PID: $PID)"
        fi
        rm -f logs/scheduler.pid
    else
        pkill -f "bin/scheduler" || true
        success "Scheduler stopped"
    fi
}

# Check service status
check_status() {
    echo -e "\n${CYAN}🎾 Tennis Court Booking System Status${NC}\n"
    
    # Docker services
    echo -e "${BLUE}Docker Services:${NC}"
    if docker-compose ps | grep -q "Up"; then
        echo -e "  MongoDB: ${GREEN}✅ Running${NC}"
        echo -e "  Redis:   ${GREEN}✅ Running${NC}"
    else
        echo -e "  Docker:  ${RED}❌ Not running${NC}"
    fi
    
    # API server
    echo -e "\n${BLUE}API Server:${NC}"
    if curl -s "http://localhost:$API_PORT/api/health" &>/dev/null; then
        echo -e "  Status:  ${GREEN}✅ Healthy${NC}"
        echo -e "  URL:     http://localhost:$API_PORT"
    else
        echo -e "  Status:  ${RED}❌ Not responding${NC}"
    fi
    
    # Scheduler
    echo -e "\n${BLUE}Scheduler:${NC}"
    if pgrep -f "bin/scheduler" &>/dev/null; then
        echo -e "  Status:  ${GREEN}✅ Running${NC}"
    else
        echo -e "  Status:  ${RED}❌ Not running${NC}"
    fi
    
    # Notification service
    echo -e "\n${BLUE}Notification Service:${NC}"
    if pgrep -f "bin/notification-service" &>/dev/null; then
        echo -e "  Status:  ${GREEN}✅ Running${NC}"
    else
        echo -e "  Status:  ${YELLOW}⚠️ Not running${NC}"
    fi
    
    echo ""
}

# Run daily scraping
run_daily_scraping() {
    log "🌅 Running daily lightweight scraping (no Firecrawl credits)"
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    python src/tennis_court_scraper.py \
        --mode=daily \
        ${VENUES:+--venues="$VENUES"} \
        ${DEBUG:+--log-level=DEBUG}
    
    success "Daily scraping completed"
}

# Run full 7-day scraping
run_full_scraping() {
    log "🚀 Running full 7-day scraping"
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    python src/tennis_court_scraper.py \
        --mode=full \
        ${VENUES:+--venues="$VENUES"} \
        ${DEBUG:+--log-level=DEBUG}
    
    success "Full scraping completed"
}

# Run Firecrawl analysis
run_analysis() {
    warn "🔥 Running Firecrawl analysis (will use credits!)"
    
    if [ -z "$FORCE" ]; then
        read -p "This will use Firecrawl credits. Continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Analysis cancelled"
            return 0
        fi
    fi
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    python src/tennis_court_scraper.py \
        --mode=analyze \
        ${VENUES:+--venues="$VENUES"} \
        ${DEBUG:+--log-level=DEBUG}
    
    success "Analysis completed"
}

# Run system tests
run_system_tests() {
    log "🧪 Running comprehensive system tests"
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    # Test 1: API endpoints
    info "Testing API endpoints..."
    python -c "
import requests
import sys

base = 'http://localhost:$API_PORT'
endpoints = [
    '/api/health',
    '/api/metrics', 
    '/api/v1/venues',
    '/api/v1/alerts/history?user_id=507f1f77bcf86cd799439011'
]

for endpoint in endpoints:
    try:
        resp = requests.get(f'{base}{endpoint}', timeout=5)
        if resp.status_code == 200:
            print(f'✅ {endpoint}')
        else:
            print(f'❌ {endpoint} ({resp.status_code})')
    except Exception as e:
        print(f'❌ {endpoint} (error: {e})')
"
    
    # Test 2: Scraper functionality
    info "Testing scraper functionality..."
    python src/tennis_court_scraper.py --mode=test ${DEBUG:+--log-level=DEBUG}
    
    success "System tests completed"
}

# Setup project
setup_project() {
    log "🔧 Setting up Tennis Court Booking System"
    
    # Create directories
    mkdir -p logs
    mkdir -p venue_analysis
    
    # Build Go binaries
    info "Building Go binaries..."
    make build
    
    # Set up Python environment
    if [ ! -d "$VENV_PATH" ]; then
        info "Creating Python virtual environment..."
        python3 -m venv "$VENV_PATH"
    fi
    
    source "$VENV_PATH/bin/activate"
    
    # Install Python dependencies
    info "Installing Python dependencies..."
    pip install -r utils/requirements.txt
    
    # Start Docker and seed data
    start_docker
    
    info "Seeding venues data..."
    cd "$PROJECT_ROOT"
    ./scripts/seed_venues.sh
    
    success "Setup completed successfully!"
    
    echo -e "\n${CYAN}🎉 Tennis Court Booking System is ready!${NC}"
    echo -e "\nNext steps:"
    echo -e "  $0 start              # Start all services"
    echo -e "  $0 scrape-daily       # Run daily scraping"
    echo -e "  $0 status             # Check system status"
}

# Clean up logs and temporary files
cleanup() {
    log "🧹 Cleaning up logs and temporary files"
    
    # Clean old logs
    find logs -name "*.log" -mtime +7 -delete 2>/dev/null || true
    find . -name "venue_analysis_*.json" -mtime +30 -delete 2>/dev/null || true
    find src -name "tennis_scraper.log" -mtime +7 -delete 2>/dev/null || true
    
    # Clean Python cache
    find . -name "__pycache__" -type d -exec rm -rf {} + 2>/dev/null || true
    find . -name "*.pyc" -delete 2>/dev/null || true
    
    success "Cleanup completed"
}

# Monitor system health
monitor_system() {
    log "📊 Starting continuous system monitoring (Ctrl+C to stop)"
    
    while true; do
        clear
        echo -e "${CYAN}🎾 Tennis Court Booking System Monitor${NC}"
        echo -e "$(date)"
        echo -e "========================================\n"
        
        check_status
        
        echo -e "\n${BLUE}Recent Activity:${NC}"
        echo -e "API Requests (last minute):"
        tail -n 20 logs/api.log 2>/dev/null | grep "$(date '+%Y/%m/%d %H:%M')" | wc -l | sed 's/^/  /'
        
        echo -e "\nPress Ctrl+C to stop monitoring"
        sleep 30
    done
}

# Generate OpenAPI specification
generate_openapi() {
    log "📋 Generating OpenAPI specification"
    
    cd "$PROJECT_ROOT"
    
    # Create OpenAPI spec
    cat > api-spec.yaml << 'EOF'
openapi: 3.0.3
info:
  title: Tennis Court Booking API
  description: Cost-effective tennis court notification and booking system
  version: 1.0.0
  contact:
    name: Tennis Court Scraper
    
servers:
  - url: http://localhost:8080
    description: Development server

paths:
  /api/health:
    get:
      summary: Health check
      responses:
        '200':
          description: System is healthy
          
  /api/metrics:
    get:
      summary: System metrics
      responses:
        '200':
          description: System metrics data
          
  /api/v1/venues:
    get:
      summary: Get all venues
      parameters:
        - name: provider
          in: query
          schema:
            type: string
            enum: [courtsides, lta_clubspark]
      responses:
        '200':
          description: List of venues
          
  /api/v1/courts/available:
    get:
      summary: Get available court slots
      parameters:
        - name: venue_ids
          in: query
          schema:
            type: string
        - name: date_from
          in: query
          schema:
            type: string
            format: date
        - name: date_to
          in: query
          schema:
            type: string
            format: date
      responses:
        '200':
          description: Available court slots
          
  /api/v1/preferences:
    get:
      summary: Get user preferences
      parameters:
        - name: user_id
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: User preferences
    put:
      summary: Update user preferences
      parameters:
        - name: user_id
          in: query
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Preferences updated
          
  /api/v1/alerts/history:
    get:
      summary: Get alert history
      parameters:
        - name: user_id
          in: query
          required: true
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
            default: 50
        - name: offset
          in: query
          schema:
            type: integer
            default: 0
      responses:
        '200':
          description: Alert history
EOF
    
    success "OpenAPI specification generated: api-spec.yaml"
}

# Show logs
show_logs() {
    log "📜 Recent system logs"
    
    echo -e "\n${BLUE}API Logs:${NC}"
    tail -n 20 logs/api.log 2>/dev/null || echo "No API logs found"
    
    echo -e "\n${BLUE}Scheduler Logs:${NC}"
    tail -n 20 logs/scheduler.log 2>/dev/null || echo "No scheduler logs found"
    
    echo -e "\n${BLUE}Scraper Logs:${NC}"
    tail -n 20 src/tennis_scraper.log 2>/dev/null || echo "No scraper logs found"
}

# Main script logic
main() {
    # Parse arguments
    COMMAND=""
    VENUES=""
    DEBUG=""
    FORCE=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --venues)
                VENUES="$2"
                shift 2
                ;;
            --debug)
                DEBUG="true"
                shift
                ;;
            --force)
                FORCE="true"
                shift
                ;;
            --port)
                API_PORT="$2"
                shift 2
                ;;
            --help)
                show_help
                exit 0
                ;;
            -*)
                error "Unknown option: $1"
                show_help
                exit 1
                ;;
            *)
                if [ -z "$COMMAND" ]; then
                    COMMAND="$1"
                fi
                shift
                ;;
        esac
    done
    
    # Show help if no command
    if [ -z "$COMMAND" ]; then
        show_help
        exit 0
    fi
    
    # Create logs directory
    mkdir -p logs
    
    # Execute command
    case $COMMAND in
        start)
            check_prerequisites
            start_docker
            start_api
            start_scheduler
            success "🎾 Tennis Court Booking System started successfully!"
            echo -e "\nUse: $0 status    # to check system status"
            echo -e "Use: $0 scrape-daily  # to run daily scraping"
            ;;
        stop)
            stop_scheduler
            stop_api
            stop_docker
            success "Tennis Court Booking System stopped"
            ;;
        restart)
            $0 stop
            sleep 2
            $0 start
            ;;
        status)
            check_status
            ;;
        scrape-daily)
            check_prerequisites
            run_daily_scraping
            ;;
        scrape-full)
            check_prerequisites
            run_full_scraping
            ;;
        scrape-analyze)
            check_prerequisites
            run_analysis
            ;;
        test-system)
            check_prerequisites
            run_system_tests
            ;;
        test-scraping)
            check_prerequisites
            cd "$PROJECT_ROOT"
            source "$VENV_PATH/bin/activate"
            python src/tennis_court_scraper.py --mode=test ${DEBUG:+--log-level=DEBUG}
            ;;
        test-api)
            run_system_tests
            ;;
        setup)
            setup_project
            ;;
        cleanup)
            cleanup
            ;;
        monitor)
            check_prerequisites
            monitor_system
            ;;
        openapi)
            generate_openapi
            ;;
        logs)
            show_logs
            ;;
        *)
            error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@" 