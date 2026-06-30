output "zone_id" {
  description = "The Azure Resource Manager ID of the DNS zone"
  value       = azurerm_dns_zone.dns_zone.id
}

output "zone_name" {
  description = "The name of the DNS zone"
  value       = azurerm_dns_zone.dns_zone.name
}

output "nameservers" {
  description = "The list of nameservers assigned to this DNS zone. Configure these at your domain registrar."
  value       = azurerm_dns_zone.dns_zone.name_servers
}

output "max_number_of_record_sets" {
  description = "Maximum number of record sets that can be created in this DNS zone"
  value       = azurerm_dns_zone.dns_zone.max_number_of_record_sets
}

output "number_of_record_sets" {
  description = "Current number of record sets in this DNS zone"
  value       = azurerm_dns_zone.dns_zone.number_of_record_sets
}

