# Create the Azure Subnet within an existing Virtual Network.
resource "azurerm_subnet" "main" {
  name                 = var.spec.name
  resource_group_name  = var.spec.resource_group
  virtual_network_name = local.vnet_name
  address_prefixes     = [var.spec.address_prefix]

  service_endpoints = var.spec.service_endpoints

  private_endpoint_network_policies             = var.spec.private_endpoint_network_policies
  private_link_service_network_policies_enabled = var.spec.private_link_service_network_policies_enabled

  dynamic "delegation" {
    for_each = var.spec.delegation != null ? [var.spec.delegation] : []
    content {
      name = delegation.value.name
      service_delegation {
        name    = delegation.value.service_name
        actions = delegation.value.actions
      }
    }
  }
}
