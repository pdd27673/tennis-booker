# Tennis Booker OCI Infrastructure

This directory contains Terraform configuration for provisioning Oracle Cloud Infrastructure (OCI) resources for the Tennis Booker application using the Always Free tier.

## Architecture Overview

The infrastructure includes:
- **Virtual Cloud Network (VCN)** with public and private subnets
- **Internet Gateway** for public internet access
- **NAT Gateway** for private subnet outbound connectivity
- **Security Lists** with appropriate ingress/egress rules
- **ARM Ampere A1 Compute Instance** (Always Free eligible)
- **Block Volume** for persistent storage

## Always Free Tier Limits

This configuration is designed to stay within OCI Always Free tier limits:
- **Compute**: Up to 4 OCPUs and 24GB RAM total (we use 2 OCPUs, 12GB RAM)
- **Block Storage**: Up to 200GB total (we use 50GB)
- **Network**: 1 VCN, unlimited subnets, gateways, and security lists
- **Load Balancer**: 1 flexible load balancer (10 Mbps)

## Prerequisites

### 1. OCI Account Setup
1. Create an OCI account at [cloud.oracle.com](https://cloud.oracle.com)
2. Complete the account verification process
3. Note your tenancy OCID from the OCI Console

### 2. API Key Setup
1. Generate an API key pair:
   ```bash
   mkdir -p ~/.oci
   openssl genrsa -out ~/.oci/oci_api_key.pem 2048
   openssl rsa -pubout -in ~/.oci/oci_api_key.pem -out ~/.oci/oci_api_key_public.pem
   chmod 600 ~/.oci/oci_api_key.pem
   ```

2. Add the public key to your OCI user:
   - Go to OCI Console → Identity & Security → Users
   - Click your username → API Keys → Add API Key
   - Upload `~/.oci/oci_api_key_public.pem`
   - Note the fingerprint displayed

3. Gather required OCIDs:
   - **Tenancy OCID**: OCI Console → Administration → Tenancy Details
   - **User OCID**: OCI Console → Identity & Security → Users → Your User
   - **Compartment OCID**: (Optional) Create a compartment or use root

### 3. SSH Key Setup
Generate an SSH key pair for instance access:
```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/tennis_booker_key
```

### 4. Terraform Installation
Install Terraform (version >= 1.0):
```bash
# macOS
brew install terraform

# Linux
wget https://releases.hashicorp.com/terraform/1.5.7/terraform_1.5.7_linux_amd64.zip
unzip terraform_1.5.7_linux_amd64.zip
sudo mv terraform /usr/local/bin/
```

## Configuration

### 1. Create Configuration Files
```bash
# Copy example files
cp terraform.tfvars.example terraform.tfvars
cp backend.conf.example backend.conf
```

### 2. Configure terraform.tfvars
Edit `terraform.tfvars` with your actual values:
```hcl
tenancy_ocid     = "ocid1.tenancy.oc1..your-actual-tenancy-ocid"
user_ocid        = "ocid1.user.oc1..your-actual-user-ocid"
fingerprint      = "your-actual-fingerprint"
private_key_path = "~/.oci/oci_api_key.pem"
region           = "us-ashburn-1"
ssh_public_key   = "ssh-rsa AAAAB3NzaC1yc2E... your-actual-public-key"
```

### 3. Create OCI Object Storage Bucket
Before running Terraform, create a bucket for state storage:
1. Go to OCI Console → Storage → Buckets
2. Create bucket named `tennis-booker-terraform-state`
3. Enable versioning and encryption
4. Note the namespace (usually your tenancy name)

### 4. Configure backend.conf
Edit `backend.conf` with your bucket details:
```hcl
bucket    = "tennis-booker-terraform-state"
namespace = "your-actual-namespace"
```

## Deployment

### 1. Initialize Terraform
```bash
terraform init -backend-config=backend.conf
```

### 2. Plan Deployment
```bash
terraform plan
```

Review the planned changes carefully. You should see:
- 1 VCN with 2 subnets
- 1 Internet Gateway and 1 NAT Gateway
- 2 Route Tables and 2 Security Lists
- Various data sources for tenancy/compartment info

### 3. Apply Configuration
```bash
terraform apply
```

Type `yes` when prompted to confirm the deployment.

### 4. Verify Deployment
1. Check OCI Console for created resources
2. Note the public IP of the instance from Terraform outputs
3. Test SSH access:
   ```bash
   ssh -i ~/.ssh/tennis_booker_key opc@<instance-public-ip>
   ```

## Outputs

After successful deployment, Terraform will output:
- **tenancy_id**: Your tenancy OCID
- **compartment_id**: Compartment being used
- **vcn_id**: Created VCN OCID
- **public_subnet_id**: Public subnet OCID
- **private_subnet_id**: Private subnet OCID
- **availability_domains**: Available ADs in the region

## File Structure

```
infrastructure/terraform/
├── main.tf                    # Main configuration with resources
├── variables.tf               # Variable definitions
├── outputs.tf                 # Output definitions
├── terraform.tfvars.example   # Example configuration
├── backend.conf.example       # Example backend config
├── terraform.tfvars          # Your actual config (gitignored)
├── backend.conf              # Your actual backend config (gitignored)
└── README.md                 # This file
```

## Security Considerations

1. **Never commit sensitive files**:
   - `terraform.tfvars` (contains OCIDs and paths)
   - `backend.conf` (contains bucket details)
   - `*.pem` files (private keys)

2. **Use least privilege**:
   - Create a dedicated compartment for resources
   - Use IAM policies to limit access

3. **Enable encryption**:
   - Terraform state is encrypted in Object Storage
   - Block volumes use encryption by default

## Troubleshooting

### Common Issues

1. **Authentication Errors**:
   - Verify OCIDs are correct
   - Check API key fingerprint
   - Ensure private key path is correct
   - Verify user has necessary permissions

2. **Resource Limits**:
   - Check Always Free tier usage in OCI Console
   - Verify you haven't exceeded OCPU/memory limits
   - Ensure you're in a supported region

3. **Network Connectivity**:
   - Verify security list rules allow SSH (port 22)
   - Check route table configurations
   - Ensure Internet Gateway is attached

### Useful Commands

```bash
# Format Terraform files
terraform fmt

# Validate configuration
terraform validate

# Show current state
terraform show

# List resources
terraform state list

# Destroy infrastructure (careful!)
terraform destroy
```

## Next Steps

After successful infrastructure deployment:
1. Configure the compute instance with Docker
2. Set up application deployment scripts
3. Configure monitoring and logging
4. Implement backup strategies
5. Set up CI/CD pipelines for automated deployment

## Support

For issues related to:
- **OCI**: Check [OCI Documentation](https://docs.oracle.com/en-us/iaas/)
- **Terraform**: Check [Terraform OCI Provider](https://registry.terraform.io/providers/hashicorp/oci/latest/docs)
- **Always Free**: Check [Always Free Resources](https://docs.oracle.com/en-us/iaas/Content/FreeTier/freetier_topic-Always_Free_Resources.htm) 