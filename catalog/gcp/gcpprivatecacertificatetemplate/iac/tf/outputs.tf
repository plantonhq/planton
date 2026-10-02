output "name" {
  description = "Full resource name of the template (projects/{project}/locations/{location}/certificateTemplates/{template_id}) -- what certificates reference"
  value       = google_privateca_certificate_template.this.id
}

output "template_id" {
  description = "The template's ID"
  value       = google_privateca_certificate_template.this.name
}
