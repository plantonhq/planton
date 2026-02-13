output "subnet_id" {
  description = "The Azure Resource Manager ID of the subnet"
  value       = azurerm_subnet.main.id
}

output "subnet_name" {
  description = "The name of the subnet"
  value       = azurerm_subnet.main.name
}

output "address_prefix" {
  description = "The IPv4 CIDR block assigned to the subnet"
  value       = var.spec.address_prefix
}
