output "issuance_config_id" {
  description = "Full resource name (projects/{project}/locations/{location}/certificateIssuanceConfigs/{name}) -- what a GcpCertManagerCert's managed.issuance_config takes"
  value       = google_certificate_manager_certificate_issuance_config.this.id
}

output "issuance_config_name" {
  description = "Name of the issuance config as it exists in GCP"
  value       = google_certificate_manager_certificate_issuance_config.this.name
}

output "location" {
  description = "The Certificate Manager location the issuance config lives in"
  value       = var.spec.location != "" ? var.spec.location : "global"
}
