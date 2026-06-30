locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "gcp_dns_zone"
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

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Convert domain name into a valid Managed Zone name by replacing dots with hyphens
  managed_zone_name = replace(var.metadata.name, ".", "-")

  # dns_name must end with a dot
  zone_dns_name = "${var.metadata.name}."

  # Prepare IAM binding members (prefix each with "serviceAccount:")
  iam_binding_members = [
    for sa in var.spec.iam_service_accounts :
    "serviceAccount:${sa}"
  ]
}
