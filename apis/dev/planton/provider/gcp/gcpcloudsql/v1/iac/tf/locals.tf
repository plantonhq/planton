locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Extract project_id from StringValueOrRef
  # Note: value_from resolution is not yet implemented - only literal values are supported
  project_id = (
    var.spec.project_id != null
    ? coalesce(var.spec.project_id.value, "")
    : ""
  )

  # Extract vpc_id from StringValueOrRef
  # Note: value_from resolution is not yet implemented - only literal values are supported
  vpc_id = (
    var.spec.network != null && var.spec.network.vpc_id != null
    ? coalesce(var.spec.network.vpc_id.value, "")
    : ""
  )

  # Base GCP labels
  base_gcp_labels = {
    "resource"      = "true"
    "resource_kind" = "gcp-cloud-sql"
    "resource_name" = var.metadata.name
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge base, org, environment labels and add resource_id
  final_gcp_labels = merge(
    local.base_gcp_labels,
    { "resource_id" = local.resource_id },
    local.org_label,
    local.env_label
  )

  # Determine availability type based on high availability setting
  availability_type = (
    var.spec.high_availability != null && var.spec.high_availability.enabled
    ? "REGIONAL"
    : "ZONAL"
  )

  # Determine if backup is enabled
  backup_enabled = (
    var.spec.backup != null && var.spec.backup.enabled
  )

  # Determine if private IP is enabled
  private_ip_enabled = (
    var.spec.network != null && var.spec.network.private_ip_enabled
  )

  # Convert database flags to list format for Terraform
  database_flags_list = [
    for name, value in var.spec.database_flags : {
      name  = name
      value = value
    }
  ]
}

