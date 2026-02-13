output "identity_id" {
  description = "The Azure Resource Manager ID of the User-Assigned Managed Identity"
  value       = azurerm_user_assigned_identity.main.id
}

output "principal_id" {
  description = "The Service Principal Object ID associated with this Managed Identity"
  value       = azurerm_user_assigned_identity.main.principal_id
}

output "client_id" {
  description = "The Client ID (Application ID) of the Managed Identity"
  value       = azurerm_user_assigned_identity.main.client_id
}

output "tenant_id" {
  description = "The Azure AD Tenant ID that the Managed Identity belongs to"
  value       = azurerm_user_assigned_identity.main.tenant_id
}
