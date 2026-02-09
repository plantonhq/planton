# main.tf

# Create the OpenStack Compute instance.
resource "openstack_compute_instance_v2" "main" {
  name = var.metadata.name

  # Flavor (exactly one of flavor_name/flavor_id set, enforced by proto validation).
  flavor_name = var.spec.flavor_name != "" ? var.spec.flavor_name : null
  flavor_id   = var.spec.flavor_id != "" ? var.spec.flavor_id : null

  # Image (optional -- not needed when booting from block_device).
  image_name = var.spec.image_name != "" ? var.spec.image_name : null
  image_id   = var.spec.image_id != "" ? var.spec.image_id : null

  # SSH keypair (optional).
  key_pair = local.key_pair

  # Network attachments (at least one, enforced by proto validation).
  dynamic "network" {
    for_each = var.spec.networks
    content {
      uuid        = network.value.uuid != null ? network.value.uuid.value : null
      port        = network.value.port != null ? network.value.port.value : null
      fixed_ip_v4 = network.value.fixed_ip_v4 != "" ? network.value.fixed_ip_v4 : null
      access_network = network.value.access_network
    }
  }

  # Security groups (optional, resolved names).
  security_groups = length(local.security_groups) > 0 ? toset(local.security_groups) : null

  # Block device mappings (optional).
  dynamic "block_device" {
    for_each = var.spec.block_device
    content {
      source_type           = block_device.value.source_type
      uuid                  = block_device.value.uuid != "" ? block_device.value.uuid : null
      destination_type      = block_device.value.destination_type != "" ? block_device.value.destination_type : null
      boot_index            = block_device.value.boot_index
      volume_size           = block_device.value.volume_size > 0 ? block_device.value.volume_size : null
      delete_on_termination = block_device.value.delete_on_termination
      volume_type           = block_device.value.volume_type != "" ? block_device.value.volume_type : null
    }
  }

  # User data (optional, ForceNew).
  user_data = var.spec.user_data != "" ? var.spec.user_data : null

  # Metadata (optional).
  metadata = length(var.spec.metadata) > 0 ? var.spec.metadata : null

  # Config drive (optional, ForceNew).
  config_drive = var.spec.config_drive

  # Server group via scheduler_hints (optional, ForceNew).
  dynamic "scheduler_hints" {
    for_each = local.server_group_id != null ? [local.server_group_id] : []
    content {
      group = scheduler_hints.value
    }
  }

  # Availability zone (optional, ForceNew).
  availability_zone = var.spec.availability_zone != "" ? var.spec.availability_zone : null

  # Tags (optional).
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
