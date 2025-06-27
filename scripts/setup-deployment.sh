#!/bin/bash

# 🎾 Tennis Booker - Deployment Setup Script
# Quick setup for OCI + Vercel deployment

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

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

show_help() {
    cat << EOF
🎾 Tennis Booker - Deployment Setup

This script helps you configure Tennis Booker for deployment:
- Backend on Oracle Cloud Infrastructure (OCI)
- Frontend on Vercel

Usage: $0 [options]

OPTIONS:
  --domain=DOMAIN       Your domain name (e.g., example.com)
  --email=EMAIL         Your email for certificates and admin user
  --api-url=URL         Frontend API URL (auto-generated if domain provided)
  --help                Show this help

EXAMPLES:
  $0 --domain=tennisbooker.com --email=admin@tennisbooker.com
  $0 --api-url=https://api.example.com --email=admin@example.com

WHAT THIS SCRIPT DOES:
1. Generates secure passwords and secrets
2. Creates environment configuration files
3. Sets up SSH keys for OCI deployment
4. Provides next steps for deployment

REQUIREMENTS:
- OpenSSL (for generating secrets)
- SSH (for key generation)

EOF
}

generate_secret() {
    local length="${1:-32}"
    openssl rand -base64 "$length" | tr -d "=+/" | cut -c1-"$length"
}

generate_password_hash() {
    local password="$1"
    # Generate bcrypt hash using Python if available, otherwise use basic method
    if command -v python3 >/dev/null 2>&1; then
        python3 -c "import bcrypt; print(bcrypt.hashpw('$password'.encode('utf-8'), bcrypt.gensalt()).decode('utf-8'))" 2>/dev/null || echo '$2y$10$'$(openssl passwd -1 "$password" | cut -d'$' -f4)
    else
        echo '$2y$10$'$(openssl passwd -1 "$password" | cut -d'$' -f4)
    fi
}

setup_ssh_keys() {
    local ssh_key_path="$HOME/.ssh/tennis_booker_key"
    
    if [[ ! -f "$ssh_key_path" ]]; then
        info "Generating SSH key pair for OCI deployment..."
        ssh-keygen -t rsa -b 4096 -f "$ssh_key_path" -N "" -C "tennis-booker-oci-$(date +%Y%m%d)"
        success "SSH key generated: $ssh_key_path"
    else
        info "SSH key already exists: $ssh_key_path"
    fi
    
    echo "$ssh_key_path.pub"
}

create_terraform_tfvars() {
    local domain="$1"
    local email="$2"
    local ssh_public_key_file="$3"
    
    local tfvars_file="$PROJECT_ROOT/infrastructure/terraform/terraform.tfvars"
    local ssh_public_key
    ssh_public_key=$(cat "$ssh_public_key_file")
    
    info "Creating Terraform configuration..."
    
    cat > "$tfvars_file" << EOF
# OCI Authentication - UPDATE WITH YOUR ACTUAL VALUES
tenancy_ocid     = "ocid1.tenancy.oc1..your-tenancy-ocid-here"
user_ocid        = "ocid1.user.oc1..your-user-ocid-here"
fingerprint      = "your-api-key-fingerprint-here"
private_key_path = "~/.oci/oci_api_key.pem"
region           = "us-ashburn-1"

# Project Configuration
project_name     = "tennis-booker"
environment      = "prod"
compartment_ocid = ""

# Backend Configuration
bucket_name = "tennis-booker-terraform-state"
namespace   = "your-namespace-here"

# Network Configuration
vcn_cidr            = "10.0.0.0/16"
public_subnet_cidr  = "10.0.1.0/24"
private_subnet_cidr = "10.0.2.0/24"

# Compute Configuration (Always Free Tier)
instance_shape          = "VM.Standard.A1.Flex"
instance_ocpus          = 2
instance_memory_in_gbs  = 12
ssh_public_key          = "$ssh_public_key"

# Storage Configuration
block_volume_size_in_gbs = 50

# Domain Configuration
domain_name = "$domain"
acme_email  = "$email"
EOF
    
    success "Terraform configuration created: $tfvars_file"
    warn "IMPORTANT: Update the OCI credentials in $tfvars_file"
}

