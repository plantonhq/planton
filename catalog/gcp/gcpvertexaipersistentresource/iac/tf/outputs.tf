output "name" {
  description = "Full resource name of the persistent resource (projects/{project}/locations/{location}/persistentResources/{persistent_resource_id})"
  value       = google_vertex_ai_persistent_resource.this.id
}

output "persistent_resource_id" {
  description = "The resource's id -- what a training job's persistent_resource_id takes"
  value       = google_vertex_ai_persistent_resource.this.name
}

output "location" {
  description = "The location the resource runs in"
  value       = google_vertex_ai_persistent_resource.this.location
}

output "state" {
  description = "The resource's state as last read (RUNNING once every pool is provisioned)"
  value       = google_vertex_ai_persistent_resource.this.state
}
