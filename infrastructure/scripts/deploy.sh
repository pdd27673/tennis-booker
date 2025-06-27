#!/bin/bash

# Tennis Booker Production Deployment Script
# This script sets up the complete production environment on OCI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
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

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    error "Please do not run this script as root"
fi

# Check if .env.prod exists
if [ ! -f ".env.prod" ]; then
    error "Production environment file .env.prod not found. Please copy env.prod.example to .env.prod and configure it."
fi

# Load environment variables
source .env.prod

# Validate required environment variables
required_vars=(
    "DOMAIN_NAME"
    "DUCKDNS_TOKEN"
    "ACME_EMAIL"
    "TRAEFIK_USER"
    "TRAEFIK_PASSWORD_HASH"
    "MONGO_ROOT_USERNAME"
    "MONGO_ROOT_PASSWORD"
    "REDIS_PASSWORD"
    "JWT_SECRET"
    "GMAIL_EMAIL"
    "GMAIL_PASSWORD"
)

log "Validating environment variables..."
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        error "Required environment variable $var is not set in .env.prod"
    fi
done
success "Environment variables validated"

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install Docker if not present
install_docker() {
    if ! command_exists docker; then
        log "Installing Docker..."
        sudo dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
        sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        sudo systemctl start docker
        sudo systemctl enable docker
        sudo usermod -aG docker $USER
        success "Docker installed successfully"
        warning "Please log out and log back in for Docker group membership to take effect"
    else
        success "Docker is already installed"
    fi
}

# Install Docker Compose if not present
install_docker_compose() {
    if ! command_exists docker-compose; then
        log "Installing Docker Compose..."
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
        success "Docker Compose installed successfully"
    else
        success "Docker Compose is already installed"
    fi
}

# Setup firewall rules
setup_firewall() {
    log "Configuring firewall..."
    
    # Check if firewalld is running
    if sudo systemctl is-active --quiet firewalld; then
        sudo firewall-cmd --permanent --add-port=80/tcp
        sudo firewall-cmd --permanent --add-port=443/tcp
        sudo firewall-cmd --permanent --add-port=8080/tcp  # Traefik dashboard
        sudo firewall-cmd --reload
        success "Firewall configured"
    else
        warning "Firewalld is not running. Please ensure ports 80, 443, and 8080 are open"
    fi
}

# Setup DuckDNS
setup_duckdns() {
    log "Setting up DuckDNS..."
    
    # Update DuckDNS with current IP
    echo "DuckDNS update script not available - configure manually if needed"
    
    # Install systemd service for automatic updates
    sudo cp infrastructure/systemd/duckdns-update.service /etc/systemd/system/
    sudo cp infrastructure/systemd/duckdns-update.timer /etc/systemd/system/
    
    # Update service file with correct path
    sudo sed -i "s|/opt/tennis-booker|$(pwd)|g" /etc/systemd/system/duckdns-update.service
    
    sudo systemctl daemon-reload
    sudo systemctl enable duckdns-update.timer
    sudo systemctl start duckdns-update.timer
    
    success "DuckDNS automatic updates configured"
}

# Create necessary directories
create_directories() {
    log "Creating necessary directories..."
    
    mkdir -p logs
    mkdir -p backups
    mkdir -p data/mongodb
    mkdir -p data/redis
    mkdir -p data/vault
    mkdir -p data/traefik
    
    # Set proper permissions
    chmod 755 logs backups data
    chmod 700 data/vault
    
    success "Directories created"
}

# Generate Traefik password if not set
generate_traefik_password() {
    if [ "$TRAEFIK_PASSWORD_HASH" = '$2y$10$example.hash.here' ]; then
        warning "Default Traefik password hash detected. Please run:"
        echo "./infrastructure/traefik/generate-password.sh admin your-secure-password"
        echo "Then update TRAEFIK_PASSWORD_HASH in .env.prod"
        exit 1
    fi
}

# Build and deploy services
deploy_services() {
    log "Building and deploying services..."
    
    # Pull latest images
    docker-compose -f docker-compose.prod.yml pull
    
    # Build custom images
    docker-compose -f docker-compose.prod.yml build --no-cache
    
    # Start services
    docker-compose -f docker-compose.prod.yml up -d
    
    success "Services deployed"
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

# Verify deployment
verify_deployment() {
    log "Verifying deployment..."
    
    # Check if domain resolves
    if nslookup "$DOMAIN_NAME" >/dev/null 2>&1; then
        success "Domain $DOMAIN_NAME resolves correctly"
    else
        warning "Domain $DOMAIN_NAME does not resolve yet. DNS propagation may take time."
    fi
    
    # Check services
    log "Service status:"
    docker-compose -f docker-compose.prod.yml ps
    
    # Test endpoints
    log "Testing endpoints..."
    
    # Test if Traefik is responding
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:80 | grep -q "404\|301\|302"; then
        success "Traefik is responding"
    else
        warning "Traefik may not be responding correctly"
    fi
    
    success "Deployment verification complete"
}

# Show post-deployment information
show_info() {
    log "Deployment complete! 🎉"
    echo
    echo "=== Access Information ==="
    echo "Frontend: https://$DOMAIN_NAME"
    echo "API: https://api.$DOMAIN_NAME"
    echo "Traefik Dashboard: https://traefik.$DOMAIN_NAME (user: $TRAEFIK_USER)"
    echo
    echo "=== Useful Commands ==="
    echo "View logs: docker-compose -f docker-compose.prod.yml logs -f [service]"
    echo "Restart services: docker-compose -f docker-compose.prod.yml restart"
    echo "Update services: docker-compose -f docker-compose.prod.yml pull && docker-compose -f docker-compose.prod.yml up -d"
    echo "Backup data: Configure your own backup solution"
    echo
    echo "=== SSL Certificate ==="
    echo "Let's Encrypt certificates will be automatically generated on first access."
    echo "This may take a few minutes. Monitor with:"
    echo "docker-compose -f docker-compose.prod.yml logs -f traefik"
    echo
}

# Main deployment flow
main() {
    log "Starting Tennis Booker production deployment..."
    
    install_docker
    install_docker_compose
    setup_firewall
    create_directories
    generate_traefik_password
    setup_duckdns
    deploy_services
    wait_for_services
    verify_deployment
    show_info
    
    success "Deployment completed successfully!"
}

# Run main function
main "$@" 