# Create the Azure Application Insights resource
resource "azurerm_application_insights" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  application_type    = var.spec.application_type
  workspace_id        = var.spec.workspace_id
  retention_in_days   = var.spec.retention_in_days
  daily_data_cap_in_gb = var.spec.daily_data_cap_in_gb
  sampling_percentage = var.spec.sampling_percentage
  tags                = local.final_tags
}
