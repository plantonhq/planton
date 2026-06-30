output "container_app_id" {
  description = "The Azure Resource Manager ID of the Container App"
  value       = azurerm_container_app.main.id
}

output "latest_revision_name" {
  description = "The name of the latest Container Revision"
  value       = azurerm_container_app.main.latest_revision_name
}

output "latest_revision_fqdn" {
  description = "The FQDN of the latest Container Revision"
  value       = azurerm_container_app.main.latest_revision_fqdn
}

output "outbound_ip_addresses" {
  description = "Outbound IP addresses of the Container App"
  value       = azurerm_container_app.main.outbound_ip_addresses
}

output "ingress_fqdn" {
  description = "The ingress FQDN of the Container App"
  value       = try(azurerm_container_app.main.ingress[0].fqdn, "")
}
