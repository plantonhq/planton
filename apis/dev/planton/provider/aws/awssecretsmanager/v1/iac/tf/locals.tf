locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base tags
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "aws_secrets_manager"
  }

  # Organization tag only if var.metadata.org is non-empty
  org_tag = (
  var.metadata.org != null && var.metadata.org != ""
  ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment tag only if var.metadata.env is non-empty
  env_tag = (
  var.metadata.env != null &&
  try(var.metadata.env, "") != ""
  ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge base, org, and environment tags
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # List of secrets to manage
  secret_names = try(var.spec.secret_names, [])
}
