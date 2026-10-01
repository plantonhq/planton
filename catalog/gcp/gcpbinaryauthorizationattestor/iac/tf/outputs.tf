output "attestor_id" {
  description = "Full resource name of the attestor (projects/{project}/attestors/{name})"
  value       = google_binary_authorization_attestor.this.id
}

output "attestor_name" {
  description = "The attestor's ID in its project"
  value       = google_binary_authorization_attestor.this.name
}

output "note_reference" {
  description = "The note attestations are stored under (projects/{project}/notes/{note})"
  value       = local.creates_note ? google_container_analysis_note.this[0].id : google_binary_authorization_attestor.this.attestation_authority_note[0].note_reference
}

output "delegation_service_account_email" {
  description = "The service account the attestor reads attestations with"
  value       = google_binary_authorization_attestor.this.attestation_authority_note[0].delegation_service_account_email
}
