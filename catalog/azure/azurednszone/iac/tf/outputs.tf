output "zone_id" {
  description = "The Azure Resource Manager ID of the DNS zone -- the seam for kinds that watch or manage the zone as a whole (Front Door custom-domain validation, AKS web-app routing)."
  value       = azurerm_dns_zone.main.id
}

output "zone_name" {
  description = "The DNS zone name -- the join key AzureDnsRecord resources address record sets through."
  value       = azurerm_dns_zone.main.name
}

output "resource_group_name" {
  description = "The resource group the zone lives in -- Azure's management plane addresses records by zone name + resource group."
  value       = azurerm_dns_zone.main.resource_group_name
}

output "nameservers" {
  description = "The four name servers Azure assigned to this zone. The zone only answers the internet once these are configured at the domain's registrar (or as parent-zone NS records for subdomain delegation)."
  value       = azurerm_dns_zone.main.name_servers
}

output "max_number_of_record_sets" {
  description = "The maximum number of record sets this zone can hold -- Azure's per-zone capacity limit."
  value       = azurerm_dns_zone.main.max_number_of_record_sets
}
