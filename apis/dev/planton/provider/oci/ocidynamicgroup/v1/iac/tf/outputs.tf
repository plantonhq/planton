output "dynamic_group_id" {
  description = "OCID of the created dynamic group"
  value       = oci_identity_dynamic_group.this.id
}
