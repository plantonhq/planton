# ---------------------------------------------------------------------------
# Tags and Resource Naming
# ---------------------------------------------------------------------------

locals {
  resource_name = coalesce(var.resource_name, "awselasticfilesystem")

  tags = merge({
    "Name" = local.resource_name
  }, var.labels)

  # ---------------------------------------------------------------------------
  # Mount Targets — set of subnet IDs for for_each
  # ---------------------------------------------------------------------------
  subnet_ids_set = toset(var.subnet_ids)

  # ---------------------------------------------------------------------------
  # Access Points — map keyed by name for for_each
  # ---------------------------------------------------------------------------
  access_point_map = { for ap in var.access_points : ap.name => ap }

  # ---------------------------------------------------------------------------
  # Lifecycle Policies — list of policy objects for dynamic block
  # Each lifecycle_policy block supports one transition type.
  # ---------------------------------------------------------------------------
  lifecycle_policies = concat(
    var.transition_to_ia != "" ? [{
      transition_to_ia = var.transition_to_ia
    }] : [],
    var.transition_to_archive != "" ? [{
      transition_to_archive = var.transition_to_archive
    }] : [],
    var.transition_to_primary_storage_class != "" ? [{
      transition_to_primary_storage_class = var.transition_to_primary_storage_class
    }] : []
  )

  # ---------------------------------------------------------------------------
  # KMS key — null when not specified (use AWS-managed key)
  # ---------------------------------------------------------------------------
  kms_key_id = var.kms_key_id != "" ? var.kms_key_id : null

  # ---------------------------------------------------------------------------
  # Availability zone — null for regional (multi-AZ) file systems
  # ---------------------------------------------------------------------------
  availability_zone_name = var.availability_zone_name != "" ? var.availability_zone_name : null

  # ---------------------------------------------------------------------------
  # Provisioned throughput — only when throughput_mode = provisioned
  # ---------------------------------------------------------------------------
  provisioned_throughput_in_mibps = var.throughput_mode == "provisioned" && var.provisioned_throughput_in_mibps > 0 ? var.provisioned_throughput_in_mibps : null
}
