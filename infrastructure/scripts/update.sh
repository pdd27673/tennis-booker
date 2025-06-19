#!/bin/bash

# Tennis Booker Update Script
# Updates the production deployment with latest changes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check if .env.prod exists
if [ ! -f ".env.prod" ]; then
    error "Production environment file .env.prod not found"
fi

# Load environment variables
source .env.prod

# Backup current deployment
backup_deployment() {
    log "Creating backup before update..."
    
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # Backup database
    docker exec tennis-mongodb mongodump --authenticationDatabase admin -u "$MONGO_ROOT_USERNAME" -p "$MONGO_ROOT_PASSWORD" --out "/tmp/backup"
    docker cp tennis-mongodb:/tmp/backup "$backup_dir/mongodb"
    
    # Backup Redis
    docker exec tennis-redis redis-cli --rdb /tmp/dump.rdb
    docker cp tennis-redis:/tmp/dump.rdb "$backup_dir/redis_dump.rdb"
    
    # Backup Vault data
    docker cp tennis-vault:/vault/data "$backup_dir/vault"
    
    # Backup Traefik certificates
    docker cp tennis-traefik:/letsencrypt "$backup_dir/traefik_certs"
    
    success "Backup created in $backup_dir"
}

# Pull latest code
update_code() {
    log "Pulling latest code..."
    
    if [ -d ".git" ]; then
        git pull origin main
        success "Code updated from Git"
    else
        warning "Not a Git repository. Please manually update code."
    fi
}

# Update services
update_services() {
    log "Updating services..."
    
    # Pull latest base images
    docker-compose -f docker-compose.prod.yml pull
    
    # Rebuild custom images
    docker-compose -f docker-compose.prod.yml build --no-cache
    
    # Restart services with zero downtime
    docker-compose -f docker-compose.prod.yml up -d --force-recreate
    
    success "Services updated"
}

# Wait for services to be healthy
wait_for_services() {
    log "Waiting for services to be healthy..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose -f docker-compose.prod.yml ps | grep -q "unhealthy"; then
            log "Attempt $attempt/$max_attempts: Some services are still starting..."
            sleep 10
            ((attempt++))
        else
            success "All services are healthy"
            return 0
        fi
    done
    
    error "Services failed to become healthy within timeout"
}

# Cleanup old images
cleanup() {
    log "Cleaning up old Docker images..."
    
    # Remove dangling images
    docker image prune -f
    
    # Remove old images (keep last 3 versions)
    docker images --format "table {{.Repository}}:{{.Tag}}\t{{.CreatedAt}}" | grep tennis | sort -k2 -r | tail -n +4 | awk '{print $1}' | xargs -r docker rmi
    
    success "Cleanup completed"
}

# Verify update
verify_update() {
    log "Verifying update..."
    
    # Check service status
    docker-compose -f docker-compose.prod.yml ps
    
    # Test endpoints
    if curl -s -f "https://$DOMAIN_NAME" >/dev/null; then
        success "Frontend is accessible"
    else
        warning "Frontend may not be accessible"
    fi
    
    if curl -s -f "https://api.$DOMAIN_NAME/health" >/dev/null; then
        success "API is accessible"
    else
        warning "API may not be accessible"
    fi
    
    success "Update verification complete"
}

# Show update information
show_info() {
    log "Update complete! 🎉"
    echo
    echo "=== Service Status ==="
    docker-compose -f docker-compose.prod.yml ps
    echo
    echo "=== Useful Commands ==="
    echo "View logs: docker-compose -f docker-compose.prod.yml logs -f [service]"
    echo "Rollback: ./infrastructure/scripts/rollback.sh"
    echo "Monitor: docker stats"
    echo
}

# Main update flow
main() {
    log "Starting Tennis Booker update..."
    
    backup_deployment
    update_code
    update_services
    wait_for_services
    cleanup
    verify_update
    show_info
    
    success "Update completed successfully!"
}

# Run main function
main "$@" 