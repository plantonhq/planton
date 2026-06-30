# Auth0ResourceServer Locals
# This file contains local variables computed from input variables

locals {
  # Core configuration
  resource_name = var.metadata.name
  identifier    = var.spec.identifier
  name          = coalesce(var.spec.name, var.metadata.name)

  # Token settings
  signing_alg          = var.spec.signing_alg
  allow_offline_access = coalesce(var.spec.allow_offline_access, false)
  token_lifetime       = var.spec.token_lifetime
  token_lifetime_for_web = var.spec.token_lifetime_for_web

  # Access control settings
  skip_consent_for_verifiable_first_party_clients = coalesce(var.spec.skip_consent_for_verifiable_first_party_clients, true)
  enforce_policies = coalesce(var.spec.enforce_policies, false)
  token_dialect    = var.spec.token_dialect

  # Scopes
  scopes = coalesce(var.spec.scopes, [])
}
