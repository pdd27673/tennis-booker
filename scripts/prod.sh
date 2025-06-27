#!/bin/bash

# 🎾 Tennis Booker - Production Management Script
# Unified script for production server operations on OCI host

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

# Configuration - Adjust paths for production environment
PROD_DIR="/home/opc"
COMPOSE_FILE="$PROD_DIR/docker-compose.yml"
ENV_FILE="$PROD_DIR/.env"
NGINX_CONFIG="/etc/nginx/conf.d/courtscout.conf"
SSL_CERT="/etc/ssl/certs/courtscout.run.place.crt"
SSL_KEY="/etc/ssl/private/courtscout.run.place.key"

show_help() {
    cat << EOF
🎾 Tennis Booker Production Management

Usage: $0 <command> [options]

CORE COMMANDS:
  start                 Start all production services
  stop                  Stop all services gracefully
  restart               Restart all services
  status                Show detailed service status
  logs [service]        Show logs (backend|scraper|notification|redis|mongodb|vault|all)

CONTAINER MANAGEMENT:
  ps                    List all containers with details
  pull                  Pull latest images
  cleanup               Remove stopped containers and unused images
  remove                Stop and remove all containers (DESTRUCTIVE)
  recreate [service]    Recreate specific service

ENVIRONMENT & CONFIG:
  env                   Show environment variables (safe)
  config                Show current configuration
  ssl                   Check SSL certificate status
  nginx                 Check nginx configuration and restart if needed

HEALTH & MONITORING:
  health                Check all service health
  disk                  Check disk usage
  memory                Check memory usage
  network               Check network connectivity

MAINTENANCE:
  backup                Backup critical data
  update                Update system packages
  cert-renew            Renew SSL certificates (if using certbot)

EXAMPLES:
  $0 start              # Start all production services
  $0 logs backend       # Show backend logs
  $0 health             # Check service health
  $0 restart backend    # Restart specific service
  $0 cleanup            # Clean up unused resources

EOF
}

