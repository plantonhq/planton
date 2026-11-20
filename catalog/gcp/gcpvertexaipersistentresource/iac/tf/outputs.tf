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

# Flattened from the provider's one-element list; code 0 and an empty
# message while the resource has no error.
output "error" {
  description = "Why the resource is in ERROR: Google's error code and message (code 0 when there is none)"
  value = {
    code    = try(google_vertex_ai_persistent_resource.this.error[0].code, 0)
    message = try(google_vertex_ai_persistent_resource.this.error[0].message, "")
  }
}
