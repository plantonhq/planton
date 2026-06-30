# Create the Azure Event Hubs namespace.
resource "azurerm_eventhub_namespace" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  sku      = var.spec.sku
  capacity = var.spec.capacity

  auto_inflate_enabled    = var.spec.auto_inflate_enabled
  maximum_throughput_units = var.spec.maximum_throughput_units
  zone_redundant          = var.spec.zone_redundant

  minimum_tls_version           = var.spec.minimum_tls_version
  public_network_access_enabled = var.spec.public_network_access_enabled

  tags = local.final_tags
}

# Create event hubs within the namespace.
resource "azurerm_eventhub" "hubs" {
  for_each = { for eh in var.spec.event_hubs : eh.name => eh }

  name         = each.value.name
  namespace_id = azurerm_eventhub_namespace.main.id

  partition_count   = each.value.partition_count
  message_retention = each.value.message_retention
}

# Create consumer groups within event hubs.
resource "azurerm_eventhub_consumer_group" "groups" {
  for_each = local.consumer_groups_map

  name                = each.value.name
  namespace_name      = azurerm_eventhub_namespace.main.name
  eventhub_name       = each.value.eventhub_name
  resource_group_name = var.spec.resource_group
  user_metadata       = each.value.user_metadata

  depends_on = [azurerm_eventhub.hubs]
}
