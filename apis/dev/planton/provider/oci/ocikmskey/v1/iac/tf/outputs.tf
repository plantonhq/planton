output "key_id" {
  description = "OCID of the KMS key"
  value       = oci_kms_key.this.id
}

output "current_key_version" {
  description = "OCID of the currently active key version"
  value       = oci_kms_key.this.current_key_version
}
