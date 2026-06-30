locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels for all resources
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "gcp_subnetwork"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ? { "organization" = var.metadata.org }
    : {}
  )

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ? { "environment" = var.metadata.env }
    : {}
  )

  # Merge base, org, and environment labels with user-provided labels
  final_labels = merge(
    local.base_labels,
    local.org_label,
    local.env_label,
    var.metadata.labels != null ? var.metadata.labels : {}
  )

  # Determine whether to enable Private Google Access (defaults to false)
  private_ip_google_access = coalesce(var.spec.private_ip_google_access, false)

  # Process secondary IP ranges (filter null/empty)
  secondary_ip_ranges = var.spec.secondary_ip_ranges != null ? [
    for range in var.spec.secondary_ip_ranges : range
    if range.range_name != "" && range.ip_cidr_range != ""
  ] : []
}