# Check if running as root for certain operations
check_root() {
    if [[ $EUID -eq 0 ]]; then
        return 0
    else
        return 1
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Source environment variables safely
source_env() {
    if [[ -f "$ENV_FILE" ]]; then
        # Export variables while hiding sensitive values in output
        info "Loading environment variables from $ENV_FILE"
        set -a
        source "$ENV_FILE"
        set +a
        success "Environment variables loaded"
    else
        warn "Environment file not found: $ENV_FILE"
        return 1
    fi
}

# Start all services
start_services() {
    info "Starting Tennis Booker production services..."
    
    source_env
    
    if [[ ! -f "$COMPOSE_FILE" ]]; then
        error "Docker Compose file not found: $COMPOSE_FILE"
        return 1
    fi
    
    cd "$PROD_DIR"
    docker-compose -f "$COMPOSE_FILE" up -d
    
    # Wait a moment for services to start
    sleep 5
    
    show_status
    success "All services started"
}

# Stop all services gracefully
stop_services() {
    info "Stopping all services gracefully..."
    
    cd "$PROD_DIR"
    if [[ -f "$COMPOSE_FILE" ]]; then
        docker-compose -f "$COMPOSE_FILE" down --timeout 30
    else
        warn "Compose file not found, stopping containers individually..."
        docker stop $(docker ps -q --filter "name=tennis-") 2>/dev/null || true
    fi
    
    success "All services stopped"
}

# Restart services
restart_services() {
    local service="$1"
    
    if [[ -n "$service" ]]; then
        info "Restarting $service..."
        cd "$PROD_DIR"
        docker-compose -f "$COMPOSE_FILE" restart "$service"
    else
        info "Restarting all services..."
        stop_services
        sleep 3
        start_services
    fi
}

# Show detailed service status
show_status() {
    info "Tennis Booker Service Status:"
    echo ""
    
    # Docker containers
    echo "🐳 Docker Containers:"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}" --filter "name=tennis-"
    echo ""
    
    # Service health checks
    echo "🏥 Health Checks:"
    
    # Backend API
    local api_response=$(curl -s http://localhost:8080/api/health 2>/dev/null)
    if [[ -n "$api_response" ]]; then
        local api_status=$(echo "$api_response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
        if [[ "$api_status" == "healthy" ]]; then
            echo "✅ Backend API: Healthy"
        else
            echo "⚠️  Backend API: Responding but status '$api_status' (Vault not connected)"
        fi
    else
        echo "❌ Backend API: Not responding"
    fi
    
    # Nginx
    if systemctl is-active --quiet nginx 2>/dev/null; then
        echo "✅ Nginx: Running"
    else
        echo "❌ Nginx: Not running"
    fi
    
    # SSL Certificate
    if [[ -f "$SSL_CERT" ]]; then
        local expiry=$(openssl x509 -enddate -noout -in "$SSL_CERT" | cut -d= -f2)
        echo "🔒 SSL Certificate: Valid until $expiry"
    else
        echo "❌ SSL Certificate: Not found"
    fi
    
    echo ""
}

# Show logs
show_logs() {
    local service="${1:-all}"
    local lines="${2:-100}"
    
    cd "$PROD_DIR"
    
    case "$service" in
        backend)
            if [[ "$lines" == "100" ]]; then
                docker logs tennis-backend --tail="$lines" -f
            else
                docker logs tennis-backend --tail="$lines"
            fi
            ;;
        scraper)
            if [[ "$lines" == "100" ]]; then
                docker logs tennis-scraper --tail="$lines" -f
            else
                docker logs tennis-scraper --tail="$lines"
            fi
            ;;
        notification|notifications)
            if [[ "$lines" == "100" ]]; then
                docker logs tennis-notifications --tail="$lines" -f
            else
                docker logs tennis-notifications --tail="$lines"
            fi
            ;;
        redis)
            if [[ "$lines" == "100" ]]; then
                docker logs tennis-redis --tail="$lines" -f
            else
                docker logs tennis-redis --tail="$lines"
            fi
            ;;
        mongodb|mongo)
            if [[ "$lines" == "100" ]]; then
                docker logs tennis-mongodb --tail="$lines" -f
            else
                docker logs tennis-mongodb --tail="$lines"
            fi
            ;;
        vault)
            if [[ "$lines" == "100" ]]; then
                docker logs tennis-vault --tail="$lines" -f
            else
                docker logs tennis-vault --tail="$lines"
            fi
            ;;
        nginx)
            if check_root; then
                tail -"$lines" /var/log/nginx/access.log /var/log/nginx/error.log
            else
                sudo tail -"$lines" /var/log/nginx/access.log /var/log/nginx/error.log
            fi
            ;;
        all|*)
            echo "=== Backend ==="
            docker logs tennis-backend --tail="$((lines/6))"
            echo -e "\n=== Scraper ==="
            docker logs tennis-scraper --tail="$((lines/6))"
            echo -e "\n=== Notifications ==="
            docker logs tennis-notifications --tail="$((lines/6))"
            ;;
    esac
}

# List containers with details
list_containers() {
    info "All Tennis Booker containers:"
    docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Image}}\t{{.CreatedAt}}\t{{.Size}}" --filter "name=tennis-"
}

# Pull latest images
pull_images() {
    info "Pulling latest images..."
    cd "$PROD_DIR"
    docker-compose -f "$COMPOSE_FILE" pull
    success "Images updated"
}

# Cleanup unused resources
cleanup_resources() {
    info "Cleaning up unused Docker resources..."
    
    # Remove stopped containers
    docker container prune -f
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes (be careful with this)
    read -p "Remove unused volumes? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker volume prune -f
    fi
    
    # Remove unused networks
    docker network prune -f
    
    success "Cleanup completed"
}

