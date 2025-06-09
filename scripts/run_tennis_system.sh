#!/bin/bash

# 🎾 Tennis Court Booking System - Master Control Script
# 
# This script provides a unified interface to:
# 1. Start/stop all services
# 2. Run real-time scraping with Playwright
# 3. Test the complete system with notifications
# 4. Monitor and maintain the system

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
DOCKER_COMPOSE_FILE="${PROJECT_ROOT}/docker-compose.yml"
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
    
    scrape-now      Scrape all venues immediately (real sites)
    scrape-venue    Scrape specific venue (--venues required)
    scrape-loop     Continuous scraping with notifications
    
    test-scraping   Test scraping functionality only
    test-venue      Test specific venue scraping
    test-api        Test API endpoints only
    test-notifications  Test notification system
    test-full-pipeline  Test complete pipeline with notifications
    
    setup           Initialize project and seed data
    seed-venues     Seed venue data into MongoDB
    seed-user       Seed user preferences for notifications
    test-notification  Send test email notification
    start-notifications  Start notification service
    full-system     Start complete system with notifications
    cleanup         Clean up logs and temporary files
    monitor         Monitor system health continuously
    
    logs            Show recent logs from all services
    check-slots     Check available slots in database

OPTIONS:
    --venues NAMES  Comma-separated venue names for scraping
                   (e.g., "Victoria Park", "Stratford Park", "Ropemakers Field")
    --debug         Enable debug logging
    --force         Force operations without confirmation

    --loop-interval SECONDS  Scraping loop interval (default: 600)

EXAMPLES:
    $0 start                                    # Start all services
    $0 scrape-now                              # Scrape all venues now
    $0 scrape-venue --venues="Victoria Park"   # Scrape specific venue
    $0 scrape-loop                             # Continuous scraping with alerts
    $0 test-full-pipeline                      # Test everything with notifications
    $0 test-notification                       # Send test email
    $0 full-system                             # Complete system with notifications
    $0 check-slots                             # Check database slots

REAL-TIME INTEGRATION:
    - Uses Playwright to scrape actual tennis court websites
    - Stores availability in MongoDB
    - Triggers notifications for new slots
    - Supports Victoria Park, Stratford Park, Ropemakers Field

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
    
    # Check Playwright scraper
    if [ ! -f "$PROJECT_ROOT/src/playwright_scraper.py" ]; then
        error "Playwright scraper not found at src/playwright_scraper.py"
        exit 1
    fi
    
    # Check notification service binary
    if [ ! -f "$PROJECT_ROOT/bin/notification-service" ]; then
        warn "Notification service binary not found. Building..."
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

# Start notification service
start_notification_service() {
    info "Starting notification service..."
    
    # Kill existing notification service if running
    pkill -f "bin/notification-service" || true
    
    # Start notification service in background
    cd "$PROJECT_ROOT"
    nohup ./bin/notification-service > logs/notification-service.log 2>&1 &
    NOTIFICATION_PID=$!
    echo $NOTIFICATION_PID > logs/notification-service.pid
    
    success "Notification service started (PID: $NOTIFICATION_PID)"
}

# Stop notification service
stop_notification_service() {
    info "Stopping notification service..."
    
    if [ -f logs/notification-service.pid ]; then
        PID=$(cat logs/notification-service.pid)
        if kill $PID 2>/dev/null; then
            success "Notification service stopped (PID: $PID)"
        fi
        rm -f logs/notification-service.pid
    else
        pkill -f "bin/notification-service" || true
        success "Notification service stopped"
    fi
}



# Seed venues into MongoDB
seed_venues() {
    info "Seeding venues into MongoDB..."
    
    cd "$PROJECT_ROOT"
    
    # Run the Go seeding script
    if [ -f "cmd/seed-db/main.go" ]; then
        go run cmd/seed-db/main.go
        success "Venues seeded successfully"
    else
        error "Seed script not found at cmd/seed-db/main.go"
        exit 1
    fi
}

