# Tenancy and Compartment Information
output "tenancy_id" {
  description = "The OCID of the tenancy"
  value       = data.oci_identity_tenancy.tenancy.id
}

output "tenancy_name" {
  description = "The name of the tenancy"
  value       = data.oci_identity_tenancy.tenancy.name
}

output "compartment_id" {
  description = "The OCID of the compartment being used"
  value       = local.compartment_id
}

output "compartment_name" {
  description = "The name of the compartment being used"
  value       = data.oci_identity_compartment.compartment.name
}

# Availability Domains
output "availability_domains" {
  description = "List of availability domains in the region"
  value       = data.oci_identity_availability_domains.ads.availability_domains[*].name
}

# Region Information
output "region" {
  description = "The OCI region"
  value       = var.region
}

# Project Information
output "project_name" {
  description = "The project name"
  value       = var.project_name
}

output "environment" {
  description = "The environment"
  value       = var.environment
}

# Common Tags
output "common_tags" {
  description = "Common tags applied to resources"
  value       = local.common_tags
}

# Network Information
output "vcn_id" {
  description = "The OCID of the VCN"
  value       = oci_core_vcn.tennis_booker_vcn.id
}

output "vcn_cidr" {
  description = "The CIDR block of the VCN"
  value       = var.vcn_cidr
}

output "public_subnet_id" {
  description = "The OCID of the public subnet"
  value       = oci_core_subnet.public_subnet.id
}

output "private_subnet_id" {
  description = "The OCID of the private subnet"
  value       = oci_core_subnet.private_subnet.id
}

output "internet_gateway_id" {
  description = "The OCID of the Internet Gateway"
  value       = oci_core_internet_gateway.tennis_booker_igw.id
}

output "nat_gateway_id" {
  description = "The OCID of the NAT Gateway"
  value       = oci_core_nat_gateway.tennis_booker_nat.id
}

# Security Lists
output "public_security_list_id" {
  description = "The OCID of the public security list"
  value       = oci_core_security_list.public_security_list.id
}

output "private_security_list_id" {
  description = "The OCID of the private security list"
  value       = oci_core_security_list.private_security_list.id
}

# Compute Instance Information
output "instance_id" {
  description = "The OCID of the compute instance"
  value       = oci_core_instance.tennis_booker_instance.id
}

output "instance_public_ip" {
  description = "The public IP address of the compute instance"
  value       = oci_core_instance.tennis_booker_instance.public_ip
}

output "instance_private_ip" {
  description = "The private IP address of the compute instance"
  value       = oci_core_instance.tennis_booker_instance.private_ip
}

output "instance_state" {
  description = "The current state of the compute instance"
  value       = oci_core_instance.tennis_booker_instance.state
}

# Block Volume Information
output "block_volume_id" {
  description = "The OCID of the block volume"
  value       = oci_core_volume.tennis_booker_volume.id
}

output "block_volume_size" {
  description = "The size of the block volume in GBs"
  value       = oci_core_volume.tennis_booker_volume.size_in_gbs
}

output "volume_attachment_id" {
  description = "The OCID of the volume attachment"
  value       = oci_core_volume_attachment.tennis_booker_volume_attachment.id
}

# SSH Connection Information
output "ssh_connection_command" {
  description = "SSH command to connect to the instance"
  value       = "ssh -i ~/.ssh/tennis_booker_key opc@${oci_core_instance.tennis_booker_instance.public_ip}"
}

# Application URLs (for future use)
output "application_url" {
  description = "URL to access the application (once deployed)"
  value       = "http://${oci_core_instance.tennis_booker_instance.public_ip}"
}

output "application_url_https" {
  description = "HTTPS URL to access the application (once SSL is configured)"
  value       = "https://${oci_core_instance.tennis_booker_instance.public_ip}"
} 