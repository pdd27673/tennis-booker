#!/bin/bash

# Tennis Court Data Retention Service Deployment Script
# This script builds and deploys the retention service to Kubernetes

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
IMAGE_NAME="tennis-booker/retention-service"
NAMESPACE="tennis-booker"
DRY_RUN=${DRY_RUN:-false}
SKIP_BUILD=${SKIP_BUILD:-false}
SKIP_TESTS=${SKIP_TESTS:-false}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
Tennis Court Data Retention Service Deployment Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -d, --dry-run           Perform a dry run (don't actually deploy)
    -s, --skip-build        Skip building the Docker image
    -t, --skip-tests        Skip running tests before deployment
    -n, --namespace NAME    Kubernetes namespace (default: tennis-booker)
    -i, --image NAME        Docker image name (default: tennis-booker/retention-service)
    --tag TAG               Docker image tag (default: latest)

Environment Variables:
    DRY_RUN                 Set to 'true' for dry run mode
    SKIP_BUILD              Set to 'true' to skip building
    SKIP_TESTS              Set to 'true' to skip tests
    KUBECONFIG              Path to kubeconfig file

Examples:
    $0                      # Deploy with defaults
    $0 --dry-run            # Dry run deployment
    $0 --skip-build         # Deploy without rebuilding image
    $0 --tag v1.2.3         # Deploy specific version

EOF
}

# Parse command line arguments
IMAGE_TAG="latest"
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -s|--skip-build)
            SKIP_BUILD=true
            shift
            ;;
        -t|--skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -i|--image)
            IMAGE_NAME="$2"
            shift 2
            ;;
        --tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if docker is available (unless skipping build)
    if [[ "$SKIP_BUILD" != "true" ]] && ! command -v docker &> /dev/null; then
        log_error "docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we can connect to Kubernetes cluster
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Check your kubeconfig."
        exit 1
    fi
    
    # Check if namespace exists
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warning "Namespace '$NAMESPACE' does not exist. Creating it..."
        if [[ "$DRY_RUN" != "true" ]]; then
            kubectl create namespace "$NAMESPACE"
        else
            log_info "[DRY RUN] Would create namespace: $NAMESPACE"
        fi
    fi
    
    log_success "Prerequisites check passed"
}

# Run tests
run_tests() {
    if [[ "$SKIP_TESTS" == "true" ]]; then
        log_info "Skipping tests (--skip-tests flag set)"
        return 0
    fi
    
    log_info "Running tests..."
    cd "$PROJECT_ROOT"
    
    # Run Go tests
    if ! make test-short; then
        log_error "Tests failed"
        exit 1
    fi
    
    log_success "All tests passed"
}

# Build Docker image
build_image() {
    if [[ "$SKIP_BUILD" == "true" ]]; then
        log_info "Skipping Docker build (--skip-build flag set)"
        return 0
    fi
    
    log_info "Building Docker image: $FULL_IMAGE_NAME"
    cd "$PROJECT_ROOT"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would build Docker image: $FULL_IMAGE_NAME"
        return 0
    fi
    
    # Build the image
    if ! docker build -f Dockerfile.retention -t "$FULL_IMAGE_NAME" .; then
        log_error "Docker build failed"
        exit 1
    fi
    
    log_success "Docker image built successfully: $FULL_IMAGE_NAME"
}

