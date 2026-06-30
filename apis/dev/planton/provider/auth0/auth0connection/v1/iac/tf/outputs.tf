# Auth0Connection Outputs
# These outputs match the Auth0ConnectionStackOutputs protobuf message

output "id" {
  description = "The unique identifier of the Auth0 connection"
  value       = auth0_connection.this.id
}

output "name" {
  description = "The unique name of the connection within the Auth0 tenant"
  value       = auth0_connection.this.name
}

output "strategy" {
  description = "The identity provider strategy type"
  value       = auth0_connection.this.strategy
}

output "is_enabled" {
  description = "Whether the connection is currently enabled (has enabled clients)"
  value       = length(local.enabled_clients) > 0
}

output "provisioning_ticket_url" {
  description = "URL for self-service connection setup (enterprise connections only)"
  value       = ""
}

output "callback_url" {
  description = "The Auth0 callback URL for this connection"
  value       = ""
}

output "metadata_url" {
  description = "SAML metadata URL (SAML connections only)"
  value       = ""
}

output "entity_id" {
  description = "SAML Service Provider Entity ID (SAML connections only)"
  value       = ""
}

output "enabled_client_ids" {
  description = "List of Auth0 application client IDs that can use this connection"
  value       = local.enabled_clients
}

output "realms" {
  description = "List of realms/domains associated with this connection"
  value       = auth0_connection.this.realms
}
