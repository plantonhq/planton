# Auth0ResourceServer Locals
# This file contains local variables computed from input variables.
# The Pulumi module's locals.go and resourceserver.go apply the same rules --
# keep them in lockstep.

locals {
  # Core configuration
  resource_name = var.metadata.name
  identifier    = var.spec.identifier
  name          = coalesce(var.spec.name, var.metadata.name)

  # Token settings
  signing_alg            = var.spec.signing_alg
  allow_offline_access   = var.spec.allow_offline_access
  token_lifetime         = var.spec.token_lifetime
  token_lifetime_for_web = var.spec.token_lifetime_for_web

  # Access control settings
  skip_consent_for_verifiable_first_party_clients = var.spec.skip_consent_for_verifiable_first_party_clients
  enforce_policies                                = var.spec.enforce_policies
  token_dialect                                   = var.spec.token_dialect

  # Scopes
  scopes = coalesce(var.spec.scopes, [])

  # The settings below are unmanaged when unset. The proto's optional fields
  # carry presence, so tfvars renders an unset one as null, which the provider
  # never sends; a declared false, zero or empty value is sent as declared.
  # Blocks are rendered only when declared (main.tf's dynamic blocks).

  # A shared signing key: marked sensitive so plans never print it.
  signing_secret = var.spec.signing_secret != null ? sensitive(var.spec.signing_secret) : null

  access_token               = var.spec.access_token
  authorization_details      = var.spec.authorization_details != null ? var.spec.authorization_details : []
  authorization_policy       = var.spec.authorization_policy
  proof_of_possession        = var.spec.proof_of_possession
  token_encryption           = var.spec.token_encryption
  subject_type_authorization = var.spec.subject_type_authorization

  # Each subject's policy block is rendered only with its policy: the provider
  # keeps a block the configuration leaves out, so a policy never declared is
  # never touched.
  subject_type_user_policy           = try(local.subject_type_authorization.user.policy, null)
  subject_type_client_policy         = try(local.subject_type_authorization.client.policy, null)
  subject_type_anonymous_user_policy = try(local.subject_type_authorization.anonymous_user.policy, null)

  # The default grants for third-party applications, keyed by subject type --
  # the key each grant's id is exported and imported under. scopes is sent
  # unless allow_all_scopes is true (the provider refuses both).
  third_party_client_default_grants = {
    for grant in(var.spec.third_party_client_default_grants != null ? var.spec.third_party_client_default_grants : []) :
    grant.subject_type => {
      scopes                      = grant.allow_all_scopes == true ? null : grant.scopes
      authorization_details_types = length(grant.authorization_details_types) > 0 ? grant.authorization_details_types : null
      allow_all_scopes            = grant.allow_all_scopes
      organization_usage          = grant.organization_usage
      allow_any_organization      = grant.allow_any_organization
    }
  }
}
