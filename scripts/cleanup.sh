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

# Clean up any PID files older than 1 day
echo "🔄 Cleaning old PID files..."
find logs/ -name "*.pid" -type f -mtime +1 -delete 2>/dev/null || true

# Clean up large log files (compress if > 10MB)
echo "📦 Compressing large log files..."
find logs/ -name "*.log" -size +10M -exec gzip {} \; 2>/dev/null || true

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