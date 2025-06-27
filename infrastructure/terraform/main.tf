terraform {
  required_version = ">= 1.0"
  required_providers {
    oci = {
      source  = "hashicorp/oci"
      version = "~> 5.30.0"
    }
  }

  # backend "oci" {
  #   # Configuration will be provided via backend.conf file
  #   # bucket = "terraform-state-bucket"
  #   # namespace = "your-namespace"
  #   # key = "tennis-booker/dev/terraform.tfstate"
  #   # region = "us-ashburn-1"
  #   # encrypt = true
  # }
  
  # Using local backend for now
}

# Configure the OCI Provider
provider "oci" {
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  region           = var.region
}

# Local values for common configurations
locals {
  compartment_id = var.compartment_ocid != "" ? var.compartment_ocid : data.oci_identity_tenancy.tenancy.id

  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
    CreatedOn   = formatdate("YYYY-MM-DD", timestamp())
  }
}

# Data sources
data "oci_identity_tenancy" "tenancy" {
  tenancy_id = var.tenancy_ocid
}

data "oci_identity_availability_domains" "ads" {
  compartment_id = local.compartment_id
}

data "oci_identity_compartment" "compartment" {
  id = local.compartment_id
}

# Virtual Cloud Network (VCN)
resource "oci_core_vcn" "tennis_booker_vcn" {
  compartment_id = local.compartment_id
  cidr_blocks    = [var.vcn_cidr]
  display_name   = "${var.project_name}-vcn"
  dns_label      = "tennisbooker"

  freeform_tags = local.common_tags
}

# Internet Gateway
resource "oci_core_internet_gateway" "tennis_booker_igw" {
  compartment_id = local.compartment_id
  vcn_id         = oci_core_vcn.tennis_booker_vcn.id
  display_name   = "${var.project_name}-igw"
  enabled        = true

  freeform_tags = local.common_tags
}

# NAT Gateway
resource "oci_core_nat_gateway" "tennis_booker_nat" {
  compartment_id = local.compartment_id
  vcn_id         = oci_core_vcn.tennis_booker_vcn.id
  display_name   = "${var.project_name}-nat"
  block_traffic  = false

  freeform_tags = local.common_tags
}

# Route Table for Public Subnet
resource "oci_core_route_table" "public_route_table" {
  compartment_id = local.compartment_id
  vcn_id         = oci_core_vcn.tennis_booker_vcn.id
  display_name   = "${var.project_name}-public-rt"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_internet_gateway.tennis_booker_igw.id
  }

  freeform_tags = local.common_tags
}

# Route Table for Private Subnet
resource "oci_core_route_table" "private_route_table" {
  compartment_id = local.compartment_id
  vcn_id         = oci_core_vcn.tennis_booker_vcn.id
  display_name   = "${var.project_name}-private-rt"

  route_rules {
    destination       = "0.0.0.0/0"
    destination_type  = "CIDR_BLOCK"
    network_entity_id = oci_core_nat_gateway.tennis_booker_nat.id
  }

  freeform_tags = local.common_tags
}

# Public Subnet
resource "oci_core_subnet" "public_subnet" {
  compartment_id             = local.compartment_id
  vcn_id                     = oci_core_vcn.tennis_booker_vcn.id
  cidr_block                 = var.public_subnet_cidr
  display_name               = "${var.project_name}-public-subnet"
  dns_label                  = "public"
  route_table_id             = oci_core_route_table.public_route_table.id
  security_list_ids          = [oci_core_security_list.public_security_list.id]
  prohibit_public_ip_on_vnic = false

  freeform_tags = local.common_tags
}

# Private Subnet
resource "oci_core_subnet" "private_subnet" {
  compartment_id             = local.compartment_id
  vcn_id                     = oci_core_vcn.tennis_booker_vcn.id
  cidr_block                 = var.private_subnet_cidr
  display_name               = "${var.project_name}-private-subnet"
  dns_label                  = "private"
  route_table_id             = oci_core_route_table.private_route_table.id
  security_list_ids          = [oci_core_security_list.private_security_list.id]
  prohibit_public_ip_on_vnic = true

  freeform_tags = local.common_tags
}

# Security List for Public Subnet
resource "oci_core_security_list" "public_security_list" {
  compartment_id = local.compartment_id
  vcn_id         = oci_core_vcn.tennis_booker_vcn.id
  display_name   = "${var.project_name}-public-sl"

  # Ingress Rules
  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"

    tcp_options {
      min = 22
      max = 22
    }
    description = "SSH access"
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"

    tcp_options {
      min = 80
      max = 80
    }
    description = "HTTP access"
  }

  ingress_security_rules {
    protocol = "6" # TCP
    source   = "0.0.0.0/0"

    tcp_options {
      min = 443
      max = 443
    }
    description = "HTTPS access"
  }

  # Allow ICMP for ping
  ingress_security_rules {
    protocol    = "1" # ICMP
    source      = "0.0.0.0/0"
    description = "ICMP ping"
  }

  # Egress Rules
  egress_security_rules {
    protocol    = "all"
    destination = "0.0.0.0/0"
    description = "All outbound traffic"
  }

  freeform_tags = local.common_tags
}

# Security List for Private Subnet
resource "oci_core_security_list" "private_security_list" {
  compartment_id = local.compartment_id
  vcn_id         = oci_core_vcn.tennis_booker_vcn.id
  display_name   = "${var.project_name}-private-sl"

  # Ingress Rules - Allow traffic from VCN
  ingress_security_rules {
    protocol    = "all"
    source      = var.vcn_cidr
    description = "All traffic from VCN"
  }

  # Egress Rules
  egress_security_rules {
    protocol    = "all"
    destination = "0.0.0.0/0"
    description = "All outbound traffic"
  }

  freeform_tags = local.common_tags
} 