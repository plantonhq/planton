output "name" {
  description = "Full resource name of the template (projects/{project}/locations/{location}/notebookRuntimeTemplates/{runtime_template_id})"
  value       = google_colab_runtime_template.this.id
}

output "runtime_template_id" {
  description = "The template's id"
  value       = google_colab_runtime_template.this.name
}

output "location" {
  description = "The region the template lives in"
  value       = google_colab_runtime_template.this.location
}