create_production_env() {
    local domain="$1"
    local email="$2"
    local api_url="$3"
    
    local env_file="$PROJECT_ROOT/.env.production"
    
    # Generate secure passwords and secrets
    local mongo_password jwt_secret redis_password traefik_password traefik_password_hash
    mongo_password=$(generate_secret 24)
    jwt_secret=$(generate_secret 64)
    redis_password=$(generate_secret 24)
    traefik_password=$(generate_secret 16)
    traefik_password_hash=$(generate_password_hash "$traefik_password")
    
    info "Creating production environment configuration..."
    
    cat > "$env_file" << EOF
# 🎾 Tennis Booker - Production Environment Configuration
# Generated on $(date)

# ================================
# GENERAL CONFIGURATION
# ================================
ENVIRONMENT=production
NODE_ENV=production

# ================================
# DOMAIN & SSL CONFIGURATION
# ================================
DOMAIN_NAME=$domain
ACME_EMAIL=$email

# CORS Origins (update with your Vercel domain)
CORS_ORIGINS=https://$domain,https://www.$domain,https://your-app.vercel.app

# ================================
# DATABASE CONFIGURATION
# ================================
MONGO_ROOT_USERNAME=admin
MONGO_ROOT_PASSWORD=$mongo_password
MONGO_URI=mongodb://admin:$mongo_password@mongodb:27017/tennis_booking?authSource=admin
DB_NAME=tennis_booking

# ================================
# REDIS CONFIGURATION
# ================================
REDIS_PASSWORD=$redis_password
REDIS_ADDR=redis:6379

# ================================
# SECURITY & AUTHENTICATION
# ================================
JWT_SECRET=$jwt_secret
HASH_COST=12

# ================================
# EMAIL CONFIGURATION (UPDATE WITH YOUR GMAIL CREDENTIALS)
# ================================
GMAIL_EMAIL=your_gmail@gmail.com
GMAIL_PASSWORD=your_gmail_app_specific_password
FROM_EMAIL=Tennis Booker <noreply@$domain>

# ================================
# SCRAPER CONFIGURATION
# ================================
SCRAPER_INTERVAL=300
SCRAPER_TIMEOUT=30000

# ================================
# TRAEFIK CONFIGURATION
# ================================
TRAEFIK_USER=admin
TRAEFIK_PASSWORD=$traefik_password
TRAEFIK_PASSWORD_HASH=$traefik_password_hash

# ================================
# USER SEEDING
# ================================
USER_EMAIL=$email

# ================================
# VAULT CONFIGURATION
# ================================
VAULT_ADDR=http://vault:8200

# ================================
# FEATURE FLAGS
# ================================
ENABLE_REGISTRATION=true
ENABLE_NOTIFICATIONS=true
ENABLE_SCRAPING=true
ENABLE_RATE_LIMITING=true

# ================================
# PERFORMANCE TUNING
# ================================
MAX_CONCURRENT_REQUESTS=100
API_TIMEOUT=30000
DATABASE_POOL_SIZE=10
EOF
    
    success "Production environment created: $env_file"
    
    # Show important credentials
    echo ""
    info "IMPORTANT CREDENTIALS (save these securely):"
    echo "  Traefik Admin: admin / $traefik_password"
    echo "  MongoDB: admin / $mongo_password"
    echo "  Redis: $redis_password"
    echo ""
    warn "Remember to update GMAIL_EMAIL and GMAIL_PASSWORD in $env_file"
}

create_frontend_env() {
    local api_url="$1"
    
    local env_file="$PROJECT_ROOT/apps/frontend/.env.production"
    
    info "Creating frontend environment configuration..."
    
    cat > "$env_file" << EOF
# 🎾 Tennis Booker Frontend - Production Environment
# Generated on $(date)

# API Configuration
VITE_API_URL=$api_url
VITE_API_TIMEOUT=30000

# Application Configuration
VITE_APP_NAME=Tennis Booker
VITE_APP_ENVIRONMENT=production

# Feature Flags
VITE_FEATURE_NOTIFICATIONS_ENABLED=true
VITE_FEATURE_ANALYTICS_ENABLED=true
VITE_FEATURE_ADVANCED_SEARCH_ENABLED=true
VITE_FEATURE_DARK_MODE_ENABLED=true
VITE_DEBUG_MODE=false

# Performance
VITE_BUILD_SOURCEMAP=false
VITE_BUILD_MINIFY=true

# Optional: Add these when you have them
# VITE_GOOGLE_ANALYTICS_ID=GA_MEASUREMENT_ID
# VITE_SENTRY_DSN=SENTRY_DSN_URL
EOF
    
    success "Frontend environment created: $env_file"
}

