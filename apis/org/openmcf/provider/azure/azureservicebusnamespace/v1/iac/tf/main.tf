# Create the Azure Service Bus namespace.
resource "azurerm_servicebus_namespace" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  sku      = var.spec.sku
  capacity = var.spec.capacity

  premium_messaging_partitions = var.spec.premium_messaging_partitions
  zone_redundant               = var.spec.zone_redundant

  minimum_tls_version            = var.spec.minimum_tls_version
  public_network_access_enabled  = var.spec.public_network_access_enabled

  tags = local.final_tags
}

# Create queues within the namespace.
resource "azurerm_servicebus_queue" "queues" {
  for_each = { for q in var.spec.queues : q.name => q }

  name         = each.value.name
  namespace_id = azurerm_servicebus_namespace.main.id

  max_size_in_megabytes                = each.value.max_size_in_megabytes
  partitioning_enabled                 = each.value.partitioning_enabled
  default_message_ttl                  = each.value.default_message_ttl
  lock_duration                        = each.value.lock_duration
  max_delivery_count                   = each.value.max_delivery_count
  requires_duplicate_detection         = each.value.requires_duplicate_detection
  requires_session                     = each.value.requires_session
  dead_lettering_on_message_expiration = each.value.dead_lettering_on_message_expiration
  forward_to                           = each.value.forward_to
  forward_dead_lettered_messages_to    = each.value.forward_dead_lettered_messages_to
}

# Create topics within the namespace.
resource "azurerm_servicebus_topic" "topics" {
  for_each = { for t in var.spec.topics : t.name => t }

  name         = each.value.name
  namespace_id = azurerm_servicebus_namespace.main.id

  max_size_in_megabytes        = each.value.max_size_in_megabytes
  partitioning_enabled         = each.value.partitioning_enabled
  default_message_ttl          = each.value.default_message_ttl
  requires_duplicate_detection = each.value.requires_duplicate_detection
  support_ordering             = each.value.support_ordering
}
