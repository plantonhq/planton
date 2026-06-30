# Local variables for computed values and helper functions

locals {
  # Filesystem type mapping from protobuf enum to DigitalOcean API format
  # Protobuf: NONE (0), EXT4 (1), XFS (2)
  # DigitalOcean API expects lowercase strings: "", "ext4", "xfs"
  filesystem_type_map = {
    "NONE" = ""
    "EXT4" = "ext4"
    "XFS"  = "xfs"
  }

  # Convert filesystem type from spec to DigitalOcean format
  filesystem_type = lookup(local.filesystem_type_map, var.spec.filesystem_type, "")

  # Determine if volume should be created from snapshot
  from_snapshot = var.spec.snapshot_id != null && var.spec.snapshot_id != ""

  # Common tags/labels for the volume
  # Note: DigitalOcean volumes support tags which are used for organization
  volume_tags = concat(
    var.spec.tags,
    var.metadata.labels != null ? [
      for k, v in var.metadata.labels : "${k}:${v}"
    ] : []
  )

  # Additional metadata tags
  resource_tags = concat(
    local.volume_tags,
    [
      "managed-by:terraform",
      "resource:digitalocean-volume"
    ]
  )

  # Description with metadata
  volume_description = var.spec.description != "" ? var.spec.description : "DigitalOcean Volume managed by Terraform"
}

