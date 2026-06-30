output "vault_id" {
  description = "OCID of the KMS vault"
  value       = oci_kms_vault.this.id
}

output "crypto_endpoint" {
  description = "Service endpoint for cryptographic operations"
  value       = oci_kms_vault.this.crypto_endpoint
}

output "management_endpoint" {
  description = "Service endpoint for key management operations"
  value       = oci_kms_vault.this.management_endpoint
}
