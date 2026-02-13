output "lb_id" {
  description = "The Azure Resource Manager ID of the Load Balancer"
  value       = azurerm_lb.main.id
}

output "lb_name" {
  description = "The name of the Load Balancer"
  value       = azurerm_lb.main.name
}

output "frontend_ip_address" {
  description = "The frontend IP address (public or private)"
  value       = azurerm_lb.main.frontend_ip_configuration[0].private_ip_address
}

output "frontend_ip_configuration_id" {
  description = "The Azure Resource Manager ID of the frontend IP configuration"
  value       = azurerm_lb.main.frontend_ip_configuration[0].id
}

output "backend_pool_id" {
  description = "The Azure Resource Manager ID of the first (default) backend address pool"
  value       = azurerm_lb_backend_address_pool.pools[var.spec.backend_pools[0].name].id
}
