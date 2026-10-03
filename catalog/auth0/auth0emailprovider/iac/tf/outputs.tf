# Auth0EmailProvider Outputs
# Maps to the Auth0EmailProviderOutputs protobuf message: the provider as
# applied.

output "name" {
  description = "The service the tenant sends through, as Auth0 names it (smtp, ses, sendgrid, ...)"
  value       = auth0_email_provider.this.name
}

output "default_from_address" {
  description = "The sender of the tenant's emails"
  value       = auth0_email_provider.this.default_from_address
}
