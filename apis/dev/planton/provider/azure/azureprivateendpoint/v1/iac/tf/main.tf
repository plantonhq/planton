# Private Endpoint
# Creates a private network interface that connects to a Private Link-enabled resource.
resource "azurerm_private_endpoint" "endpoint" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  subnet_id           = var.spec.subnet_id
  tags                = local.final_tags

  private_service_connection {
    name                           = local.connection_name
    private_connection_resource_id = var.spec.private_connection_resource_id
    is_manual_connection           = false
    subresource_names              = var.spec.subresource_names
  }

  dynamic "private_dns_zone_group" {
    for_each = var.spec.private_dns_zone_id != null ? [1] : []
    content {
      name                 = local.dns_zone_group_name
      private_dns_zone_ids = [var.spec.private_dns_zone_id]
    }
  }
}
