# DigitalOcean Volume Resource
# This module creates a block storage volume in DigitalOcean with the specified configuration

resource "digitalocean_volume" "this" {
  # Volume identification
  name        = var.spec.volume_name
  description = local.volume_description

  # Region placement - volume can only attach to Droplets in the same region
  region = var.spec.region

  # Size in GiB (1-16000)
  size = var.spec.size_gib

  # Filesystem pre-formatting
  # If set, DigitalOcean will format the volume before making it available
  # Options: "" (none), "ext4", or "xfs"
  # Recommended: Use "xfs" for databases, "ext4" for general purpose
  initial_filesystem_type = local.filesystem_type

  # Optional: Create volume from snapshot
  # If snapshot_id is provided, the new volume will be created from the snapshot
  # The volume must be at least as large as the snapshot
  snapshot_id = local.from_snapshot ? var.spec.snapshot_id : null

  # Tags for organization and cost allocation
  tags = local.resource_tags

  # Lifecycle management
  lifecycle {
    # Prevent accidental destruction of volumes with data
    # Remove this if you need to force replacement
    prevent_destroy = false

    # Ignore changes to tags from external sources
    ignore_changes = [
      # Allow external tag modifications without Terraform drift
    ]
  }
}

