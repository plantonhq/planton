# Output the service account email
# This is always available and can be used for IAM bindings or attachments
output "email" {
  description = "The email address of the created service account"
  value       = google_service_account.main.email
}

# Output the base64-encoded private key (if created)
# This is sensitive and should be stored securely in a secret manager
output "key_base64" {
  description = "The base64-encoded private key JSON for the service account (if create_key was true)"
  value       = local.create_key ? google_service_account_key.main[0].private_key : null
  sensitive   = true
}

