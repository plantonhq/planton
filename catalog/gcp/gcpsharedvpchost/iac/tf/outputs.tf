# The project ID that is now the Shared VPC host -- the resolved value, so it
# is populated even when the spec left project_id empty. What a
# GcpSharedVpcServiceProject's host_project_id references.
output "host_project_id" {
  description = "The project ID enabled as the Shared VPC host"
  value       = google_compute_shared_vpc_host_project.this.project
}
