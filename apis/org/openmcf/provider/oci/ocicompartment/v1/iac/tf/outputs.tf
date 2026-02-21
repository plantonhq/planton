output "compartment_id" {
  description = "OCID of the created compartment"
  value       = oci_identity_compartment.this.id
}
