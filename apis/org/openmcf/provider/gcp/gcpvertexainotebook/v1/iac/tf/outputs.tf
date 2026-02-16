output "instance_id" {
  description = "The fully qualified instance ID (projects/{project}/locations/{location}/instances/{instance_id})"
  value       = google_workbench_instance.this.id
}

output "instance_name" {
  description = "The short instance name"
  value       = google_workbench_instance.this.name
}

output "proxy_uri" {
  description = "The JupyterLab proxy URI for accessing the notebook"
  value       = google_workbench_instance.this.proxy_uri
}

output "state" {
  description = "The current state of the instance (ACTIVE, STOPPED, etc.)"
  value       = google_workbench_instance.this.state
}

output "creator" {
  description = "Email address of the entity that created the instance"
  value       = google_workbench_instance.this.creator
}

output "create_time" {
  description = "RFC3339 timestamp of when the instance was created"
  value       = google_workbench_instance.this.create_time
}
