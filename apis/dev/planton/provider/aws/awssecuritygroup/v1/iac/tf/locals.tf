locals {
  ###########################################################################
  # Common resource metadata logic (consistent with other modules)
  ###########################################################################

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
    "resource_kind" = "aws_security_group"
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

  # Merge base, org, and environment labels into final_labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  ###########################################################################
  # Security Group specific logic
  ###########################################################################

  # Security Group name
  security_group_name = var.metadata.name

  # Description with explicit value
  description = var.spec.description

  # VPC ID
  vpc_id = var.spec.vpc_id
}

