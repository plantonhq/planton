# locals.tf

locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # The resource name from metadata
  resource_name = var.metadata.name

  # Extract key_pair from StringValueOrRef (optional)
  key_pair = var.spec.key_pair != null ? var.spec.key_pair.value : null

  # Extract server_group_id from StringValueOrRef (optional)
  server_group_id = var.spec.server_group_id != null ? var.spec.server_group_id.value : null

  # Extract security group names from repeated StringValueOrRef
  security_groups = [for sg in var.spec.security_groups : sg.value]

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "openstack_instance"
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
}
