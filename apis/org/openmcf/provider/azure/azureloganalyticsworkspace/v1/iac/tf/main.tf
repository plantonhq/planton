# Create the Azure Log Analytics Workspace
resource "azurerm_log_analytics_workspace" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  sku                 = var.spec.sku
  retention_in_days   = var.spec.retention_in_days
  daily_quota_gb      = var.spec.daily_quota_gb
  tags                = local.final_tags
}
