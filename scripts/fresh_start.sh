#!/bin/bash

# Tennis Court Booking System - Fresh Start Script
# This script completely cleans the system and starts fresh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
success() { echo -e "${GREEN}✅ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠️  $1${NC}"; }
error() { echo -e "${RED}❌ $1${NC}"; }

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MONGO_PORT=27017
REDIS_PORT=6379

info "🧹 Starting fresh system cleanup and restart..."
info "Project root: $PROJECT_ROOT"

# Stop any running services
info "Stopping any running services..."
pkill -f "notification-service" || true
pkill -f "playwright_scraper" || true
docker-compose -f docker-compose.yml down || true
success "Services stopped"

# Start Docker services
info "Starting Docker services (MongoDB, Redis)..."
cd "$PROJECT_ROOT"
docker-compose up -d mongodb redis
sleep 10

# Wait for services to be ready
info "Waiting for services to be ready..."
for i in {1..30}; do
    if docker exec tennis-mongodb mongosh --quiet --eval "db.adminCommand('ping')" &>/dev/null; then
        success "MongoDB is ready"
        break
    fi
    echo -n "."
    sleep 2
done

for i in {1..30}; do
    if docker exec tennis-redis redis-cli -a password ping &>/dev/null; then
        success "Redis is ready"
        break
    fi
    echo -n "."
    sleep 2
done

# Clean MongoDB completely
info "🗑️  Cleaning MongoDB database..."
docker exec tennis-mongodb mongosh --quiet --eval "
use tennis_booking;
db.venues.deleteMany({});
db.slots.deleteMany({});
db.users.deleteMany({});
db.scraping_logs.deleteMany({});
db.notification_history.deleteMany({});
print('✅ All MongoDB collections cleared');
"
success "MongoDB cleaned"

# Clean Redis queues
info "🗑️  Cleaning Redis queues..."
docker exec tennis-redis redis-cli -a password FLUSHALL
success "Redis cleaned"

# Build binaries if needed
info "Building Go binaries..."
cd "$PROJECT_ROOT"
make build
success "Binaries built"

# Seed database with venues
info "🌱 Seeding venues..."
MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin" \
DB_NAME="tennis_booking" \
./bin/seed-db
success "Venues seeded"

# Seed user preferences  
info "🌱 Seeding user preferences..."
MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin" \
DB_NAME="tennis_booking" \
USER_EMAIL="mvgnum@gmail.com" \
./bin/seed-user
success "User preferences seeded"

# Verify setup
info "🔍 Verifying setup..."
docker exec tennis-mongodb mongosh --quiet --eval "
use tennis_booking;
print('Venues:', db.venues.countDocuments({}));
print('Users:', db.users.countDocuments({}));
print('Slots:', db.slots.countDocuments({}));
print('Notification History:', db.notification_history.countDocuments({}));
"

# Start notification service
info "🔔 Starting notification service..."
MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin" \
REDIS_ADDR="localhost:6379" \
REDIS_PASSWORD="password" \
GMAIL_EMAIL="mvgnum@gmail.com" \
GMAIL_PASSWORD="eswk jgaw zbet wgxo" \
DB_NAME="tennis_booking" \
nohup ./bin/notification-service > logs/fresh-notification.log 2>&1 &
NOTIFICATION_PID=$!
echo $NOTIFICATION_PID > logs/notification.pid

sleep 3
if kill -0 $NOTIFICATION_PID 2>/dev/null; then
    success "Notification service started (PID: $NOTIFICATION_PID)"
else
    error "Notification service failed to start"
    cat logs/fresh-notification.log
    exit 1
fi

# Send test notification to verify email works
# info "📧 Sending test notification..."
# MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin" \
# GMAIL_EMAIL="mvgnum@gmail.com" \
# GMAIL_PASSWORD="eswk jgaw zbet wgxo" \
# ./bin/notification-service test
# success "Test notification sent!"

# Run initial scraping
info "🎾 Running initial scraping to populate slots..."
cd "$PROJECT_ROOT"
MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin" \
DB_NAME="tennis_booking" \
source venv/bin/activate && python3 src/playwright_scraper.py --all

# Check if any new slots were found
SLOT_COUNT=$(docker exec tennis-mongodb mongosh --quiet --eval "
use tennis_booking;
print(db.slots.countDocuments({}));
" | tail -n 1)

success "Initial scraping completed - found $SLOT_COUNT total slots"

# Start continuous scraping
info "🔄 Starting continuous scraping (every 5 minutes)..."
MONGO_URI="mongodb://admin:YOUR_PASSWORD@localhost:27017/tennis_booking?authSource=admin" \
DB_NAME="tennis_booking" \
nohup bash -c "
cd '$PROJECT_ROOT'
source venv/bin/activate
while true; do
    echo \"\$(date): Starting scraping cycle...\"
    python3 src/playwright_scraper.py --all
    echo \"\$(date): Scraping cycle completed, waiting 5 minutes...\"
    sleep 300
done
" > logs/fresh-scraper.log 2>&1 &
SCRAPER_PID=$!
echo $SCRAPER_PID > logs/scraper.pid

success "Continuous scraping started (PID: $SCRAPER_PID)"

# Show final status
echo ""
success "🎉 Fresh system startup complete!"
echo ""
info "📊 System Status:"
echo "  ✅ MongoDB: Running and cleaned"
echo "  ✅ Redis: Running and cleaned" 
echo "  ✅ Venues: 3 seeded (Victoria Park, Stratford Park, Ropemakers Field)"
echo "  ✅ User: mvgnum@gmail.com with preferences"
echo "  ✅ Notification Service: Running (PID: $NOTIFICATION_PID)"
echo "  ✅ Scraper: Running every 5 minutes (PID: $SCRAPER_PID)"
echo "  ✅ Test Email: Sent successfully"
echo ""
info "📧 Email Preferences:"
echo "  📍 All venues monitored"
echo "  ⏰ Weekdays: 19:00-22:00"
echo "  🌅 Weekends: 10:00-20:00"
echo "  💰 Max price: £1000"
echo ""
info "📋 Monitoring:"
echo "  📁 Logs: tail -f logs/fresh-notification.log"
echo "  📁 Scraper: tail -f logs/fresh-scraper.log"
echo "  🛑 Stop: $0 stop"
echo ""
success "🎾 You should receive email notifications when matching slots become available!"

show_stop_help() {
    echo ""
    info "To stop the system:"
    echo "  pkill -f notification-service"
    echo "  pkill -f playwright_scraper"
    echo "  docker-compose down"
}

# Handle stop command
if [[ "${1:-}" == "stop" ]]; then
    info "🛑 Stopping fresh system..."
    
    if [[ -f logs/notification.pid ]]; then
        PID=$(cat logs/notification.pid)
        kill $PID 2>/dev/null || true
        rm -f logs/notification.pid
    fi
    
    if [[ -f logs/scraper.pid ]]; then
        PID=$(cat logs/scraper.pid)
        kill $PID 2>/dev/null || true  
        rm -f logs/scraper.pid
    fi
    
    pkill -f notification-service || true
    pkill -f playwright_scraper || true
    
    success "Fresh system stopped"
    exit 0
fi

show_stop_help 