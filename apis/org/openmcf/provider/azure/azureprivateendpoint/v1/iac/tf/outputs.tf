output "private_endpoint_id" {
  description = "The Azure Resource Manager ID of the Private Endpoint"
  value       = azurerm_private_endpoint.endpoint.id
}

output "private_ip_address" {
  description = "The private IP address allocated to the Private Endpoint"
  value       = azurerm_private_endpoint.endpoint.private_service_connection[0].private_ip_address
}

output "network_interface_id" {
  description = "The Azure Resource Manager ID of the network interface"
  value       = azurerm_private_endpoint.endpoint.network_interface[0].id
}
