output "app_gateway_id" {
  description = "The Azure Resource Manager ID of the Application Gateway"
  value       = azurerm_application_gateway.main.id
}

output "app_gateway_name" {
  description = "The name of the Application Gateway"
  value       = azurerm_application_gateway.main.name
}
