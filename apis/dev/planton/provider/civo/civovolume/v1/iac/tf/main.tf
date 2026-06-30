# Create Civo Volume
resource "civo_volume" "main" {
  name   = local.volume_name
  size_gb = local.size_gib
  region = local.region

  # Note: The Civo provider doesn't expose the following parameters:
  # - filesystem_type: Users must format the volume manually after creation
  #   or use cloud-init/configuration management to automate formatting
  # - snapshot_id: Snapshot functionality is not available on public Civo cloud
  # - tags: The Civo Volume provider doesn't currently support tags
}

# Informational output for limitations
# These are documented here for users who reference the Terraform module directly
# Filesystem type requested: ${local.filesystem_type}
# Snapshot ID specified: ${local.snapshot_id}
# Tags specified: ${join(", ", local.tags)}

