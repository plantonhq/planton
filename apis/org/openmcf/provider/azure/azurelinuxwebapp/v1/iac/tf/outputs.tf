output "web_app_id" {
  description = "The Azure Resource Manager ID of the Web App"
  value       = azurerm_linux_web_app.main.id
}

output "default_hostname" {
  description = "The default hostname of the Web App ({name}.azurewebsites.net)"
  value       = azurerm_linux_web_app.main.default_hostname
}

output "outbound_ip_addresses" {
  description = "Outbound IP addresses used by the Web App"
  value       = split(",", azurerm_linux_web_app.main.outbound_ip_addresses)
}

output "identity_principal_id" {
  description = "The principal ID of the system-assigned managed identity"
  value       = try(azurerm_linux_web_app.main.identity[0].principal_id, "")
}

output "identity_tenant_id" {
  description = "The tenant ID of the system-assigned managed identity"
  value       = try(azurerm_linux_web_app.main.identity[0].tenant_id, "")
}

output "custom_domain_verification_id" {
  description = "The custom domain verification ID for DNS TXT record verification"
  value       = azurerm_linux_web_app.main.custom_domain_verification_id
  sensitive   = true
}

output "kind" {
  description = "The resource kind string as reported by Azure (e.g., app,linux)"
  value       = azurerm_linux_web_app.main.kind
}
