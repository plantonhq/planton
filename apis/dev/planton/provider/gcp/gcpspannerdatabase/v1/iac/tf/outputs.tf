output "database_id" {
  description = "Fully qualified database ID (projects/{project}/instances/{instance}/databases/{name})"
  value       = "projects/${local.project_id}/instances/${local.instance_name}/databases/${google_spanner_database.this.name}"
}

output "database_name" {
  description = "Short database name"
  value       = google_spanner_database.this.name
}

output "state" {
  description = "Database state (CREATING or READY)"
  value       = google_spanner_database.this.state
}
