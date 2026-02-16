output "instance_id" {
  description = "Fully qualified instance ID (projects/{project}/instances/{name})"
  value       = "projects/${google_spanner_instance.this.project}/instances/${google_spanner_instance.this.name}"
}

output "instance_name" {
  description = "Short instance name"
  value       = google_spanner_instance.this.name
}

output "state" {
  description = "Instance state (CREATING or READY)"
  value       = google_spanner_instance.this.state
}
