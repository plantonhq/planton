output "drg_id" {
  description = "OCID of the DRG."
  value       = oci_core_drg.this.id
}

output "default_export_drg_route_distribution_id" {
  description = "OCID of the default export route distribution created with the DRG."
  value       = oci_core_drg.this.default_export_drg_route_distribution_id
}