print_next_steps() {
    local domain="$1"
    local api_url="$2"
    
    echo ""
    success "🎾 Setup completed! Next steps:"
    echo ""
    echo "1. 🏗️  BACKEND DEPLOYMENT (OCI):"
    echo "   a. Set up OCI account and API keys"
    echo "   b. Update infrastructure/terraform/terraform.tfvars with your OCI credentials"
    echo "   c. Deploy infrastructure:"
    echo "      cd $PROJECT_ROOT"
    echo "      ./infrastructure/scripts/deploy-oci.sh init"
    echo "      ./infrastructure/scripts/deploy-oci.sh apply"
    echo "   d. Deploy application:"
    echo "      ./infrastructure/scripts/deploy-oci.sh deploy-app"
    echo ""
    echo "2. 🌐 FRONTEND DEPLOYMENT (Vercel):"
    echo "   a. Install Vercel CLI: npm i -g vercel"
    echo "   b. Login: vercel login"
    echo "   c. Deploy:"
    echo "      cd apps/frontend"
    echo "      ./deploy-vercel.sh setup"
    echo "      ./deploy-vercel.sh deploy --api-url=$api_url"
    echo ""
    echo "3. 🔧 CONFIGURATION:"
    echo "   a. Update Gmail credentials in .env.production"
    echo "   b. Point your domain DNS to OCI instance IP"
    echo "   c. Update CORS_ORIGINS in .env.production with your Vercel domain"
    echo ""
    echo "4. 📋 DOCUMENTATION:"
    echo "   - Read DEPLOYMENT.md for detailed instructions"
    echo "   - Check OCI Always Free limits"
    echo "   - Configure monitoring and backups"
    echo ""
    info "Files created:"
    echo "  - infrastructure/terraform/terraform.tfvars"
    echo "  - .env.production"
    echo "  - apps/frontend/.env.production"
    echo "  - ~/.ssh/tennis_booker_key (SSH key pair)"
    echo ""
    warn "⚠️  Remember to:"
    echo "  - Keep your .env.production file secure"
    echo "  - Update OCI credentials in terraform.tfvars"
    echo "  - Set up Gmail app-specific password"
    echo "  - Configure your domain DNS"
    echo ""
    success "Happy deploying! 🚀"
}

# Parse command line arguments
DOMAIN=""
EMAIL=""
API_URL=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --domain=*)
            DOMAIN="${1#*=}"
            shift
            ;;
        --email=*)
            EMAIL="${1#*=}"
            shift
            ;;
        --api-url=*)
            API_URL="${1#*=}"
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate inputs
if [[ -z "$DOMAIN" && -z "$API_URL" ]]; then
    error "Either --domain or --api-url must be provided"
    show_help
    exit 1
fi

if [[ -z "$EMAIL" ]]; then
    error "--email is required"
    show_help
    exit 1
fi

# Auto-generate API URL if not provided
if [[ -z "$API_URL" && -n "$DOMAIN" ]]; then
    API_URL="https://api.$DOMAIN/api"
fi

# Default domain if not provided
if [[ -z "$DOMAIN" ]]; then
    DOMAIN="example.com"
    warn "No domain provided, using example.com (update later)"
fi

# Check prerequisites
info "Checking prerequisites..."
if ! command -v openssl >/dev/null 2>&1; then
    error "OpenSSL is required but not installed"
    exit 1
fi

if ! command -v ssh-keygen >/dev/null 2>&1; then
    error "SSH is required but not installed"
    exit 1
fi

# Run setup
info "🎾 Setting up Tennis Booker deployment..."
info "Domain: $DOMAIN"
info "Email: $EMAIL"
info "API URL: $API_URL"
echo ""

# Setup SSH keys
SSH_PUBLIC_KEY_FILE=$(setup_ssh_keys)

# Create configuration files
create_terraform_tfvars "$DOMAIN" "$EMAIL" "$SSH_PUBLIC_KEY_FILE"
create_production_env "$DOMAIN" "$EMAIL" "$API_URL"
create_frontend_env "$API_URL"

# Print next steps
print_next_steps "$DOMAIN" "$API_URL"