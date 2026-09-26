output "application_insights_id" {
  description = "The Azure Resource Manager ID of the Application Insights resource"
  value       = azurerm_application_insights.main.id
}

output "application_insights_name" {
  description = "The name of the Application Insights resource"
  value       = azurerm_application_insights.main.name
}

output "instrumentation_key" {
  description = "The instrumentation key for classic SDK configuration (prefer the connection string)"
  value       = nonsensitive(azurerm_application_insights.main.instrumentation_key)
}

output "connection_string" {
  description = "The connection string SDKs are configured with -- the composition seam app kinds reference"
  value       = nonsensitive(azurerm_application_insights.main.connection_string)
}

output "app_id" {
  description = "The Application ID used when querying telemetry via the REST API"
  value       = azurerm_application_insights.main.app_id
}
