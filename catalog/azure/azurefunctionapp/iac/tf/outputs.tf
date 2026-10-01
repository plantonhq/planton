output "function_app_id" {
  description = "The Azure Resource Manager ID of the Function App"
  value       = azurerm_linux_function_app.main.id
}

output "default_hostname" {
  description = "The default hostname of the Function App ({name}.azurewebsites.net)"
  value       = azurerm_linux_function_app.main.default_hostname
}

output "outbound_ip_addresses" {
  description = "Outbound IP addresses used by the Function App"
  value       = azurerm_linux_function_app.main.outbound_ip_address_list
}

output "identity_principal_id" {
  description = "The principal ID of the system-assigned managed identity"
  value       = try(azurerm_linux_function_app.main.identity[0].principal_id, "")
}

output "identity_tenant_id" {
  description = "The tenant ID of the system-assigned managed identity"
  value       = try(azurerm_linux_function_app.main.identity[0].tenant_id, "")
}

# The verification ID is published in a DNS TXT record. The provider marks
# the attribute sensitive; nonsensitive() unwraps it.
output "custom_domain_verification_id" {
  description = "The custom domain verification ID for DNS TXT record verification"
  value       = nonsensitive(azurerm_linux_function_app.main.custom_domain_verification_id)
}

output "kind" {
  description = "The resource kind string as reported by Azure (e.g., functionapp,linux)"
  value       = azurerm_linux_function_app.main.kind
}

output "possible_outbound_ip_addresses" {
  description = "Every outbound IP the platform could ever route this app through -- use for durable firewall allowlists"
  value       = azurerm_linux_function_app.main.possible_outbound_ip_address_list
}

output "hosting_environment_id" {
  description = "ARM ID of the App Service Environment hosting the app (empty outside ASE)"
  value       = azurerm_linux_function_app.main.hosting_environment_id
}

# The username is not a secret. The provider marks the whole site_credential
# block sensitive; nonsensitive() unwraps it.
output "site_credential_name" {
  description = "The site-level publishing credential's username (Kudu/SCM basic auth)"
  value       = try(nonsensitive(azurerm_linux_function_app.main.site_credential[0].name), "")
}

output "site_credential_password" {
  description = "The site-level publishing credential's password -- grants deploy access while basic-auth publishing is enabled"
  value       = try(azurerm_linux_function_app.main.site_credential[0].password, "")
  sensitive   = true
}
