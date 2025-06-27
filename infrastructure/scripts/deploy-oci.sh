#!/bin/bash

# 🎾 Tennis Booker - OCI Deployment Script
# Automated deployment to Oracle Cloud Infrastructure

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
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/infrastructure/terraform"
BACKEND_CONF="$TERRAFORM_DIR/backend.conf"
TFVARS_FILE="$TERRAFORM_DIR/terraform.tfvars"

show_help() {
    cat << EOF
🎾 Tennis Booker OCI Deployment

Usage: $0 <command> [options]

COMMANDS:
  init                  Initialize Terraform (first time setup)
  plan                  Plan Terraform changes
  apply                 Apply Terraform changes (deploy)
  destroy               Destroy infrastructure
  output                Show Terraform outputs
  connect               Connect to the deployed instance
  deploy-app            Deploy application to existing instance
  status                Check deployment status

OPTIONS:
  --auto-approve        Auto-approve Terraform apply/destroy
  --var-file=FILE       Use custom tfvars file
  --backend-config=FILE Use custom backend config

EXAMPLES:
  $0 init               # Initialize Terraform
  $0 plan               # Preview changes
  $0 apply              # Deploy infrastructure
  $0 deploy-app         # Deploy application to instance
  $0 connect            # SSH to the instance

SETUP REQUIREMENTS:
1. Configure OCI credentials:
   - Create API key in OCI Console
   - Save private key to ~/.oci/oci_api_key.pem
   - Copy terraform.tfvars.example to terraform.tfvars
   - Update terraform.tfvars with your OCI details

2. Set up Terraform backend (optional):
   - Create Object Storage bucket in OCI
   - Copy backend.conf.example to backend.conf
   - Update backend.conf with your bucket details

EOF
}

