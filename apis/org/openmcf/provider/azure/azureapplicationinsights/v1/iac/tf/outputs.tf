output "app_insights_id" {
  description = "The Azure Resource Manager ID of the Application Insights resource"
  value       = azurerm_application_insights.main.id
}

output "instrumentation_key" {
  description = "The instrumentation key for classic SDK configuration"
  value       = azurerm_application_insights.main.instrumentation_key
  sensitive   = true
}

output "connection_string" {
  description = "The connection string for SDK configuration"
  value       = azurerm_application_insights.main.connection_string
  sensitive   = true
}

output "app_id" {
  description = "The Application ID for API access"
  value       = azurerm_application_insights.main.app_id
}
