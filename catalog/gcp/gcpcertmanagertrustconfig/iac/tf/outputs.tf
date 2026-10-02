output "trust_config_id" {
  description = "Full resource name (projects/{project}/locations/{location}/trustConfigs/{name}) -- what a server TLS policy and a backend authentication config take"
  value       = google_certificate_manager_trust_config.this.id
}

output "trust_config_name" {
  description = "Name of the trust config as it exists in GCP"
  value       = google_certificate_manager_trust_config.this.name
}

output "location" {
  description = "The Certificate Manager location the trust config lives in"
  value       = local.location
}