check_prerequisites() {
    info "Checking prerequisites..."
    
    local missing=()
    
    command -v terraform >/dev/null 2>&1 || missing+=("terraform")
    command -v ssh >/dev/null 2>&1 || missing+=("ssh")
    command -v ssh-keygen >/dev/null 2>&1 || missing+=("ssh-keygen")
    
    if [ ${#missing[@]} -ne 0 ]; then
        error "Missing required tools: ${missing[*]}"
        echo ""
        echo "Install instructions:"
        echo "  terraform: https://www.terraform.io/downloads.html"
        echo "  ssh/ssh-keygen: Usually pre-installed on macOS/Linux"
        return 1
    fi
    
    # Check if tfvars file exists
    if [ ! -f "$TFVARS_FILE" ]; then
        error "terraform.tfvars file not found!"
        echo ""
        echo "Please create it by copying the example:"
        echo "  cp $TERRAFORM_DIR/terraform.tfvars.example $TFVARS_FILE"
        echo "  # Edit $TFVARS_FILE with your OCI credentials"
        return 1
    fi
    
    # Check if SSH key exists
    if [ ! -f ~/.ssh/tennis_booker_key ]; then
        warn "SSH key not found. Generating new key pair..."
        ssh-keygen -t rsa -b 4096 -f ~/.ssh/tennis_booker_key -N "" -C "tennis-booker-oci"
        success "SSH key generated: ~/.ssh/tennis_booker_key"
        echo ""
        warn "Please add the public key to your terraform.tfvars file:"
        echo "ssh_public_key = \"$(cat ~/.ssh/tennis_booker_key.pub)\""
    fi
    
    success "Prerequisites check passed"
}

terraform_init() {
    info "Initializing Terraform..."
    cd "$TERRAFORM_DIR"
    
    local backend_args=()
    if [ -f "$BACKEND_CONF" ]; then
        backend_args+=("-backend-config=$BACKEND_CONF")
        info "Using backend configuration: $BACKEND_CONF"
    else
        warn "No backend configuration found. State will be stored locally."
        warn "For production, create backend.conf for remote state storage."
    fi
    
    terraform init "${backend_args[@]}"
    success "Terraform initialized"
}

terraform_plan() {
    info "Planning Terraform changes..."
    cd "$TERRAFORM_DIR"
    
    terraform plan -var-file="$TFVARS_FILE" -out=tfplan
    success "Terraform plan completed"
}

terraform_apply() {
    info "Applying Terraform changes..."
    cd "$TERRAFORM_DIR"
    
    local apply_args=("-var-file=$TFVARS_FILE")
    
    if [ "$1" = "--auto-approve" ]; then
        apply_args+=("-auto-approve")
    fi
    
    terraform apply "${apply_args[@]}"
    success "Infrastructure deployed successfully!"
    
    # Show outputs
    echo ""
    info "Deployment outputs:"
    terraform output
}

terraform_destroy() {
    warn "This will destroy ALL infrastructure!"
    
    if [ "$1" != "--auto-approve" ]; then
        read -p "Are you sure you want to continue? (yes/no): " confirm
        if [ "$confirm" != "yes" ]; then
            info "Destroy cancelled"
            return 0
        fi
    fi
    
    info "Destroying infrastructure..."
    cd "$TERRAFORM_DIR"
    
    local destroy_args=("-var-file=$TFVARS_FILE")
    
    if [ "$1" = "--auto-approve" ]; then
        destroy_args+=("-auto-approve")
    fi
    
    terraform destroy "${destroy_args[@]}"
    success "Infrastructure destroyed"
}

show_outputs() {
    info "Terraform outputs:"
    cd "$TERRAFORM_DIR"
    terraform output
}

connect_instance() {
    info "Connecting to instance..."
    cd "$TERRAFORM_DIR"
    
    local public_ip
    public_ip=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    
    if [ -z "$public_ip" ]; then
        error "Could not get instance public IP. Is the infrastructure deployed?"
        return 1
    fi
    
    info "Connecting to $public_ip..."
    ssh -i ~/.ssh/tennis_booker_key -o StrictHostKeyChecking=no opc@"$public_ip"
}

deploy_application() {
    info "Deploying application to instance..."
    cd "$TERRAFORM_DIR"
    
    local public_ip
    public_ip=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    
    if [ -z "$public_ip" ]; then
        error "Could not get instance public IP. Is the infrastructure deployed?"
        return 1
    fi
    
    info "Deploying to $public_ip..."
    
    # Create deployment script
    cat > /tmp/deploy-tennis-booker.sh << 'DEPLOY_SCRIPT'
#!/bin/bash
set -e

info() { echo -e "\033[0;34mℹ️  $1\033[0m"; }
success() { echo -e "\033[0;32m✅ $1\033[0m"; }

info "Setting up Tennis Booker application..."

# Setup block volume if not already done
if [ ! -d "/opt/tennis-booker/data" ]; then
    info "Setting up block volume..."
    sudo /opt/tennis-booker/setup-volume.sh
fi

# Clone or update repository
if [ ! -d "/opt/tennis-booker/.git" ]; then
    info "Cloning repository..."
    cd /opt/tennis-booker
    # You'll need to replace this with your actual repository URL
    git clone https://github.com/yourusername/tennis-booker.git .
else
    info "Updating repository..."
    cd /opt/tennis-booker
    git pull origin main
fi

# Create environment file
if [ ! -f "/opt/tennis-booker/.env" ]; then
    info "Creating environment file..."
    cp .env.example .env
    # Update .env file with production values
    sed -i 's/ENVIRONMENT=development/ENVIRONMENT=production/' .env
    sed -i 's/localhost/tennis-booker.yourdomain.com/' .env
fi

# Deploy with Docker Compose
info "Deploying with Docker Compose..."
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml up -d --build

success "Tennis Booker deployed successfully!"

info "Checking service status..."
docker-compose -f docker-compose.prod.yml ps

info "Application logs (last 20 lines):"
docker-compose -f docker-compose.prod.yml logs --tail=20
DEPLOY_SCRIPT

    # Copy and run deployment script
    scp -i ~/.ssh/tennis_booker_key -o StrictHostKeyChecking=no \
        /tmp/deploy-tennis-booker.sh opc@"$public_ip":/tmp/
    
    ssh -i ~/.ssh/tennis_booker_key -o StrictHostKeyChecking=no opc@"$public_ip" \
        "chmod +x /tmp/deploy-tennis-booker.sh && /tmp/deploy-tennis-booker.sh"
    
    success "Application deployment completed!"
    
    # Clean up
    rm /tmp/deploy-tennis-booker.sh
    
    info "Application should be available at:"
    echo "  http://$public_ip (if no domain configured)"
    echo "  https://yourdomain.com (if domain configured in .env)"
}

check_status() {
    info "Checking deployment status..."
    cd "$TERRAFORM_DIR"
    
    local public_ip
    public_ip=$(terraform output -raw instance_public_ip 2>/dev/null || echo "")
    
    if [ -z "$public_ip" ]; then
        error "Infrastructure not deployed"
        return 1
    fi
    
    info "Instance: $public_ip"
    
    # Check if SSH is available
    if ssh -i ~/.ssh/tennis_booker_key -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
       opc@"$public_ip" "echo 'SSH connection successful'" 2>/dev/null; then
        success "SSH connection: OK"
        
        # Check Docker status
        if ssh -i ~/.ssh/tennis_booker_key -o StrictHostKeyChecking=no \
           opc@"$public_ip" "sudo systemctl is-active docker" 2>/dev/null | grep -q "active"; then
            success "Docker service: Running"
            
            # Check application containers
            ssh -i ~/.ssh/tennis_booker_key -o StrictHostKeyChecking=no \
                opc@"$public_ip" "cd /opt/tennis-booker && docker-compose -f docker-compose.prod.yml ps"
        else
            warn "Docker service: Not running"
        fi
    else
        error "SSH connection: Failed"
    fi
}

# Main command handling
main() {
    case "${1:-help}" in
        init)
            check_prerequisites
            terraform_init
            ;;
        plan)
            check_prerequisites
            terraform_plan
            ;;
        apply)
            check_prerequisites
            terraform_apply "$2"
            ;;
        destroy)
            terraform_destroy "$2"
            ;;
        output|outputs)
            show_outputs
            ;;
        connect|ssh)
            connect_instance
            ;;
        deploy-app)
            deploy_application
            ;;
        status)
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