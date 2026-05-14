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

output "client_secret" {
  description = "The OAuth 2.0 client secret for confidential client authentication"
  value       = data.auth0_client.this.client_secret
  sensitive   = true
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

output "allowed_clients" {
  description = "List of client IDs allowed to perform delegation for this client"
  value       = auth0_client.this.allowed_clients
}

output "token_endpoint_auth_method" {
  description = "The authentication method for the token endpoint"
  value       = data.auth0_client.this.token_endpoint_auth_method
}
