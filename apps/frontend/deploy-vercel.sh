#!/bin/bash

# 🎾 Tennis Booker Frontend - Vercel Deployment Script

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
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

show_help() {
    cat << EOF
🎾 Tennis Booker Frontend - Vercel Deployment

Usage: $0 <command> [options]

COMMANDS:
  deploy                Deploy to Vercel (production)
  deploy-preview        Deploy preview version
  env                   Manage environment variables
  logs                  Show deployment logs
  status                Check deployment status
  setup                 Initial Vercel setup

OPTIONS:
  --api-url=URL         Set the backend API URL
  --domain=DOMAIN       Set custom domain
  --env-file=FILE       Load environment from file

EXAMPLES:
  $0 setup                                    # Initial setup
  $0 deploy --api-url=https://api.example.com # Deploy with custom API
  $0 env set VITE_API_URL https://api.example.com # Set environment variable
  $0 logs                                     # View logs

SETUP REQUIREMENTS:
1. Install Vercel CLI: npm i -g vercel
2. Login to Vercel: vercel login
3. Have your backend API URL ready

EOF
}

check_prerequisites() {
    info "Checking prerequisites..."
    
    # Check if we're in the frontend directory
    if [[ ! -f "package.json" ]]; then
        error "This script must be run from the frontend directory"
        return 1
    fi
    
    # Check Vercel CLI
    if ! command -v vercel >/dev/null 2>&1; then
        error "Vercel CLI not found. Install it with: npm i -g vercel"
        return 1
    fi
    
    # Check if logged in to Vercel
    if ! vercel whoami >/dev/null 2>&1; then
        error "Not logged in to Vercel. Run: vercel login"
        return 1
    fi
    
    success "Prerequisites check passed"
}

setup_vercel() {
    info "Setting up Vercel project..."
    
    # Initialize Vercel project
    vercel
    
    success "Vercel project setup complete"
    info "Next steps:"
    echo "  1. Set your backend API URL: $0 env set VITE_API_URL https://your-backend-url.com/api"
    echo "  2. Deploy: $0 deploy"
}

deploy_production() {
    local api_url="$1"
    local custom_domain="$2"
    
    info "Deploying to Vercel production..."
    
    # Set API URL if provided
    if [[ -n "$api_url" ]]; then
        info "Setting API URL: $api_url"
        vercel env add VITE_API_URL production <<< "$api_url" || true
    fi
    
    # Build and deploy
    vercel --prod
    
    success "Deployment completed!"
    
    # Get deployment URL
    local deployment_url
    deployment_url=$(vercel ls | grep "$(pwd | xargs basename)" | head -1 | awk '{print $2}')
    
    if [[ -n "$deployment_url" ]]; then
        info "Deployment URL: https://$deployment_url"
    fi
    
    # Configure custom domain if provided
    if [[ -n "$custom_domain" ]]; then
        info "Adding custom domain: $custom_domain"
        vercel domains add "$custom_domain" || warn "Failed to add domain. You may need to configure it manually in Vercel dashboard."
    fi
}

deploy_preview() {
    info "Deploying preview to Vercel..."
    
    vercel
    
    success "Preview deployment completed!"
}

manage_env() {
    local action="$1"
    local key="$2"
    local value="$3"
    local environment="${4:-production}"
    
    case "$action" in
        set)
            if [[ -z "$key" || -z "$value" ]]; then
                error "Usage: $0 env set KEY VALUE [environment]"
                return 1
            fi
            info "Setting $key for $environment environment..."
            echo "$value" | vercel env add "$key" "$environment"
            success "Environment variable set"
            ;;
        list|ls)
            info "Environment variables:"
            vercel env ls
            ;;
        remove|rm)
            if [[ -z "$key" ]]; then
                error "Usage: $0 env remove KEY [environment]"
                return 1
            fi
            vercel env rm "$key" "$environment"
            success "Environment variable removed"
            ;;
        pull)
            info "Pulling environment variables..."
            vercel env pull .env.local
            success "Environment variables saved to .env.local"
            ;;
        *)
            error "Unknown env action: $action"
            echo "Available actions: set, list, remove, pull"
            return 1
            ;;
    esac
}

show_logs() {
    info "Showing Vercel logs..."
    vercel logs --follow
}

check_status() {
    info "Checking deployment status..."
    
    # List deployments
    vercel ls
    
    # Show current domains
    info "Configured domains:"
    vercel domains ls
    
    # Show environment variables
    info "Environment variables:"
    vercel env ls
}

load_env_file() {
    local env_file="$1"
    
    if [[ ! -f "$env_file" ]]; then
        error "Environment file not found: $env_file"
        return 1
    fi
    
    info "Loading environment variables from $env_file..."
    
    # Read each line from the env file
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip comments and empty lines
        if [[ "$line" =~ ^[[:space:]]*# ]] || [[ -z "${line// }" ]]; then
            continue
        fi
        
        # Extract key and value
        if [[ "$line" =~ ^([^=]+)=(.*)$ ]]; then
            local key="${BASH_REMATCH[1]}"
            local value="${BASH_REMATCH[2]}"
            
            # Remove quotes if present
            value="${value#\"}"
            value="${value%\"}"
            value="${value#\'}"
            value="${value%\'}"
            
            info "Setting $key..."
            echo "$value" | vercel env add "$key" production || warn "Failed to set $key"
        fi
    done < "$env_file"
    
    success "Environment variables loaded from $env_file"
}

# Parse command line arguments
API_URL=""
CUSTOM_DOMAIN=""
ENV_FILE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --api-url=*)
            API_URL="${1#*=}"
            shift
            ;;
        --domain=*)
            CUSTOM_DOMAIN="${1#*=}"
            shift
            ;;
        --env-file=*)
            ENV_FILE="${1#*=}"
            shift
            ;;
        *)
            break
            ;;
    esac
done

# Main command handling
main() {
    case "${1:-help}" in
        setup)
            check_prerequisites
            setup_vercel
            ;;
        deploy)
            check_prerequisites
            if [[ -n "$ENV_FILE" ]]; then
                load_env_file "$ENV_FILE"
            fi
            deploy_production "$API_URL" "$CUSTOM_DOMAIN"
            ;;
        deploy-preview|preview)
            check_prerequisites
            deploy_preview
            ;;
        env)
            check_prerequisites
            manage_env "$2" "$3" "$4" "$5"
            ;;
        logs)
            check_prerequisites
            show_logs
            ;;
        status)
            check_prerequisites
            check_status
            ;;
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