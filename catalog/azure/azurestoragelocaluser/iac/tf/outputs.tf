output "local_user_id" {
  description = "The Azure Resource Manager ID of the local user"
  value       = azurerm_storage_account_local_user.main.id
}

output "user_name" {
  description = "The user's name within the account"
  value       = azurerm_storage_account_local_user.main.name
}

# The full SFTP login -- what a client passes as the username when
# connecting to {account}.blob.core.windows.net on port 22.
output "sftp_username" {
  description = "The full SFTP login: {account-name}.{user-name}"
  value       = local.sftp_username
}

# Azure generates the SID at creation; Azure Files NTFS-style ACLs
# reference principals by SID. An identifier, not a credential: the
# provider marks it sensitive, and nonsensitive() unwraps it.
output "sid" {
  description = "The user's unique Security Identifier"
  value       = nonsensitive(azurerm_storage_account_local_user.main.sid)
}

# Returned by Azure exactly once, at the creation that enabled
# ssh_password_enabled; empty when password auth is off. Losing it
# means regenerating it (flip ssh_password_enabled off and on).
output "password" {
  description = "The Azure-generated SSH password"
  value       = azurerm_storage_account_local_user.main.password
  sensitive   = true
}

# Parsed from the account ARM ID -- consumers frequently need the
# account/user pair.
output "storage_account_name" {
  description = "The name of the storage account the user lives on"
  value       = local.storage_account_name
}
