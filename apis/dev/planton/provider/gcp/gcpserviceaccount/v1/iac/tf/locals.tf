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
    "resource_kind" = "gcp_service_account"
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

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Computed service account email (used for IAM bindings)
  service_account_email = google_service_account.main.email

  # Determine whether to create a key (defaults to false if not specified)
  create_key = coalesce(var.spec.create_key, false)

  # Project IAM roles (filter empty strings)
  project_iam_roles = var.spec.project_iam_roles != null ? [
    for role in var.spec.project_iam_roles : role
    if role != ""
  ] : []

  # Organization IAM roles (filter empty strings)
  org_iam_roles = var.spec.org_iam_roles != null ? [
    for role in var.spec.org_iam_roles : role
    if role != ""
  ] : []
}

