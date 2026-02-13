output "resource_group_id" {
  description = "The Azure Resource Manager ID of the resource group"
  value       = azurerm_resource_group.main.id
}

output "resource_group_name" {
  description = "The name of the resource group"
  value       = azurerm_resource_group.main.name
}

output "region" {
  description = "The Azure region where the resource group was created"
  value       = azurerm_resource_group.main.location
}
