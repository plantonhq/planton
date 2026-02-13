locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_application_gateway"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Auto-derived internal names
  gateway_ip_config_name  = "${var.spec.name}-gw-ip-config"
  frontend_ip_config_name = "${var.spec.name}-frontend-ip-config"

  # Create a map of frontend port name -> port for listeners
  frontend_ports = { for listener in var.spec.http_listeners : "${listener.name}-port" => listener.port }

  # Determine if autoscale is configured
  use_autoscale = var.spec.autoscale != null
}
