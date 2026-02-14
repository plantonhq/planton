output "namespace_id" {
  description = "The Azure Resource Manager ID of the Service Bus namespace"
  value       = azurerm_servicebus_namespace.main.id
}

output "namespace_name" {
  description = "The name of the Service Bus namespace"
  value       = azurerm_servicebus_namespace.main.name
}

output "endpoint" {
  description = "The Service Bus endpoint URL"
  value       = azurerm_servicebus_namespace.main.endpoint
}

output "primary_connection_string" {
  description = "The primary connection string from the default RootManageSharedAccessKey"
  value       = azurerm_servicebus_namespace.main.default_primary_connection_string
  sensitive   = true
}

output "primary_key" {
  description = "The primary SAS key from the default RootManageSharedAccessKey"
  value       = azurerm_servicebus_namespace.main.default_primary_key
  sensitive   = true
}

output "queue_ids" {
  description = "Map of queue names to their Azure Resource Manager IDs"
  value       = { for k, v in azurerm_servicebus_queue.queues : k => v.id }
}

output "topic_ids" {
  description = "Map of topic names to their Azure Resource Manager IDs"
  value       = { for k, v in azurerm_servicebus_topic.topics : k => v.id }
}
