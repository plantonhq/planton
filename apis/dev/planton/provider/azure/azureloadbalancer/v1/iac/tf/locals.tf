locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base tags for Azure resources
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_load_balancer"
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

  # Merge base, org, and environment tags
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Auto-derive frontend IP configuration name
  frontend_config_name = "${var.spec.name}-frontend"

  # Determine if this is an internal LB (subnet_id set, no public_ip_id)
  is_internal = var.spec.subnet_id != null && var.spec.subnet_id != "" && (var.spec.public_ip_id == null || var.spec.public_ip_id == "")

  # Build backend pool map for rule references
  backend_pool_map = { for pool in azurerm_lb_backend_address_pool.pools : pool.name => pool.id }

  # Build probe map for rule references
  probe_map = { for probe in azurerm_lb_probe.probes : probe.name => probe.id }
}
