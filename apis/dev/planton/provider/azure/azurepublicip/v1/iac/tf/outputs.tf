output "public_ip_id" {
  description = "The Azure Resource Manager ID of the Public IP"
  value       = azurerm_public_ip.main.id
}

output "ip_address" {
  description = "The allocated static IPv4 address"
  value       = azurerm_public_ip.main.ip_address
}

output "fqdn" {
  description = "The fully qualified domain name (if domain_name_label is set)"
  value       = azurerm_public_ip.main.fqdn
}

output "public_ip_name" {
  description = "The name of the Public IP resource"
  value       = azurerm_public_ip.main.name
}
