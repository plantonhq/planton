# outputs.tf - GCP Project outputs

output "project_id" {
  description = "The unique project ID (globally unique across all GCP)"
  value       = google_project.this.project_id
}

output "project_number" {
  description = "The numeric identifier of the project (assigned by Google)"
  value       = google_project.this.number
}

output "project_name" {
  description = "The display name of the project"
  value       = google_project.this.name
}

output "enabled_apis" {
  description = "List of APIs that were enabled on the project"
  value       = local.enabled_apis
}

