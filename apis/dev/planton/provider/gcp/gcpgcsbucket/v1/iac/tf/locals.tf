# locals.tf

locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.spec.bucket_name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_name" = var.spec.bucket_name
    "resource_kind" = "gcp_gcs_bucket"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null &&
    try(var.metadata.env, "") != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, env, and user-provided labels (user labels take precedence)
  final_labels = merge(local.base_labels, local.org_label, local.env_label, var.spec.gcp_labels)
}
