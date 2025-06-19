# OCI Authentication Variables
variable "tenancy_ocid" {
  description = "The OCID of the tenancy"
  type        = string
  sensitive   = true
}

variable "user_ocid" {
  description = "The OCID of the user"
  type        = string
  sensitive   = true
}

variable "fingerprint" {
  description = "The fingerprint of the public key"
  type        = string
  sensitive   = true
}

variable "private_key_path" {
  description = "The path to the private key file"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "The OCI region"
  type        = string
  default     = "us-ashburn-1"
}

# Project Configuration
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "tennis-booker"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "compartment_ocid" {
  description = "The OCID of the compartment (optional, defaults to tenancy root)"
  type        = string
  default     = ""
}

# Backend Configuration
variable "bucket_name" {
  description = "OCI Object Storage bucket name for Terraform state"
  type        = string
  default     = "terraform-state-bucket"
}

variable "namespace" {
  description = "OCI Object Storage namespace"
  type        = string
  default     = ""
}

# Network Configuration
variable "vcn_cidr" {
  description = "CIDR block for the VCN"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidr" {
  description = "CIDR block for the public subnet"
  type        = string
  default     = "10.0.1.0/24"
}

variable "private_subnet_cidr" {
  description = "CIDR block for the private subnet"
  type        = string
  default     = "10.0.2.0/24"
}

# Compute Configuration
variable "instance_shape" {
  description = "Shape of the compute instance"
  type        = string
  default     = "VM.Standard.A1.Flex"
}

variable "instance_ocpus" {
  description = "Number of OCPUs for the instance (Always Free: up to 4 total)"
  type        = number
  default     = 2
}

variable "instance_memory_in_gbs" {
  description = "Memory in GBs for the instance (Always Free: up to 24GB total)"
  type        = number
  default     = 12
}

variable "ssh_public_key" {
  description = "SSH public key for instance access"
  type        = string
  default     = ""
}

# Storage Configuration
variable "block_volume_size_in_gbs" {
  description = "Size of the block volume in GBs (Always Free: up to 200GB total)"
  type        = number
  default     = 50
} 