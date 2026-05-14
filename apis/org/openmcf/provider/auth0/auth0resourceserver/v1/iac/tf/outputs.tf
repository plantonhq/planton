# Auth0ResourceServer Outputs
# This file exports outputs from the Auth0 Resource Server deployment

output "id" {
  description = "The internal Auth0 identifier for this resource server"
  value       = auth0_resource_server.this.id
}

output "identifier" {
  description = "The API identifier (audience) for this resource server"
  value       = auth0_resource_server.this.identifier
}

output "name" {
  description = "The friendly display name of the resource server"
  value       = auth0_resource_server.this.name
}

output "signing_alg" {
  description = "The algorithm used to sign tokens for this API"
  value       = auth0_resource_server.this.signing_alg
}

output "signing_secret" {
  description = "The secret used for signing tokens (HS256 only)"
  value       = auth0_resource_server.this.signing_secret
  sensitive   = true
}

output "token_lifetime" {
  description = "The token validity duration in seconds"
  value       = auth0_resource_server.this.token_lifetime
}

output "token_lifetime_for_web" {
  description = "The token validity for implicit/hybrid flows"
  value       = auth0_resource_server.this.token_lifetime_for_web
}

output "allow_offline_access" {
  description = "Indicates if refresh tokens can be issued"
  value       = auth0_resource_server.this.allow_offline_access
}

output "skip_consent_for_verifiable_first_party_clients" {
  description = "Indicates consent skip setting for first-party clients"
  value       = auth0_resource_server.this.skip_consent_for_verifiable_first_party_clients
}

output "enforce_policies" {
  description = "Indicates if RBAC is enabled for this API"
  value       = auth0_resource_server.this.enforce_policies
}

output "token_dialect" {
  description = "The access token format configured for this API"
  value       = auth0_resource_server.this.token_dialect
}

output "is_system" {
  description = "Indicates if this is a system-managed resource server (e.g. Auth0 Management API)"
  value       = auth0_resource_server.this.is_system
}

output "client_id" {
  description = "The client ID associated with this resource server, if one has been linked"
  value       = auth0_resource_server.this.client_id
}
