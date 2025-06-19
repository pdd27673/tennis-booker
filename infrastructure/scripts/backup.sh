#!/bin/bash

# Tennis Booker Backup Script
# Creates comprehensive backups of all data and configurations

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

# Create backup directory
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

log "Creating backup in $BACKUP_DIR"

# Backup MongoDB
backup_mongodb() {
    log "Backing up MongoDB..."
    
    if docker ps | grep -q tennis-mongodb; then
        # Create dump inside container
        docker exec tennis-mongodb mongodump \
            --authenticationDatabase admin \
            -u "$MONGO_ROOT_USERNAME" \
            -p "$MONGO_ROOT_PASSWORD" \
            --out /tmp/backup
        
        # Copy dump to host
        docker cp tennis-mongodb:/tmp/backup "$BACKUP_DIR/mongodb"
        
        # Cleanup container
        docker exec tennis-mongodb rm -rf /tmp/backup
        
        # Compress backup
        tar -czf "$BACKUP_DIR/mongodb.tar.gz" -C "$BACKUP_DIR" mongodb
        rm -rf "$BACKUP_DIR/mongodb"
        
        success "MongoDB backup completed"
    else
        warning "MongoDB container not running, skipping backup"
    fi
}

# Backup Redis
backup_redis() {
    log "Backing up Redis..."
    
    if docker ps | grep -q tennis-redis; then
        # Create RDB dump
        docker exec tennis-redis redis-cli --rdb /tmp/dump.rdb
        
        # Copy dump to host
        docker cp tennis-redis:/tmp/dump.rdb "$BACKUP_DIR/redis_dump.rdb"
        
        # Cleanup container
        docker exec tennis-redis rm -f /tmp/dump.rdb
        
        # Compress backup
        gzip "$BACKUP_DIR/redis_dump.rdb"
        
        success "Redis backup completed"
    else
        warning "Redis container not running, skipping backup"
    fi
}

# Backup Vault data
backup_vault() {
    log "Backing up Vault data..."
    
    if docker ps | grep -q tennis-vault; then
        # Copy vault data
        docker cp tennis-vault:/vault/data "$BACKUP_DIR/vault"
        
        # Compress backup
        tar -czf "$BACKUP_DIR/vault.tar.gz" -C "$BACKUP_DIR" vault
        rm -rf "$BACKUP_DIR/vault"
        
        success "Vault backup completed"
    else
        warning "Vault container not running, skipping backup"
    fi
}

# Backup Traefik certificates
backup_traefik() {
    log "Backing up Traefik certificates..."
    
    if docker ps | grep -q tennis-traefik; then
        # Copy certificates
        docker cp tennis-traefik:/letsencrypt "$BACKUP_DIR/traefik_certs"
        
        # Compress backup
        tar -czf "$BACKUP_DIR/traefik_certs.tar.gz" -C "$BACKUP_DIR" traefik_certs
        rm -rf "$BACKUP_DIR/traefik_certs"
        
        success "Traefik certificates backup completed"
    else
        warning "Traefik container not running, skipping backup"
    fi
}

# Backup configuration files
backup_configs() {
    log "Backing up configuration files..."
    
    # Create config backup directory
    mkdir -p "$BACKUP_DIR/configs"
    
    # Copy important config files
    cp .env.prod "$BACKUP_DIR/configs/" 2>/dev/null || warning "No .env.prod file found"
    cp docker-compose.prod.yml "$BACKUP_DIR/configs/" 2>/dev/null || warning "No docker-compose.prod.yml file found"
    cp -r infrastructure/ "$BACKUP_DIR/configs/" 2>/dev/null || warning "No infrastructure directory found"
    
    # Compress configs
    tar -czf "$BACKUP_DIR/configs.tar.gz" -C "$BACKUP_DIR" configs
    rm -rf "$BACKUP_DIR/configs"
    
    success "Configuration backup completed"
}

# Create backup manifest
create_manifest() {
    log "Creating backup manifest..."
    
    cat > "$BACKUP_DIR/manifest.txt" << EOF
Tennis Booker Backup Manifest
=============================
Backup Date: $(date)
Backup Directory: $BACKUP_DIR
Domain: $DOMAIN_NAME

Contents:
$(ls -la "$BACKUP_DIR")

Backup Sizes:
$(du -h "$BACKUP_DIR"/* 2>/dev/null || echo "No backup files found")

Total Size: $(du -sh "$BACKUP_DIR" | cut -f1)

Restore Instructions:
1. Stop all services: docker-compose -f docker-compose.prod.yml down
2. Extract backups to appropriate locations
3. Restart services: docker-compose -f docker-compose.prod.yml up -d

MongoDB Restore:
docker exec -i tennis-mongodb mongorestore --authenticationDatabase admin -u $MONGO_ROOT_USERNAME -p $MONGO_ROOT_PASSWORD --drop /tmp/restore/
docker cp mongodb.tar.gz tennis-mongodb:/tmp/
docker exec tennis-mongodb tar -xzf /tmp/mongodb.tar.gz -C /tmp/

Redis Restore:
docker cp redis_dump.rdb.gz tennis-redis:/tmp/
docker exec tennis-redis gunzip /tmp/redis_dump.rdb.gz
docker exec tennis-redis redis-cli --rdb /tmp/redis_dump.rdb

Vault Restore:
docker cp vault.tar.gz tennis-vault:/tmp/
docker exec tennis-vault tar -xzf /tmp/vault.tar.gz -C /vault/

Traefik Restore:
docker cp traefik_certs.tar.gz tennis-traefik:/tmp/
docker exec tennis-traefik tar -xzf /tmp/traefik_certs.tar.gz -C /
EOF
    
    success "Backup manifest created"
}

# Cleanup old backups
cleanup_old_backups() {
    log "Cleaning up old backups..."
    
    local retention_days=${BACKUP_RETENTION_DAYS:-30}
    
    # Remove backups older than retention period
    find backups/ -type d -name "20*" -mtime +$retention_days -exec rm -rf {} \; 2>/dev/null || true
    
    success "Old backups cleaned up (retention: $retention_days days)"
}

# Upload to S3 (if configured)
upload_to_s3() {
    if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$S3_BUCKET" ]; then
        log "Uploading backup to S3..."
        
        # Create archive
        tar -czf "$BACKUP_DIR.tar.gz" -C backups "$(basename "$BACKUP_DIR")"
        
        # Upload to S3 (requires aws cli)
        if command -v aws >/dev/null 2>&1; then
            aws s3 cp "$BACKUP_DIR.tar.gz" "s3://$S3_BUCKET/tennis-booker-backups/"
            rm "$BACKUP_DIR.tar.gz"
            success "Backup uploaded to S3"
        else
            warning "AWS CLI not found, skipping S3 upload"
        fi
    fi
}

# Show backup information
show_info() {
    log "Backup complete! 📦"
    echo
    echo "=== Backup Information ==="
    echo "Location: $BACKUP_DIR"
    echo "Size: $(du -sh "$BACKUP_DIR" | cut -f1)"
    echo "Files:"
    ls -la "$BACKUP_DIR"
    echo
    echo "=== Restore Command ==="
    echo "./infrastructure/scripts/restore.sh $BACKUP_DIR"
    echo
}

# Main backup flow
main() {
    log "Starting Tennis Booker backup..."
    
    backup_mongodb
    backup_redis
    backup_vault
    backup_traefik
    backup_configs
    create_manifest
    cleanup_old_backups
    upload_to_s3
    show_info
    
    success "Backup completed successfully!"
}

# Run main function
main "$@" 