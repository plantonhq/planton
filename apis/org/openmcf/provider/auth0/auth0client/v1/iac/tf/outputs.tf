# Auth0Client Outputs
# These outputs match the Auth0ClientStackOutputs protobuf message

output "id" {
  description = "The unique identifier of the Auth0 client"
  value       = auth0_client.this.id
}

output "client_id" {
  description = "The OAuth 2.0 client identifier (public identifier)"
  value       = auth0_client.this.client_id
}

output "name" {
  description = "The name of the application"
  value       = auth0_client.this.name
}

output "application_type" {
  description = "The type of application (native, spa, regular_web, non_interactive)"
  value       = auth0_client.this.app_type
}

output "signing_keys" {
  description = "Signing keys for this client (for RS256 token verification)"
  value       = auth0_client.this.signing_keys
  sensitive   = true
}