# Remove all containers (DESTRUCTIVE)
remove_all() {
    warn "This will STOP and REMOVE all Tennis Booker containers!"
    read -p "Are you sure? Type 'YES' to continue: " confirmation
    
    if [[ "$confirmation" != "YES" ]]; then
        info "Operation cancelled"
        return 0
    fi
    
    info "Stopping and removing all containers..."
    cd "$PROD_DIR"
    docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans
    
    success "All containers removed"
}

# Recreate specific service
recreate_service() {
    local service="$1"
    
    if [[ -z "$service" ]]; then
        error "Service name required"
        echo "Available services: backend, scraper, notifications, redis, mongodb, vault"
        return 1
    fi
    
    info "Recreating $service..."
    cd "$PROD_DIR"
    docker-compose -f "$COMPOSE_FILE" up -d --force-recreate "tennis-$service"
    
    success "$service recreated"
}

# Show environment variables (safe)
show_env() {
    info "Environment Configuration (sensitive values hidden):"
    
    if [[ -f "$ENV_FILE" ]]; then
        # Show env vars but hide sensitive values
        grep -E "^[A-Z_]+" "$ENV_FILE" | sed 's/\(.*=\).*/\1***HIDDEN***/' | head -20
        echo "..."
        echo "Total variables: $(grep -c "^[A-Z_]" "$ENV_FILE" 2>/dev/null || echo "0")"
    else
        error "Environment file not found: $ENV_FILE"
    fi
}

# Show configuration
show_config() {
    info "Current Configuration:"
    echo "📁 Production Directory: $PROD_DIR"
    echo "🐳 Compose File: $COMPOSE_FILE"
    echo "🔧 Environment File: $ENV_FILE"
    echo "🌐 Nginx Config: $NGINX_CONFIG"
    echo "🔒 SSL Certificate: $SSL_CERT"
    echo ""
    
    show_env
}

# Check SSL status
check_ssl() {
    info "SSL Certificate Status:"
    
    if [[ -f "$SSL_CERT" ]]; then
        openssl x509 -in "$SSL_CERT" -text -noout | grep -E "(Subject:|Not After:)"
        echo ""
        
        # Check certificate validity
        if openssl x509 -checkend 604800 -noout -in "$SSL_CERT"; then
            success "Certificate is valid for more than 7 days"
        else
            warn "Certificate expires within 7 days!"
        fi
    else
        error "SSL certificate not found"
    fi
}

# Check and restart nginx if needed
check_nginx() {
    info "Checking Nginx configuration..."
    
    if check_root; then
        nginx -t
        if [[ $? -eq 0 ]]; then
            success "Nginx configuration is valid"
            systemctl reload nginx
            success "Nginx reloaded"
        else
            error "Nginx configuration has errors"
        fi
    else
        sudo nginx -t
        if [[ $? -eq 0 ]]; then
            success "Nginx configuration is valid"
            sudo systemctl reload nginx
            success "Nginx reloaded"
        else
            error "Nginx configuration has errors"
        fi
    fi
}

