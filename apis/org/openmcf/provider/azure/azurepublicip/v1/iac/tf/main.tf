# Create the Azure Public IP Address.
# SKU is always Standard (Basic was retired Sept 2025).
# Allocation is always Static (Standard SKU requires it).
resource "azurerm_public_ip" "main" {
  name                    = var.spec.name
  location                = var.spec.region
  resource_group_name     = var.spec.resource_group
  allocation_method       = "Static"
  sku                     = "Standard"
  domain_name_label       = var.spec.domain_name_label
  zones                   = var.spec.zones
  idle_timeout_in_minutes = var.spec.idle_timeout_in_minutes
  tags                    = local.final_tags
}
