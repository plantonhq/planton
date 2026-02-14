locals {
  resource_id = var.metadata.id != null ? var.metadata.id : var.metadata.name

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azurelinuxwebapp"
    "resource_name" = var.metadata.name
  }

  org_tag = var.metadata.org != null ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != null ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Merge Application Insights connection string into app settings
  ai_settings = var.spec.application_insights_connection_string != null ? {
    APPLICATIONINSIGHTS_CONNECTION_STRING = var.spec.application_insights_connection_string
  } : {}

  merged_app_settings = merge(
    var.spec.app_settings,
    local.ai_settings,
  )
}
