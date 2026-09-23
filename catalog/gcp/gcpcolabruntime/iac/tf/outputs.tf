output "name" {
  description = "Full resource name of the runtime (projects/{project}/locations/{location}/notebookRuntimes/{runtime_id})"
  value       = google_colab_runtime.this.id
}

output "runtime_id" {
  description = "The runtime's id"
  value       = google_colab_runtime.this.name
}

output "location" {
  description = "The region the runtime lives in"
  value       = google_colab_runtime.this.location
}
