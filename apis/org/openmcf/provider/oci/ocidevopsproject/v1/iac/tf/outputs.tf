output "project_id" {
  description = "OCID of the DevOps project"
  value       = oci_devops_project.this.id
}

output "namespace" {
  description = "Namespace associated with the project"
  value       = oci_devops_project.this.namespace
}
