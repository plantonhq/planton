# Auth0ClientFromMetadataDocument Main Resources
#
# The application is registered in the tenant the provider's credential
# belongs to, which must allow registration from metadata documents
# (Auth0TenantSettings); Auth0 refuses the registration otherwise.
#
# What the provider does with this resource, and why the module sends so
# little:
# - Create POSTs the document's URL to /api/v2/clients/cimd/register. Auth0
#   fetches the document and registers the application from it (name,
#   redirect URIs, logo, keys, application type, grant types, description).
#   The registration is an upsert keyed by the URL, so a URL the tenant has
#   already registered is taken over rather than refused.
# - The provider then PATCHes the client with the settings the configuration
#   declares -- only those. Everything the spec leaves unset renders as null
#   or no block (locals.tf), so Auth0 keeps what the document or the tenant
#   set, and an adopted application plans no change for a setting it never
#   declared.
# - A change of external_client_id_version makes the provider register again
#   (Auth0 fetches the document anew) before the PATCH, so declared settings
#   win over the refreshed document every time.
# - external_client_id forces replacement: a new URL is a new application (a
#   new client_id). Destroy deletes the client.
#
# Every read also asks Auth0 to preview the document, which is where the
# validation outputs come from (outputs.tf).
resource "auth0_client_cimd" "this" {
  external_client_id         = var.spec.external_client_id
  external_client_id_version = var.spec.external_client_id_version

  app_type                    = var.spec.app_type
  grant_types                 = local.grant_types
  description                 = var.spec.description
  allowed_origins             = local.allowed_origins
  web_origins                 = local.web_origins
  oidc_conformant             = var.spec.oidc_conformant
  require_proof_of_possession = var.spec.require_proof_of_possession
  redirection_policy          = var.spec.redirection_policy
  client_metadata             = local.client_metadata

  skip_non_verifiable_callback_uri_confirmation_prompt = var.spec.skip_non_verifiable_callback_uri_confirmation_prompt

  organization_discovery_methods = local.organization_discovery_methods

  dynamic "default_organization" {
    for_each = local.default_organization
    content {
      organization_id = default_organization.value.organization_id
      flows           = default_organization.value.flows
    }
  }

  # Inside a declared block, a field left unset is null: the provider keeps
  # Auth0's value for it (the block's fields are computed).
  dynamic "jwt_configuration" {
    for_each = local.jwt_configuration
    content {
      alg                 = jwt_configuration.value.alg
      lifetime_in_seconds = jwt_configuration.value.lifetime_in_seconds
    }
  }

  dynamic "refresh_token" {
    for_each = local.refresh_token
    content {
      rotation_type                = refresh_token.value.rotation_type
      expiration_type              = refresh_token.value.expiration_type
      leeway                       = refresh_token.value.leeway
      token_lifetime               = refresh_token.value.token_lifetime
      infinite_token_lifetime      = refresh_token.value.infinite_token_lifetime
      idle_token_lifetime          = refresh_token.value.idle_token_lifetime
      infinite_idle_token_lifetime = refresh_token.value.infinite_idle_token_lifetime
    }
  }

  # client_credentials is required inside the block (spec validation). An
  # unset enforce is null, which the provider defaults to true -- the same
  # default the Pulumi engine gets from the bridged provider.
  dynamic "token_quota" {
    for_each = local.token_quota
    content {
      client_credentials {
        enforce  = token_quota.value.client_credentials.enforce
        per_day  = token_quota.value.client_credentials.per_day
        per_hour = token_quota.value.client_credentials.per_hour
      }
    }
  }
}
