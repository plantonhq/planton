output "container_app_id" {
  description = "The Azure Resource Manager ID of the Container App"
  value       = azurerm_container_app.main.id
}

output "container_app_name" {
  description = "The name of the Container App"
  value       = azurerm_container_app.main.name
}

output "latest_revision_name" {
  description = "The name of the latest revision"
  value       = azurerm_container_app.main.latest_revision_name
}

output "latest_revision_fqdn" {
  description = "The FQDN of the latest revision (bypasses traffic splitting)"
  value       = azurerm_container_app.main.latest_revision_fqdn
}

output "outbound_ip_addresses" {
  description = "Outbound IP addresses used by the app for egress traffic"
  value       = azurerm_container_app.main.outbound_ip_addresses
}

output "ingress_fqdn" {
  description = "The app's main FQDN (empty when ingress is not configured)"
  value       = try(azurerm_container_app.main.ingress[0].fqdn, "")
}

# The verification ID is published in a DNS TXT record. The provider marks
# the attribute sensitive; nonsensitive() unwraps it.
output "custom_domain_verification_id" {
  description = "The TXT-record value proving domain ownership when binding a custom domain"
  value       = nonsensitive(azurerm_container_app.main.custom_domain_verification_id)
}

output "identity_principal_id" {
  description = "The principal ID of the app's system-assigned managed identity (empty unless SYSTEM_ASSIGNED is enabled)"
  value       = try(azurerm_container_app.main.identity[0].principal_id, "")
}
