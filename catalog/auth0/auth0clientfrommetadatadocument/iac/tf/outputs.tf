# Auth0ClientFromMetadataDocument Outputs
# Maps to the Auth0ClientFromMetadataDocumentOutputs protobuf message: the
# application as Auth0 registered it, what Auth0 took from the metadata
# document, and the document's validation as of the last read.

output "client_id" {
  description = "The application's id in the Management API (tpc_...)"
  value       = auth0_client_cimd.this.client_id
}

output "external_client_id" {
  description = "The metadata document's URL, the client_id the application presents in sign-in flows"
  value       = auth0_client_cimd.this.external_client_id
}

output "name" {
  description = "The application's name, from the document's client_name"
  value       = auth0_client_cimd.this.name
}

output "app_type" {
  description = "The application type Auth0 holds"
  value       = auth0_client_cimd.this.app_type
}

output "grant_types" {
  description = "The grants Auth0 holds for the application"
  value       = auth0_client_cimd.this.grant_types
}

output "callbacks" {
  description = "The redirect URIs, from the document's redirect_uris"
  value       = auth0_client_cimd.this.callbacks
}

output "logo_uri" {
  description = "The logo shown on the consent screen, from the document"
  value       = auth0_client_cimd.this.logo_uri
}

output "jwks_uri" {
  description = "Where the application publishes its public keys, from the document"
  value       = auth0_client_cimd.this.jwks_uri
}

output "third_party_security_mode" {
  description = "The security mode Auth0 registered the application in (strict)"
  value       = auth0_client_cimd.this.third_party_security_mode
}

output "external_metadata_created_by" {
  description = "Who registered the application: admin or client"
  value       = auth0_client_cimd.this.external_metadata_created_by
}

# The validation list holds one entry when Auth0 previewed the document and
# none otherwise. An absent entry reads as not valid with nothing to report --
# the same rule as the Pulumi module's summarizeValidation.
output "validation_valid" {
  description = "Whether the document passed Auth0's validation at the last read"
  value       = coalesce(try(auth0_client_cimd.this.validation[0].valid, null), false)
}

output "validation_warnings" {
  description = "What Auth0 ignored in the document at the last read"
  value       = try(coalescelist(auth0_client_cimd.this.validation[0].warnings, []), [])
}

output "validation_violations" {
  description = "What prevents Auth0 from processing the document fully at the last read"
  value       = try(coalescelist(auth0_client_cimd.this.validation[0].violations, []), [])
}
