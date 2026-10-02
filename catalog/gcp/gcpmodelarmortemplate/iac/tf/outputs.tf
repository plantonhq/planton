output "name" {
  description = "Full resource name of the template (projects/{project}/locations/{location}/templates/{template_id})"
  value       = google_model_armor_template.this.name
}

output "template_id" {
  description = "The template's id"
  value       = google_model_armor_template.this.template_id
}

output "location" {
  description = "The location the template lives in"
  value       = google_model_armor_template.this.location
}
