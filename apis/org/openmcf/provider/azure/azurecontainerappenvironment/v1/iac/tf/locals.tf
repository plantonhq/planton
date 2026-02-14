locals {
  resource_id = var.metadata.id != null ? var.metadata.id : var.metadata.name

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_container_app_environment"
    "resource_name" = var.metadata.name
  }

  org_tag = var.metadata.org != null ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != null ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Auto-derive logs_destination: "log-analytics" when workspace provided, else null
  logs_destination = var.spec.log_analytics_workspace_id != null ? "log-analytics" : null
}
