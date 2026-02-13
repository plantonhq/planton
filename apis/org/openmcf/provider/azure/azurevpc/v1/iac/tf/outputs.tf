output "vnet_id" {
  description = "The Azure resource ID of the created Virtual Network"
  value       = azurerm_virtual_network.main.id
}

output "nodes_subnet_id" {
  description = "The Azure resource ID of the primary subnet for nodes"
  value       = azurerm_subnet.nodes.id
}

output "resource_group_name" {
  description = "The name of the resource group"
  value       = local.resource_group_name
}

output "location" {
  description = "The Azure region where resources were deployed"
  value       = var.spec.region
}

output "nat_gateway_id" {
  description = "The Azure resource ID of the NAT Gateway (if enabled)"
  value       = var.spec.is_nat_gateway_enabled ? azurerm_nat_gateway.main[0].id : null
}

