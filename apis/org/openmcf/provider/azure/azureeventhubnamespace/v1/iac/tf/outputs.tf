output "namespace_id" {
  description = "The Azure Resource Manager ID of the Event Hubs namespace"
  value       = azurerm_eventhub_namespace.main.id
}

output "namespace_name" {
  description = "The name of the Event Hubs namespace"
  value       = azurerm_eventhub_namespace.main.name
}

output "primary_connection_string" {
  description = "The primary connection string from the default RootManageSharedAccessKey"
  value       = azurerm_eventhub_namespace.main.default_primary_connection_string
  sensitive   = true
}

output "primary_key" {
  description = "The primary SAS key from the default RootManageSharedAccessKey"
  value       = azurerm_eventhub_namespace.main.default_primary_key
  sensitive   = true
}

output "event_hub_ids" {
  description = "Map of event hub names to their Azure Resource Manager IDs"
  value       = { for k, v in azurerm_eventhub.hubs : k => v.id }
}
