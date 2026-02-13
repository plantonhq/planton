# Create Azure Container Registry
resource "azurerm_container_registry" "acr" {
  name                = local.registry_name
  resource_group_name = local.resource_group_name
  location            = var.spec.region
  sku                 = local.sku_name
  admin_enabled       = var.spec.admin_user_enabled

  # Allow Azure services to bypass network rules
  network_rule_bypass_option = "AzureServices"

  tags = local.tags

  lifecycle {
    ignore_changes = [
      tags["LastModified"],
    ]
  }
}

# Create geo-replications for Premium SKU
resource "azurerm_container_registry_geo_replication" "replicas" {
  for_each = local.sku_name == "Premium" ? toset(var.spec.geo_replication_regions) : []

  container_registry_name = azurerm_container_registry.acr.name
  resource_group_name     = local.resource_group_name
  location                = each.value
  zone_redundancy_enabled = true

  tags = local.tags

  lifecycle {
    ignore_changes = [
      tags["LastModified"],
    ]
  }
}

