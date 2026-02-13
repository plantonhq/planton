# ScalewayBlockVolume Terraform Module
#
# Creates a Scaleway Block Storage volume with the specified performance
# tier (SBS 5K or SBS 15K) and size. Block volumes are raw block devices
# that persist independently of Instance lifecycle.
#
# This module wraps a single `scaleway_block_volume` resource.

resource "scaleway_block_volume" "volume" {
  name = local.volume_name
  iops = local.iops
  tags = local.standard_tags

  # Size in GB (5-10240). Can be increased in-place, cannot be shrunk.
  size_in_gb = local.size_gb

  # Zone is set via the provider configuration (from spec.zone).
  # The volume inherits the provider's zone automatically.

  # Optional: create volume from snapshot.
  snapshot_id = local.from_snapshot ? var.spec.snapshot_id : null

  lifecycle {
    # Block volumes contain user data -- prevent accidental destruction.
    # Set to true if you need to allow Terraform to destroy and recreate.
    prevent_destroy = false

    # Ignore IOPS changes only if explicitly marked as externally managed.
    # By default, we manage all fields via IaC.
  }
}
