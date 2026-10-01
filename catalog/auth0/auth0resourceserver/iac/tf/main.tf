# Auth0ResourceServer Main Resources
# This file creates the Auth0 Resource Server (API), its scopes, and the
# default grants every third-party application gets on it.

# Auth0 Resource Server Resource -- the twin of the Pulumi module's
# resourceserver.go. Every setting is sent only when the spec declares it
# (null attributes and absent blocks are never sent), so an API adopted into
# this kind keeps what Auth0 holds for it.
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
  enforce_policies                                = local.enforce_policies
  token_dialect                                   = local.token_dialect

  # Unmanaged when unset
  allow_online_access                         = var.spec.allow_online_access
  allow_online_access_with_ephemeral_sessions = var.spec.allow_online_access_with_ephemeral_sessions
  consent_policy                              = var.spec.consent_policy
  token_lifetime_for_anonymous_access_tokens  = var.spec.token_lifetime_for_anonymous_access_tokens
  verification_location                       = var.spec.verification_location
  signing_secret                              = local.signing_secret

  # A declared claims_mapping is sent with its whole claim list: zero
  # custom_claims blocks clear the claims.
  dynamic "access_token" {
    for_each = local.access_token != null ? [local.access_token] : []
    content {
      dynamic "claims_mapping" {
        for_each = access_token.value.claims_mapping != null ? [access_token.value.claims_mapping] : []
        content {
          dynamic "custom_claims" {
            for_each = claims_mapping.value.custom_claims
            content {
              name       = custom_claims.value.name
              expression = custom_claims.value.expression
            }
          }
        }
      }
    }
  }

  # The Rich Authorization Request types, or the single entry that disables
  # them. No entries: the API keeps the types it has.
  dynamic "authorization_details" {
    for_each = local.authorization_details
    content {
      type    = authorization_details.value.type
      disable = authorization_details.value.disable
    }
  }

  dynamic "authorization_policy" {
    for_each = local.authorization_policy != null ? [local.authorization_policy] : []
    content {
      policy_id = authorization_policy.value.policy_id
    }
  }

  dynamic "proof_of_possession" {
    for_each = local.proof_of_possession != null ? [local.proof_of_possession] : []
    content {
      disable      = proof_of_possession.value.disable
      mechanism    = proof_of_possession.value.mechanism
      required     = proof_of_possession.value.required
      required_for = proof_of_possession.value.required_for
    }
  }

  # The access policy. Each subject's block is rendered only with its policy
  # (locals.tf), so a policy the spec leaves out keeps its live value.
  dynamic "subject_type_authorization" {
    for_each = local.subject_type_authorization != null ? [local.subject_type_authorization] : []
    content {
      dynamic "user" {
        for_each = local.subject_type_user_policy != null ? [local.subject_type_user_policy] : []
        content {
          policy = user.value
        }
      }
      dynamic "client" {
        for_each = local.subject_type_client_policy != null ? [local.subject_type_client_policy] : []
        content {
          policy = client.value
        }
      }
      dynamic "anonymous_user" {
        for_each = local.subject_type_anonymous_user_policy != null ? [local.subject_type_anonymous_user_policy] : []
        content {
          policy = anonymous_user.value
        }
      }
    }
  }

  # Token encryption: the format and the API's public key, or disable.
  dynamic "token_encryption" {
    for_each = local.token_encryption != null ? [local.token_encryption] : []
    content {
      disable = token_encryption.value.disable
      format  = token_encryption.value.format

      dynamic "encryption_key" {
        for_each = token_encryption.value.encryption_key != null ? [token_encryption.value.encryption_key] : []
        content {
          algorithm = encryption_key.value.algorithm
          pem       = encryption_key.value.pem
          kid       = encryption_key.value.kid
          name      = encryption_key.value.name
        }
      }
    }
  }
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

# Default grants for third-party applications -- the twin of the Pulumi
# module's default_grants.go. A default grant names no application
# (default_for replaces client_id): every third-party application gets its
# scopes on this API without a grant of its own, and a grant made for one
# application takes precedence. One per subject type, keyed by it: the key each
# grant's id is exported and imported under. The grants follow the scopes,
# since a grant may only name scopes the API defines. Destroy deletes them.
resource "auth0_client_grant" "third_party_client_default_grants" {
  for_each = local.third_party_client_default_grants

  audience     = auth0_resource_server.this.identifier
  default_for  = "third_party_clients"
  subject_type = each.key

  scopes                      = each.value.scopes
  authorization_details_types = each.value.authorization_details_types
  allow_all_scopes            = each.value.allow_all_scopes
  organization_usage          = each.value.organization_usage
  allow_any_organization      = each.value.allow_any_organization

  depends_on = [auth0_resource_server_scopes.this]
}
