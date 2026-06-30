# Outputs aligned to AzureContainerRegistryStackOutputs proto

output "registry_login_server" {
  description = "The registry's login server URL (hostname for pulling/pushing images)"
  value       = azurerm_container_registry.acr.login_server
}

output "registry_resource_id" {
  description = "The Azure Resource Manager ID of the container registry"
  value       = azurerm_container_registry.acr.id
}

# Additional helpful outputs

output "resource_group_name" {
  description = "Name of the resource group containing the container registry"
  value       = local.resource_group_name
}

output "registry_name" {
  description = "Name of the container registry"
  value       = azurerm_container_registry.acr.name
}

output "admin_username" {
  description = "Admin username for the registry (if admin user is enabled)"
  value       = var.spec.admin_user_enabled ? azurerm_container_registry.acr.admin_username : null
  sensitive   = true
}

output "admin_password" {
  description = "Admin password for the registry (if admin user is enabled)"
  value       = var.spec.admin_user_enabled ? azurerm_container_registry.acr.admin_password : null
  sensitive   = true
}

