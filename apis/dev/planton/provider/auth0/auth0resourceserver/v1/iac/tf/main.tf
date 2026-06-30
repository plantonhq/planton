# Auth0ResourceServer Main Resources
# This file creates the Auth0 Resource Server (API)

# Auth0 Resource Server Resource
resource "auth0_resource_server" "this" {
  identifier = local.identifier
  name       = local.name

  # Signing algorithm (optional)
  signing_alg = local.signing_alg

  # Token settings
  allow_offline_access   = local.allow_offline_access
  token_lifetime         = local.token_lifetime
  token_lifetime_for_web = local.token_lifetime_for_web

  # Access control settings
  skip_consent_for_verifiable_first_party_clients = local.skip_consent_for_verifiable_first_party_clients
  enforce_policies = local.enforce_policies
  token_dialect    = local.token_dialect
}

# Auth0 Resource Server Scopes
# Creates scopes (permissions) for the resource server
resource "auth0_resource_server_scopes" "this" {
  count = length(local.scopes) > 0 ? 1 : 0

  resource_server_identifier = auth0_resource_server.this.identifier

  dynamic "scopes" {
    for_each = local.scopes
    content {
      name        = scopes.value.name
      description = scopes.value.description
    }
  }
}
