# locals.tf

locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # The resource name from metadata (used for labeling, not passed to OpenStack)
  resource_name = var.metadata.name

  # Extract floating_ip from StringValueOrRef (required field)
  # This is typically an IP address like "203.0.113.42" (not a UUID)
  floating_ip = var.spec.floating_ip.value

  # Extract port_id from StringValueOrRef (required field)
  port_id = var.spec.port_id.value

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "openstack_floating_ip_associate"
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
