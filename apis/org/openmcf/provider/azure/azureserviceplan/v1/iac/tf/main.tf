# Create the Azure App Service Plan.
resource "azurerm_service_plan" "main" {
  name                         = var.spec.name
  location                     = var.spec.region
  resource_group_name          = var.spec.resource_group
  os_type                      = var.spec.os_type
  sku_name                     = var.spec.sku_name
  worker_count                 = var.spec.worker_count
  zone_balancing_enabled       = var.spec.zone_balancing_enabled
  per_site_scaling_enabled     = var.spec.per_site_scaling_enabled
  maximum_elastic_worker_count = var.spec.maximum_elastic_worker_count
  tags                         = local.final_tags
}
