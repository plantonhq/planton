output "workspace_id" {
  description = "The Azure Resource Manager ID of the Log Analytics Workspace"
  value       = azurerm_log_analytics_workspace.main.id
}

output "workspace_name" {
  description = "The name of the Log Analytics Workspace"
  value       = azurerm_log_analytics_workspace.main.name
}

output "primary_shared_key" {
  description = "The primary shared key for agent authentication"
  value       = azurerm_log_analytics_workspace.main.primary_shared_key
  sensitive   = true
}

output "secondary_shared_key" {
  description = "The secondary shared key for agent authentication"
  value       = azurerm_log_analytics_workspace.main.secondary_shared_key
  sensitive   = true
}
