output "vault_id" {
  description = "The Azure Resource Manager ID of the Key Vault"
  value       = azurerm_key_vault.main.id
}

output "vault_name" {
  description = "The name of the Key Vault"
  value       = azurerm_key_vault.main.name
}

output "vault_uri" {
  description = "The URI of the Key Vault for accessing secrets, keys, and certificates"
  value       = azurerm_key_vault.main.vault_uri
}

output "secret_id_map" {
  description = "Map of secret names to their full secret IDs"
  value = {
    for name, secret in azurerm_key_vault_secret.secrets :
    name => secret.id
  }
}

output "region" {
  description = "The Azure region where the Key Vault was deployed"
  value       = var.spec.region
}

output "resource_group" {
  description = "The resource group name where the Key Vault was created"
  value       = var.spec.resource_group
}

