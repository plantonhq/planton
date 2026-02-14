output "environment_id" {
  description = "The Azure Resource Manager ID of the Container App Environment"
  value       = azurerm_container_app_environment.main.id
}

output "default_domain" {
  description = "The default publicly resolvable domain for apps in this environment"
  value       = azurerm_container_app_environment.main.default_domain
}

output "static_ip_address" {
  description = "The static IP address of the environment"
  value       = azurerm_container_app_environment.main.static_ip_address
}

output "platform_reserved_cidr" {
  description = "The IP range reserved for environment infrastructure"
  value       = azurerm_container_app_environment.main.platform_reserved_cidr
}

output "platform_reserved_dns_ip_address" {
  description = "The IP address reserved for the internal DNS server"
  value       = azurerm_container_app_environment.main.platform_reserved_dns_ip_address
}
