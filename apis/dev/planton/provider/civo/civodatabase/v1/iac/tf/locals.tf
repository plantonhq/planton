# Local values for Civo Database Terraform Module

locals {
  # Resource ID (prefer metadata.id if set, otherwise use metadata.name)
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  # Base labels for resource tagging
  base_labels = {
    "planton:resource"      = "true"
    "planton:resource_id"   = local.resource_id
    "planton:resource_name" = var.metadata.name
    "planton:resource_kind" = "civo_database"
  }

  # Organization label (if provided)
  org_label = var.metadata.org != null && var.metadata.org != "" ? {
    "planton:organization" = var.metadata.org
  } : {}

  # Environment label (if provided)
  env_label = var.metadata.env != null && var.metadata.env != "" ? {
    "planton:environment" = var.metadata.env
  } : {}

  # Merge all labels
  final_labels = merge(
    local.base_labels,
    local.org_label,
    local.env_label,
    var.metadata.labels != null ? var.metadata.labels : {}
  )

  # Extract network ID from literal value or reference
  network_id = try(var.spec.network_id.value, "")

  # Extract first firewall ID if provided (Civo currently supports one firewall per database)
  firewall_id = length(var.spec.firewall_ids) > 0 ? try(var.spec.firewall_ids[0].value, "") : ""

  # Combine metadata tags and spec tags
  all_tags = concat(
    var.metadata.tags != null ? var.metadata.tags : [],
    var.spec.tags
  )

  # Total nodes = primary + replicas
  total_nodes = var.spec.replicas + 1
}

