output "zone_id" {
  description = "The Azure Resource Manager ID of the Private DNS Zone"
  value       = azurerm_private_dns_zone.zone.id
}

output "zone_name" {
  description = "The name of the Private DNS Zone"
  value       = azurerm_private_dns_zone.zone.name
}
