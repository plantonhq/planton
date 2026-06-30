locals {
  resource_id = (
  var.metadata.id != null && var.metadata.id != ""
  ) ? var.metadata.id : var.metadata.name

  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "gcp_artifact_registry_repo"
  }

  org_label = (
  var.metadata.org != null && var.metadata.org != ""
  ) ? {
    "organization" = var.metadata.org
  } : {}

  env_label = (
  var.metadata.env != null && var.metadata.env != ""
  ) ? {
    "organization" = var.metadata.env
  } : {}

  final_labels = merge(
    local.base_labels,
    local.org_label,
    local.env_label
  )

  # Direct spec values
  repo_format = var.spec.repo_format
  region      = var.spec.region
  enable_public_access = try(var.spec.enable_public_access, false)

  # Extract project_id from StringValueOrRef (value_from reference resolution not yet implemented)
  # If value is set, use it; otherwise, value_from references will result in empty string (not implemented)
  project_id = coalesce(
    try(var.spec.project_id.value, null),
    ""
  )
}