# Health check all services
health_check() {
    info "Performing comprehensive health check..."
    echo ""
    
    # Docker daemon
    if docker info >/dev/null 2>&1; then
        success "Docker daemon: Running"
    else
        error "Docker daemon: Not running"
    fi
    
    # Container health
    echo ""
    echo "Container Health:"
    docker ps --format "table {{.Names}}\t{{.Status}}" --filter "name=tennis-" | while read line; do
        echo "$line"
    done
    
    # API endpoints
    echo ""
    echo "API Health:"
    local api_response=$(curl -s http://localhost:8080/api/health 2>/dev/null)
    if [[ -n "$api_response" ]]; then
        local api_status=$(echo "$api_response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
        if [[ "$api_status" == "healthy" ]]; then
            success "Backend API: Healthy"
        else
            warn "Backend API: Responding but status '$api_status' (Vault not connected)"
            # Still check external API
            if curl -s -f https://api.courtscout.run.place/api/health >/dev/null 2>&1; then
                success "External HTTPS API: Responding"
            else
                error "External HTTPS API: Not responding"
            fi
        fi
    else
        error "Backend API: Not responding"
    fi
    
    # Database connectivity
    echo ""
    echo "Database Health:"
    if docker exec tennis-mongodb mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
        success "MongoDB: Connected"
    else
        error "MongoDB: Connection failed"
    fi
    
    if docker exec tennis-redis redis-cli ping >/dev/null 2>&1; then
        success "Redis: Connected"
    else
        error "Redis: Connection failed"
    fi
}

# Check disk usage
check_disk() {
    info "Disk Usage:"
    df -h /
    echo ""
    
    info "Docker Disk Usage:"
    docker system df
}

# Check memory usage
check_memory() {
    info "Memory Usage:"
    free -h
    echo ""
    
    info "Container Memory Usage:"
    docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}" $(docker ps --filter "name=tennis-" -q)
}

# Check network connectivity
check_network() {
    info "Network Connectivity:"
    
    # External connectivity
    if ping -c 1 google.com >/dev/null 2>&1; then
        success "External connectivity: OK"
    else
        error "External connectivity: Failed"
    fi
    
    # Local services
    if nc -z localhost 8080; then
        success "Backend API port: Open"
    else
        error "Backend API port: Closed"
    fi
    
    if nc -z localhost 443; then
        success "HTTPS port: Open"
    else
        error "HTTPS port: Closed"
    fi
}

# Backup critical data
backup_data() {
    info "Creating backup of critical data..."
    
    local backup_dir="/home/opc/backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # Backup environment file
    if [[ -f "$ENV_FILE" ]]; then
        cp "$ENV_FILE" "$backup_dir/"
        success "Environment file backed up"
    fi
    
    # Backup nginx config
    if [[ -f "$NGINX_CONFIG" ]]; then
        cp "$NGINX_CONFIG" "$backup_dir/"
        success "Nginx config backed up"
    fi
    
    # Backup MongoDB data
    docker exec tennis-mongodb mongodump --out /tmp/backup >/dev/null 2>&1
    docker cp tennis-mongodb:/tmp/backup "$backup_dir/mongodb_backup"
    success "MongoDB data backed up"
    
    info "Backup completed: $backup_dir"
}

# Update system packages
update_system() {
    if check_root; then
        info "Updating system packages..."
        yum update -y
        success "System updated"
    else
        info "Updating system packages (requires sudo)..."
        sudo yum update -y
        success "System updated"
    fi
}

# Main command handling
main() {
    case "${1:-help}" in
        # Core operations
        start|up)
            start_services
            ;;
        stop|down)
            stop_services
            ;;
        restart)
            restart_services "$2"
            ;;
        status)
            show_status
            ;;
        logs|log)
            show_logs "$2" "$3"
            ;;
        
        # Container management
        ps|containers)
            list_containers
            ;;
        pull)
            pull_images
            ;;
        cleanup)
            cleanup_resources
            ;;
        remove)
            remove_all
            ;;
        recreate)
            recreate_service "$2"
            ;;
        
        # Environment & config
        env)
            show_env
            ;;
        config)
            show_config
            ;;
        ssl)
            check_ssl
            ;;
        nginx)
            check_nginx
            ;;
        
        # Health & monitoring
        health)
            health_check
            ;;
        disk)
            check_disk
            ;;
        memory|mem)
            check_memory
            ;;
        network|net)
            check_network
            ;;
        
        # Maintenance
        backup)
            backup_data
            ;;
        update)
            update_system
            ;;
        cert-renew)
            warn "Certificate renewal not implemented yet"
            ;;
        
        # Help
        help|-h|--help)
            show_help
            ;;
        
        *)
            error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 