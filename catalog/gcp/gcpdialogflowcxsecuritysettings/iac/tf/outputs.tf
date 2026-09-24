output "name" {
  description = "Full resource name of the security settings (projects/{project}/locations/{location}/securitySettings/{id}) -- what an agent's security_settings field takes"
  value       = google_dialogflow_cx_security_settings.this.id
}

output "security_settings_id" {
  description = "The id Google assigned the security settings"
  value       = google_dialogflow_cx_security_settings.this.name
}

output "location" {
  description = "The location the settings live in; only agents there can use them"
  value       = google_dialogflow_cx_security_settings.this.location
}
