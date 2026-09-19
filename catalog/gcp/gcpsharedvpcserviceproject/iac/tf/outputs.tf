# The project ID now attached as a service project.
output "service_project_id" {
  description = "The attached service project's ID"
  value       = google_compute_shared_vpc_service_project.this.service_project
}

# The host project ID it is attached to.
output "host_project_id" {
  description = "The Shared VPC host project's ID"
  value       = google_compute_shared_vpc_service_project.this.host_project
}
