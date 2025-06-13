#!/bin/bash

# Cleanup script for tennis-booker project
# Run this script to clean up temporary files, logs, and build artifacts

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🧹 Starting tennis-booker cleanup..."

# Clean Python cache files (excluding venv)
echo "🐍 Cleaning Python cache files..."
find . -path "./venv" -prune -o -name "*.pyc" -delete -o -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true

# Clean old log files (keep recent ones)
echo "📝 Cleaning old log files..."
if [ -f "scraper_orchestrator.log" ]; then
    rm -f scraper_orchestrator.log
    echo "   ✅ Removed scraper_orchestrator.log"
fi

if [ -f "playwright_scraper.log" ]; then
    rm -f playwright_scraper.log
    echo "   ✅ Removed playwright_scraper.log"
fi

# Clean up old debug/test files
echo "🔧 Cleaning debug files..."
find . -name "debug_*" -type f -delete 2>/dev/null || true
find . -name "temp_*" -type f -delete 2>/dev/null || true
find . -name "test_temp*" -type f -delete 2>/dev/null || true
find . -name "test_auto_redis.py" -type f -delete 2>/dev/null || true

# Clean Go build artifacts
echo "🔨 Cleaning Go build artifacts..."
go clean
go mod tidy

# Clean backup files
echo "💾 Cleaning backup files..."
find . -name "*.bak" -type f -delete 2>/dev/null || true
find . -name "*.orig" -type f -delete 2>/dev/null || true
find . -name "*~" -type f -delete 2>/dev/null || true

# Clean up notification service binary if it exists in wrong location
if [ -f "notification-service" ]; then
    echo "🔧 Removing notification-service binary from root..."
    rm -f notification-service
fi

# Clean up PID files (all of them since they're only valid for current session)
echo "🔄 Cleaning PID files..."
if [ -d "logs" ]; then
    find logs/ -name "*.pid" -type f -delete 2>/dev/null || true
    echo "   ✅ Removed all PID files"
fi

# Clean up large log files (compress if > 10MB)
echo "📦 Compressing large log files..."
find logs/ -name "*.log" -size +10M -exec gzip {} \; 2>/dev/null || true

# Clean MongoDB data (Redis queue data, old slots, logs)
echo "🗄️ Cleaning MongoDB data..."
if command -v docker >/dev/null 2>&1 && docker ps | grep -q tennis-mongodb; then
    echo "   🔄 Clearing Redis queue data from MongoDB..."
    docker exec tennis-mongodb mongosh tennis_booking --eval "
        db.notification_queue.deleteMany({});
        print('✅ Cleared notification_queue collection');
        
        // Clean old slots (older than 7 days)
        const sevenDaysAgo = new Date();
        sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
        const oldSlotsResult = db.slots.deleteMany({scraped_at: {\$lt: sevenDaysAgo}});
        print('✅ Removed ' + oldSlotsResult.deletedCount + ' old slots');
        
        // Clean old scraping logs (older than 30 days)
        const thirtyDaysAgo = new Date();
        thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
        const oldLogsResult = db.scraping_logs.deleteMany({created_at: {\$lt: thirtyDaysAgo}});
        print('✅ Removed ' + oldLogsResult.deletedCount + ' old scraping logs');
    " 2>/dev/null || echo "   ⚠️ MongoDB cleanup failed (container may not be running)"
else
    echo "   ⚠️ MongoDB container not running, skipping database cleanup"
fi

# Clean Redis data
echo "🔴 Cleaning Redis data..."
if command -v docker >/dev/null 2>&1 && docker ps | grep -q tennis-redis; then
    echo "   🔄 Flushing Redis queues..."
    docker exec tennis-redis redis-cli FLUSHDB 2>/dev/null && echo "   ✅ Redis queues cleared" || echo "   ⚠️ Redis cleanup failed"
else
    echo "   ⚠️ Redis container not running, skipping Redis cleanup"
fi

# Optional: Clean Docker containers and images (commented out for safety)
# echo "🐳 Cleaning Docker artifacts..."
# docker system prune -f 2>/dev/null || true

echo "✅ Cleanup completed!"
echo ""
echo "📊 Current disk usage in project:"
du -sh . 2>/dev/null || echo "   (disk usage check failed)"

echo ""
echo "📁 Log files status:"
ls -lh logs/ 2>/dev/null || echo "   (no logs directory)" 