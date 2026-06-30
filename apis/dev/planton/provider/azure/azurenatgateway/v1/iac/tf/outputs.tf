output "nat_gateway_id" {
  description = "Resource ID of the created NAT Gateway"
  value       = azurerm_nat_gateway.main.id
}

output "public_ip_addresses" {
  description = "List of public IP addresses allocated to the NAT Gateway"
  value       = local.use_ip_prefix ? [] : [azurerm_public_ip.nat[0].ip_address]
}

output "public_ip_prefix_id" {
  description = "Resource ID of the Public IP Prefix, if a prefix was created"
  value       = local.use_ip_prefix ? azurerm_public_ip_prefix.nat[0].id : ""
}

output "resource_group" {
  description = "Resource group where NAT Gateway was created"
  value       = local.resource_group
}

output "location" {
  description = "Azure region where NAT Gateway was deployed"
  value       = local.location
}

