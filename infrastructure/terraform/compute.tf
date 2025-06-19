# Data source for ARM-based images
data "oci_core_images" "arm_images" {
  compartment_id           = local.compartment_id
  operating_system         = "Oracle Linux"
  operating_system_version = "8"
  shape                    = var.instance_shape
  sort_by                  = "TIMECREATED"
  sort_order               = "DESC"
}

# Compute Instance
resource "oci_core_instance" "tennis_booker_instance" {
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  compartment_id      = local.compartment_id
  display_name        = "${var.project_name}-instance"
  shape               = var.instance_shape

  shape_config {
    ocpus         = var.instance_ocpus
    memory_in_gbs = var.instance_memory_in_gbs
  }

  create_vnic_details {
    subnet_id                 = oci_core_subnet.public_subnet.id
    display_name              = "${var.project_name}-vnic"
    assign_public_ip          = true
    assign_private_dns_record = true
    hostname_label            = "tennisbooker"
  }

  source_details {
    source_type             = "image"
    source_id               = data.oci_core_images.arm_images.images[0].id
    boot_volume_size_in_gbs = 50
  }

  metadata = {
    ssh_authorized_keys = var.ssh_public_key
    user_data = base64encode(templatefile("${path.module}/cloud-init.yaml", {
      project_name = var.project_name
    }))
  }

  freeform_tags = local.common_tags

  lifecycle {
    ignore_changes = [source_details[0].source_id]
  }
}

# Block Volume for persistent storage
resource "oci_core_volume" "tennis_booker_volume" {
  availability_domain = data.oci_identity_availability_domains.ads.availability_domains[0].name
  compartment_id      = local.compartment_id
  display_name        = "${var.project_name}-volume"
  size_in_gbs         = var.block_volume_size_in_gbs

  freeform_tags = local.common_tags
}

# Volume Attachment
resource "oci_core_volume_attachment" "tennis_booker_volume_attachment" {
  attachment_type = "iscsi"
  instance_id     = oci_core_instance.tennis_booker_instance.id
  volume_id       = oci_core_volume.tennis_booker_volume.id
  display_name    = "${var.project_name}-volume-attachment"

  # Use consistent device path
  device = "/dev/oracleoci/oraclevdb"
} 