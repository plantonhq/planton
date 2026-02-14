output "profile_id" {
  description = "The Azure Resource Manager ID of the Front Door profile"
  value       = azurerm_cdn_frontdoor_profile.main.id
}

output "profile_name" {
  description = "The name of the Front Door profile"
  value       = azurerm_cdn_frontdoor_profile.main.name
}

output "resource_guid" {
  description = "The Front Door resource GUID assigned by Azure"
  value       = azurerm_cdn_frontdoor_profile.main.resource_guid
}

output "endpoint_ids" {
  description = "Map of endpoint names to their Azure Resource Manager IDs"
  value       = { for k, v in azurerm_cdn_frontdoor_endpoint.endpoints : k => v.id }
}

output "endpoint_hostnames" {
  description = "Map of endpoint names to their generated hostnames"
  value       = { for k, v in azurerm_cdn_frontdoor_endpoint.endpoints : k => v.host_name }
}