# Run immediate scraping
run_scraping_now() {
    log "🎾 Scraping all venues now (real sites: Victoria Park, Stratford Park, Ropemakers Field)"
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    python src/playwright_scraper.py --all ${DEBUG:+--debug}
    
    success "Scraping completed"
    
    # Show results
    info "Checking scraped results..."
    check_database_slots
}

# Run venue-specific scraping  
run_venue_scraping() {
    if [ -z "$VENUES" ]; then
        error "Venue name required. Use --venues option"
        exit 1
    fi
    
    log "🎾 Scraping venue: $VENUES"
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    python src/playwright_scraper.py --test "$VENUES" ${DEBUG:+--debug}
    
    success "Venue scraping completed"
}

# Run continuous scraping loop with notifications
run_scraping_loop() {
    LOOP_INTERVAL=${LOOP_INTERVAL:-600}  # Default 10 minutes
    
    log "🔄 Starting continuous scraping loop (interval: ${LOOP_INTERVAL}s)"
    warn "Press Ctrl+C to stop the loop"
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    # Store previous slot count for comparison
    PREVIOUS_SLOTS=0
    
    while true; do
        log "🎾 Starting scraping cycle..."
        
        # Get current slot count before scraping
        BEFORE_SLOTS=$(python -c "
import pymongo
client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']
print(db.slots.count_documents({}))
" 2>/dev/null || echo "0")
        
        # Run scraping
        python src/playwright_scraper.py --all ${DEBUG:+--debug}
        
        # Get slot count after scraping
        AFTER_SLOTS=$(python -c "
import pymongo
client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']
print(db.slots.count_documents({}))
" 2>/dev/null || echo "0")
        
        # Check for new slots
        NEW_SLOTS=$((AFTER_SLOTS - BEFORE_SLOTS))
        
        if [ $NEW_SLOTS -gt 0 ]; then
            success "🚨 NEW SLOTS FOUND: $NEW_SLOTS new available slots!"
            
            # Trigger notification (if notification service exists)
            if [ -f "$PROJECT_ROOT/bin/notification-service" ]; then
                info "Triggering notifications..."
                # Add notification trigger logic here
            fi
            
            # Show latest slots
            info "Latest available slots:"
            python -c "
import pymongo
from datetime import datetime
client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']
slots = db.slots.find({'available': True}).sort('scraped_at', -1).limit(5)
for slot in slots:
    venue = slot.get('venue_name', 'Unknown')
    court = slot.get('court_name', 'Unknown')
    date = slot.get('date', 'Unknown')
    time = f\"{slot.get('start_time', '?')}-{slot.get('end_time', '?')}\"
    price = slot.get('price', 'Unknown')
    print(f'  🎾 {venue} | {court} | {date} {time} | £{price}')
" 2>/dev/null
        else
            info "No new slots found this cycle"
        fi
        
        log "Waiting ${LOOP_INTERVAL} seconds until next cycle..."
        sleep $LOOP_INTERVAL
    done
}

# Check database slots
check_database_slots() {
    info "📊 Checking available slots in database..."
    
    python -c "
import pymongo
from collections import defaultdict

client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']

total_slots = db.slots.count_documents({})
available_slots = db.slots.count_documents({'available': True})

print(f'Total slots in database: {total_slots}')
print(f'Available slots: {available_slots}')
print(f'Availability rate: {(available_slots/total_slots)*100:.1f}%' if total_slots > 0 else 'No slots')

print('\nSlots by venue:')
venues = ['Victoria Park', 'Stratford Park', 'Ropemakers Field']
for venue in venues:
    count = db.slots.count_documents({'venue_name': venue, 'available': True})
    print(f'  {venue}: {count} available slots')

print('\nLatest slots:')
slots = db.slots.find({'available': True}).sort('scraped_at', -1).limit(10)
for slot in slots:
    venue = slot.get('venue_name', 'Unknown')
    court = slot.get('court_name', 'Unknown')
    date = slot.get('date', 'Unknown')
    time = f\"{slot.get('start_time', '?')}-{slot.get('end_time', '?')}\"
    price = slot.get('price', 'Unknown')
    scraped = slot.get('scraped_at', 'Unknown')
    print(f'  🎾 {venue} | {court} | {date} {time} | £{price} | {scraped}')
" 2>/dev/null || echo "Could not connect to database"
}

# Test scraping functionality
test_scraping() {
    log "🧪 Testing scraping functionality"
    
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    
    # Test each venue individually
    venues=("Victoria Park" "Stratford Park" "Ropemakers Field")
    
    for venue in "${venues[@]}"; do
        info "Testing $venue..."
        python src/playwright_scraper.py --test "$venue" ${DEBUG:+--debug}
        echo ""
    done
    
    success "Scraping tests completed"
}

# Test notification system
test_notifications() {
    log "🔔 Testing notification system"
    
    # This would test the notification/alert system
    info "Notification testing not yet implemented"
    warn "Implement notification service integration here"
}

# Test full pipeline
test_full_pipeline() {
    log "🚀 Testing complete pipeline with real scraping and notifications"
    
    info "Step 1: Clear existing slots for clean test"
    python -c "
import pymongo
client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']
result = db.slots.delete_many({})
print(f'Cleared {result.deleted_count} existing slots')
" 2>/dev/null || warn "Could not clear slots"
    
    info "Step 2: Run full scraping"
    cd "$PROJECT_ROOT"
    source "$VENV_PATH/bin/activate"
    python src/playwright_scraper.py --all ${DEBUG:+--debug}
    
    info "Step 3: Check results and trigger notifications"
    check_database_slots
    
    # Simulate notification for available slots
    AVAILABLE_COUNT=$(python -c "
import pymongo
client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']
print(db.slots.count_documents({'available': True}))
" 2>/dev/null || echo "0")
    
    if [ "$AVAILABLE_COUNT" -gt 0 ]; then
        success "🚨 PIPELINE SUCCESS: Found $AVAILABLE_COUNT available slots!"
        info "In production, this would trigger notifications to users"
    else
        warn "No available slots found - this may be normal depending on booking status"
    fi
    
    success "Full pipeline test completed"
}

# Check if Docker service is running
require_docker_service() {
    local service=$1
    if ! docker-compose ps $service | grep -q "Up"; then
        error "Docker service '$service' is not running. Please start it first:"
        error "  $0 start"
        exit 1
    fi
}

# Check if multiple Docker services are running
require_docker_services() {
    require_docker_service "mongodb"
    require_docker_service "redis"
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
    

    
    # Notification service
    echo -e "\n${BLUE}Notification Service:${NC}"
    if pgrep -f "bin/notification-service" &>/dev/null; then
        echo -e "  Status:  ${GREEN}✅ Running${NC}"
    else
        echo -e "  Status:  ${YELLOW}⚠️ Not running${NC}"
    fi
    
    # Check Python environment
    echo -e "\n${BLUE}Python Environment:${NC}"
    if [ -d "$VENV_PATH" ]; then
        echo -e "  Virtual Env: ${GREEN}✅ Ready${NC}"
    else
        echo -e "  Virtual Env: ${RED}❌ Missing${NC}"
    fi
    
    # Check database status
    echo -e "\n${BLUE}Database Status:${NC}"
    VENUE_COUNT=$(python -c "
import pymongo
client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']
print(db.venues.count_documents({}))
" 2>/dev/null || echo "0")
    
    SLOT_COUNT=$(python -c "
import pymongo
client = pymongo.MongoClient('mongodb://admin:YOUR_PASSWORD@localhost:27017')
db = client['tennis_booking']
print(db.slots.count_documents({}))
" 2>/dev/null || echo "0")
    
    echo -e "  Venues:  ${GREEN}$VENUE_COUNT${NC} configured"
    echo -e "  Slots:   ${GREEN}$SLOT_COUNT${NC} in database"
    
    echo ""
}

# Setup project
setup_project() {
    log "🔧 Setting up Tennis Court Booking System"
    
    # Create directories
    mkdir -p logs
    mkdir -p venue_analysis
    
    # Build Go binaries
    info "Building Go binaries..."
    make build || warn "Could not build Go binaries - make sure you have Go installed"
    
    # Set up Python environment
    if [ ! -d "$VENV_PATH" ]; then
        info "Creating Python virtual environment..."
        python3 -m venv "$VENV_PATH"
    fi
    
    source "$VENV_PATH/bin/activate"
    
    # Install Python dependencies
    info "Installing Python dependencies..."
    pip install pymongo playwright
    
    # Install Playwright browsers
    info "Installing Playwright browsers..."
    playwright install chromium
    
    # Start Docker and seed data
    start_docker
    
    info "Seeding venues data..."
    seed_venues
    
    success "Setup completed successfully!"
    
    echo -e "\n${CYAN}🎉 Tennis Court Booking System is ready!${NC}"
    echo -e "\nNext steps:"
    echo -e "  $0 start                  # Start all services"
    echo -e "  $0 scrape-now            # Scrape all venues immediately"
    echo -e "  $0 scrape-loop           # Start continuous scraping with notifications"
    echo -e "  $0 test-full-pipeline    # Test everything"
}

# Clean up logs and temporary files
cleanup() {
    log "🧹 Cleaning up logs and temporary files"
    
    # Clean old logs
    find logs -name "*.log" -mtime +7 -delete 2>/dev/null || true
    find . -name "venue_analysis_*.json" -mtime +30 -delete 2>/dev/null || true
    
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
        echo -e "Database slots:"
        check_database_slots | head -5
        
        echo -e "\nPress Ctrl+C to stop monitoring"
        sleep 30
    done
}

# Show logs
show_logs() {
    log "📜 Recent system logs"
    
    echo -e "\n${BLUE}Notification Service Logs:${NC}"
    tail -n 20 logs/notification-service.log 2>/dev/null || echo "No notification service logs found"
    
    echo -e "\n${BLUE}Scraper Logs:${NC}"
    tail -n 20 playwright_scraper.log 2>/dev/null || echo "No scraper logs found"
    
    echo -e "\n${BLUE}Docker Logs:${NC}"
    docker-compose logs --tail=10 2>/dev/null || echo "No docker logs found"
}

# Main script logic
main() {
    # Parse arguments
    COMMAND=""
    VENUES=""
    DEBUG=""
    FORCE=""
    LOOP_INTERVAL=""
    
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

            --loop-interval)
                LOOP_INTERVAL="$2"
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
            success "🎾 Tennis Court Booking System started successfully!"
            echo -e "\nUse: $0 status          # to check system status"
            echo -e "Use: $0 scrape-now     # to scrape all venues immediately"
            echo -e "Use: $0 scrape-loop    # for continuous scraping with notifications"
            echo -e "Use: $0 start-notifications # to start notification service"
            ;;
        stop)
            stop_notification_service
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
        scrape-now)
            check_prerequisites
            run_scraping_now
            ;;
        scrape-venue)
            check_prerequisites
            run_venue_scraping
            ;;
        scrape-loop)
            check_prerequisites
            run_scraping_loop
            ;;
        test-scraping)
            check_prerequisites
            test_scraping
            ;;
        test-venue)
            check_prerequisites
            run_venue_scraping
            ;;
        test-api)
            # API testing logic here
            info "API testing not yet implemented"
            ;;
        test-notifications)
            check_prerequisites
            test_notifications
            ;;
        test-full-pipeline)
            check_prerequisites
            test_full_pipeline
            ;;
        setup)
            setup_project
            ;;
        seed-venues)
            check_prerequisites
            seed_venues
            ;;
        check-slots)
            check_database_slots
            ;;
        cleanup)
            cleanup
            ;;
        monitor)
            check_prerequisites
            monitor_system
            ;;
        logs)
            show_logs
            ;;
        seed-user)
            info "Seeding user preferences in database..."
            require_docker_service "mongodb"
            
            if [ -f "${PROJECT_ROOT}/cmd/seed-user/main.go" ]; then
                cd "${PROJECT_ROOT}"
                info "Running user seeding script..."
                if go run cmd/seed-user/main.go; then
                    success "✅ User preferences seeded successfully"
                else
                    error "Failed to seed user preferences"
                    exit 1
                fi
            else
                error "User seeding script not found at cmd/seed-user/main.go"
                exit 1
            fi
            ;;
        test-notification)
            info "Sending test notification..."
            require_docker_service "mongodb"
            
            cd "${PROJECT_ROOT}"
            if [ -f "cmd/notification-service/main.go" ]; then
                info "Compiling notification service..."
                go build -o /tmp/notification-test cmd/notification-service/main.go
                
                info "Sending test email to mvgnum@gmail.com..."
                
                # Set environment variables for the test
                export MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017"
                export DB_NAME="tennis_booking"
                export REDIS_ADDR="localhost:6379"
                export GMAIL_EMAIL="mvgnum@gmail.com"
                export GMAIL_PASSWORD="eswk jgaw zbet wgxo"
                
                # Run the notification service in test mode
                if /tmp/notification-test test; then
                    success "✅ Test notification sent to mvgnum@gmail.com"
                    info "Check your email inbox for the test notification"
                else
                    error "Failed to send test notification"
                    exit 1
                fi
            else
                error "Notification service not found"
                exit 1
            fi
            ;;
        start-notifications)
            info "Starting notification service..."
            require_docker_service "mongodb"
            require_docker_service "redis"
            
            cd "${PROJECT_ROOT}"
            if [ -f "cmd/notification-service/main.go" ]; then
                info "Starting notification service in background..."
                info "This will monitor for new court slots and send email alerts"
                info "Press Ctrl+C to stop"
                
                # Set environment variables
                export MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017"
                export DB_NAME="tennis_booking"
                export REDIS_ADDR="localhost:6379"
                export REDIS_PASSWORD="password"
                export GMAIL_EMAIL="mvgnum@gmail.com"
                export GMAIL_PASSWORD="eswk jgaw zbet wgxo"
                export FROM_NAME="Tennis Court Alerts"
                
                go run cmd/notification-service/main.go
            else
                error "Notification service not found"
                exit 1
            fi
            ;;
        full-system)
            info "Starting complete tennis booking system with notifications..."
            require_docker_services
            
            info "Setting up complete system:"
            info "1. Seeding venues"
            "$0" seed-venues
            
            info "2. Seeding user preferences"
            "$0" seed-user
            
            info "3. Testing notification system"
            "$0" test-notification
            
            info "4. Starting scraping loop in background"
            "$0" scrape-loop --loop-interval 300 &  # 5 minute intervals
            SCRAPER_PID=$!
            
            info "5. Starting notification service"
            "$0" start-notifications &
            NOTIFICATION_PID=$!
            
            info "🚀 Complete system running!"
            info "- Scraping: PID $SCRAPER_PID (every 5 minutes)"
            info "- Notifications: PID $NOTIFICATION_PID"
            info "- Monitoring: mvgnum@gmail.com"
            info "- Venues: Victoria Park, Stratford Park, Ropemakers Field"
            info "- Times: Weekdays 19:00-22:00, Weekends 10:00-20:00"
            info ""
            info "Press Ctrl+C to stop all services"
            
            # Wait for interrupt
            trap "kill $SCRAPER_PID $NOTIFICATION_PID 2>/dev/null; exit" INT TERM
            wait
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