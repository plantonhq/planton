output "container_instance_id" {
  description = "OCID of the container instance"
  value       = oci_container_instances_container_instance.this.id
}

output "container_ids" {
  description = "Comma-separated OCIDs of the individual containers"
  value       = join(",", [for c in oci_container_instances_container_instance.this.containers : c.container_id])
}
