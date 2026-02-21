output "secret_id" {
  description = "OCID of the Vault Secret"
  value       = oci_vault_secret.this.id
}

output "current_version_number" {
  description = "Version number of the currently active secret version"
  value       = oci_vault_secret.this.current_version_number
}
