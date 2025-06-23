#!/bin/bash

# 🧹 Tennis Booker Cleanup Script
# Comprehensive cleanup of temporary files, logs, and build artifacts

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
success() { echo -e "${GREEN}✅ $1${NC}"; }
warn() { echo -e "${YELLOW}⚠️  $1${NC}"; }

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Parse arguments
DEEP_CLEAN=false
KEEP_LOGS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --deep)
            DEEP_CLEAN=true
            shift
            ;;
        --keep-logs)
            KEEP_LOGS=true
            shift
            ;;
        -h|--help)
            echo "🧹 Tennis Booker Cleanup Script"
            echo ""
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --deep       Deep clean (includes node_modules, venv, etc.)"
            echo "  --keep-logs  Keep log files"
            echo "  -h, --help   Show this help"
            exit 0
            ;;
        *)
            warn "Unknown option: $1"
            shift
            ;;
    esac
done

info "🧹 Starting Tennis Booker cleanup..."

# Clean Python cache files
info "🐍 Cleaning Python cache files..."
find . -path "./*/venv" -prune -o -name "*.pyc" -delete -o -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
find . -name "*.pyo" -delete 2>/dev/null || true
find . -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
success "Python cache files cleaned"

# Clean JavaScript/Node files
info "📦 Cleaning JavaScript build files..."
find . -name ".next" -exec rm -rf {} + 2>/dev/null || true
find . -name "dist" -path "*/apps/frontend/dist" -exec rm -rf {} + 2>/dev/null || true
success "JavaScript build files cleaned"

# Clean Go build files
info "🐹 Cleaning Go build files..."
find . -path "*/apps/backend/bin" -exec rm -rf {} + 2>/dev/null || true
if [ -d "apps/backend" ]; then
    cd apps/backend && go clean && cd "$PROJECT_ROOT"
fi
success "Go build files cleaned"

# Clean log files (unless --keep-logs)
if [ "$KEEP_LOGS" = false ]; then
    info "📝 Cleaning log files..."
    rm -rf logs/*.log 2>/dev/null || true
    rm -rf logs/*.pid 2>/dev/null || true
    rm -f scraper_orchestrator.log playwright_scraper.log 2>/dev/null || true
    success "Log files cleaned"
else
    info "📝 Keeping log files (--keep-logs specified)"
fi

# Clean temporary and debug files
info "🔧 Cleaning temporary files..."
find . -name "debug_*" -type f -delete 2>/dev/null || true
find . -name "temp_*" -type f -delete 2>/dev/null || true
find . -name "*.tmp" -delete 2>/dev/null || true
find . -name ".DS_Store" -delete 2>/dev/null || true
find . -name "*.bak" -type f -delete 2>/dev/null || true
find . -name "*.orig" -type f -delete 2>/dev/null || true
find . -name "*~" -type f -delete 2>/dev/null || true
success "Temporary files cleaned"

# Clean PID files
info "🔄 Cleaning PID files..."
if [ -d "logs" ]; then
    find logs/ -name "*.pid" -type f -delete 2>/dev/null || true
fi
success "PID files cleaned"

# Deep clean option
if [ "$DEEP_CLEAN" = true ]; then
    info "🔥 Deep cleaning (removing dependencies)..."
    
    # Remove node_modules
    find . -name "node_modules" -exec rm -rf {} + 2>/dev/null || true
    
    # Remove Python virtual environments
    find . -name "venv" -path "*/apps/scraper/venv" -exec rm -rf {} + 2>/dev/null || true
    
    success "Deep clean completed"
fi

# Clean Docker resources (if containers are not running)
info "🐳 Cleaning Docker resources..."
if command -v docker >/dev/null 2>&1; then
    # Clean stopped containers
    docker container prune -f 2>/dev/null || true
    
    # Clean unused volumes
    docker volume prune -f 2>/dev/null || true
    
    success "Docker resources cleaned"
else
    warn "Docker not available, skipping Docker cleanup"
fi

# Show cleanup summary
info "📊 Cleanup Summary:"
echo "  ✅ Python cache files cleaned"
echo "  ✅ JavaScript build files cleaned"
echo "  ✅ Go build files cleaned"
if [ "$KEEP_LOGS" = false ]; then
    echo "  ✅ Log files cleaned"
else
    echo "  ⏭️ Log files kept"
fi
echo "  ✅ Temporary files cleaned"
echo "  ✅ PID files cleaned"
if [ "$DEEP_CLEAN" = true ]; then
    echo "  ✅ Dependencies removed (deep clean)"
fi
echo "  ✅ Docker resources cleaned"

success "🎾 Tennis Booker cleanup completed!"

# Show next steps
info "💡 Next steps:"
if [ "$DEEP_CLEAN" = true ]; then
    echo "  Run: ./scripts/dev.sh setup    # Reinstall dependencies"
fi
echo "  Run: ./scripts/dev.sh start    # Start development environment"