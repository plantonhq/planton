locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Generate NAT Gateway name
  nat_gateway_name = "natgw-${var.metadata.name}"

  # Resource group and region are explicit spec fields
  resource_group = var.spec.resource_group
  location       = var.spec.region

  # Base tags for Azure resources
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_nat_gateway"
    "resource_name" = var.metadata.name
  }

  # Organization tag only if var.metadata.org is non-empty
  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment tag only if var.metadata.env is non-empty
  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, environment, and user-provided tags
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.spec.tags)

  # Determine if we're creating a prefix or individual IP
  use_ip_prefix = var.spec.public_ip_prefix_length != null && var.spec.public_ip_prefix_length > 0
}