# Deploy to Kubernetes
deploy_to_kubernetes() {
    log_info "Deploying retention service to Kubernetes..."
    cd "$PROJECT_ROOT"
    
    # Update image tag in manifests if not using latest
    if [[ "$IMAGE_TAG" != "latest" ]]; then
        log_info "Updating image tag in manifests to: $IMAGE_TAG"
        
        # Create temporary manifests with updated image tag
        sed "s|tennis-booker/retention-service:latest|$FULL_IMAGE_NAME|g" \
            k8s/retention-service-cronjob.yaml > /tmp/retention-service-cronjob-temp.yaml
        sed "s|tennis-booker/retention-service:latest|$FULL_IMAGE_NAME|g" \
            k8s/retention-service-monitoring.yaml > /tmp/retention-service-monitoring-temp.yaml
        
        CRONJOB_MANIFEST="/tmp/retention-service-cronjob-temp.yaml"
        MONITORING_MANIFEST="/tmp/retention-service-monitoring-temp.yaml"
    else
        CRONJOB_MANIFEST="k8s/retention-service-cronjob.yaml"
        MONITORING_MANIFEST="k8s/retention-service-monitoring.yaml"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would apply the following manifests:"
        log_info "  - $CRONJOB_MANIFEST"
        log_info "  - $MONITORING_MANIFEST"
        
        log_info "[DRY RUN] CronJob manifest preview:"
        kubectl apply --dry-run=client -f "$CRONJOB_MANIFEST"
        
        log_info "[DRY RUN] Monitoring manifest preview:"
        kubectl apply --dry-run=client -f "$MONITORING_MANIFEST"
        
        return 0
    fi
    
    # Apply the manifests
    log_info "Applying CronJob manifest..."
    kubectl apply -f "$CRONJOB_MANIFEST"
    
    log_info "Applying monitoring manifest..."
    kubectl apply -f "$MONITORING_MANIFEST"
    
    # Clean up temporary files
    if [[ "$IMAGE_TAG" != "latest" ]]; then
        rm -f /tmp/retention-service-cronjob-temp.yaml /tmp/retention-service-monitoring-temp.yaml
    fi
    
    log_success "Retention service deployed successfully"
}

# Verify deployment
verify_deployment() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would verify deployment"
        return 0
    fi
    
    log_info "Verifying deployment..."
    
    # Check if CronJob was created
    if kubectl get cronjob tennis-court-retention-service -n "$NAMESPACE" &> /dev/null; then
        log_success "CronJob created successfully"
        
        # Show CronJob details
        log_info "CronJob details:"
        kubectl describe cronjob tennis-court-retention-service -n "$NAMESPACE"
    else
        log_error "CronJob was not created"
        exit 1
    fi
    
    # Check if monitoring resources were created
    if kubectl get prometheusrule tennis-court-retention-alerts -n "$NAMESPACE" &> /dev/null; then
        log_success "Monitoring alerts configured"
    else
        log_warning "Monitoring alerts not found (may not be available in this cluster)"
    fi
}

# Run a test job
run_test_job() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would run test job"
        return 0
    fi
    
    log_info "Running test job to verify functionality..."
    
    # Apply test job manifest
    kubectl apply -f k8s/retention-service-monitoring.yaml
    
    # Wait for job completion
    log_info "Waiting for test job to complete (timeout: 5 minutes)..."
    if kubectl wait --for=condition=complete job/retention-service-test-job -n "$NAMESPACE" --timeout=300s; then
        log_success "Test job completed successfully"
        
        # Show logs
        log_info "Test job logs:"
        kubectl logs job/retention-service-test-job -n "$NAMESPACE"
    else
        log_error "Test job failed or timed out"
        log_info "Test job logs:"
        kubectl logs job/retention-service-test-job -n "$NAMESPACE" || true
        exit 1
    fi
}

# Main deployment function
main() {
    log_info "Starting Tennis Court Data Retention Service deployment..."
    log_info "Configuration:"
    log_info "  - Image: $FULL_IMAGE_NAME"
    log_info "  - Namespace: $NAMESPACE"
    log_info "  - Dry Run: $DRY_RUN"
    log_info "  - Skip Build: $SKIP_BUILD"
    log_info "  - Skip Tests: $SKIP_TESTS"
    
    check_prerequisites
    run_tests
    build_image
    deploy_to_kubernetes
    verify_deployment
    run_test_job
    
    log_success "Deployment completed successfully!"
    
    if [[ "$DRY_RUN" != "true" ]]; then
        log_info "Next steps:"
        log_info "  1. Monitor the CronJob: kubectl get cronjob tennis-court-retention-service -n $NAMESPACE"
        log_info "  2. Check job history: kubectl get jobs -n $NAMESPACE -l app=tennis-court-retention-service"
        log_info "  3. View logs: kubectl logs -l app=tennis-court-retention-service -n $NAMESPACE"
        log_info "  4. Monitor alerts in your monitoring system"
    fi
}

# Run main function
main "$@" 