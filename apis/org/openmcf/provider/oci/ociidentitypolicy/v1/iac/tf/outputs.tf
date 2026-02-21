output "policy_id" {
  description = "OCID of the created policy"
  value       = oci_identity_policy.this.id
}
