output "volume_id" {
  description = "OCID of the block volume"
  value       = oci_core_volume.this.id
}
