# Private DNS Zone
# Global resource (no location parameter) for private name resolution within VNets.
resource "azurerm_private_dns_zone" "zone" {
  name                = var.spec.name
  resource_group_name = var.spec.resource_group
  tags                = local.final_tags
}

# Virtual Network Link
# Links the private DNS zone to a VNet for DNS resolution.
# A zone without a VNet link is unreachable (DD03 bundling).
resource "azurerm_private_dns_zone_virtual_network_link" "vnet_link" {
  name                  = local.vnet_link_name
  resource_group_name   = var.spec.resource_group
  private_dns_zone_name = azurerm_private_dns_zone.zone.name
  virtual_network_id    = var.spec.vnet_id
  registration_enabled  = var.spec.registration_enabled
  tags                  = local.final_tags
}
