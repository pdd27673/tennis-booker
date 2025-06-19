#!/bin/bash

# Tennis Court Booking System - Cloud Deployment Script
# This script helps deploy the system to various cloud platforms

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
REGISTRY_NAME="tennis-booking"
IMAGE_TAG=${IMAGE_TAG:-"latest"}

show_help() {
    cat << EOF
🎾 Tennis Court Booking System - Cloud Deployment

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    build               Build Docker images locally
    push-registry       Push images to Docker registry
    deploy-docker       Deploy using Docker Compose
    deploy-aws          Deploy to AWS ECS
    deploy-gcp          Deploy to GCP Cloud Run
    deploy-azure        Deploy to Azure Container Instances
    
Options:
    --tag TAG          Image tag (default: latest)
    --registry REG     Docker registry URL
    --help            Show this help

Examples:
    $0 build
    $0 deploy-docker
    $0 push-registry --registry myregistry.io
    $0 deploy-aws --tag v1.0.0

EOF
}

build_images() {
    info "Building Docker images..."
    
    docker build -f Dockerfile.notification -t ${REGISTRY_NAME}/notification-service:${IMAGE_TAG} .
    docker build -f Dockerfile.scraper -t ${REGISTRY_NAME}/scraper-service:${IMAGE_TAG} .
    
    success "Images built successfully!"
    docker images | grep ${REGISTRY_NAME}
}

push_to_registry() {
    local registry=$1
    if [ -z "$registry" ]; then
        error "Registry URL required. Use --registry option"
        exit 1
    fi
    
    info "Pushing images to registry: $registry"
    
    # Tag images for registry
    docker tag ${REGISTRY_NAME}/notification-service:${IMAGE_TAG} ${registry}/notification-service:${IMAGE_TAG}
    docker tag ${REGISTRY_NAME}/scraper-service:${IMAGE_TAG} ${registry}/scraper-service:${IMAGE_TAG}
    
    # Push images
    docker push ${registry}/notification-service:${IMAGE_TAG}
    docker push ${registry}/scraper-service:${IMAGE_TAG}
    
    success "Images pushed to registry!"
}

deploy_docker() {
    info "Deploying with Docker Compose..."
    
    # Check if .env exists
    if [ ! -f .env ]; then
        warn "No .env file found. Creating from template..."
        cp env.prod.example .env
        warn "Please edit .env file with your configuration before running again"
        exit 1
    fi
    
    # Deploy using production compose file
    docker-compose -f docker-compose.prod.yml up -d
    
    success "System deployed! 🎾"
    echo ""
    echo "Services running:"
    docker-compose -f docker-compose.prod.yml ps
    
    echo ""
    echo "To check logs:"
    echo "  docker-compose -f docker-compose.prod.yml logs -f"
    echo ""
    echo "To stop:"
    echo "  docker-compose -f docker-compose.prod.yml down"
}

deploy_aws() {
    info "AWS ECS deployment instructions..."
    cat << EOF

📋 AWS ECS Deployment Steps:

1. Setup AWS CLI and login:
   aws configure

2. Create ECR repositories:
   aws ecr create-repository --repository-name tennis-booking/notification-service
   aws ecr create-repository --repository-name tennis-booking/scraper-service

3. Get ECR login:
   aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT.dkr.ecr.us-east-1.amazonaws.com

4. Build and push images:
   $0 build
   $0 push-registry --registry ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/tennis-booking

5. Create ECS cluster and services using AWS Console or CLI

6. Set environment variables in ECS task definitions

EOF
}

deploy_gcp() {
    info "GCP Cloud Run deployment instructions..."
    cat << EOF

📋 GCP Cloud Run Deployment Steps:

1. Setup gcloud CLI and login:
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID

2. Enable APIs:
   gcloud services enable run.googleapis.com
   gcloud services enable cloudbuild.googleapis.com

3. Build and deploy notification service:
   gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/notification-service .
   gcloud run deploy notification-service --image gcr.io/YOUR_PROJECT_ID/notification-service --platform managed

4. Build and deploy scraper service:
   gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/scraper-service -f Dockerfile.scraper .
   gcloud run deploy scraper-service --image gcr.io/YOUR_PROJECT_ID/scraper-service --platform managed

5. Set environment variables in Cloud Run console

EOF
}

deploy_azure() {
    info "Azure Container Instances deployment instructions..."
    cat << EOF

📋 Azure Container Instances Deployment Steps:

1. Setup Azure CLI and login:
   az login

2. Create resource group:
   az group create --name tennis-booking-rg --location eastus

3. Create container registry:
   az acr create --resource-group tennis-booking-rg --name tennisbookingregistry --sku Basic

4. Build and push images:
   az acr build --registry tennisbookingregistry --image notification-service .
   az acr build --registry tennisbookingregistry --image scraper-service -f Dockerfile.scraper .

5. Create container instances using Azure Portal or CLI

EOF
}

# Parse command line arguments
COMMAND=""
REGISTRY=""

while [[ $# -gt 0 ]]; do
    case $1 in
        build|push-registry|deploy-docker|deploy-aws|deploy-gcp|deploy-azure)
            COMMAND="$1"
            shift
            ;;
        --tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        --registry)
            REGISTRY="$2"
            shift 2
            ;;
        --help)
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

# Execute command
case $COMMAND in
    build)
        build_images
        ;;
    push-registry)
        push_to_registry "$REGISTRY"
        ;;
    deploy-docker)
        deploy_docker
        ;;
    deploy-aws)
        deploy_aws
        ;;
    deploy-gcp)
        deploy_gcp
        ;;
    deploy-azure)
        deploy_azure
        ;;
    "")
        error "No command specified"
        show_help
        exit 1
        ;;
esac 