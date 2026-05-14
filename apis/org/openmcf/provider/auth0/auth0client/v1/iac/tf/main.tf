# Auth0Client Main Resources
# This file creates the Auth0 client (application)

# Auth0 Client Resource
resource "auth0_client" "this" {
  name        = local.client_name
  app_type    = local.application_type
  description = local.description

  # Logo URI (optional)
  logo_uri = local.logo_uri

  # URLs configuration
  callbacks           = length(local.callbacks) > 0 ? local.callbacks : null
  allowed_logout_urls = length(local.allowed_logout_urls) > 0 ? local.allowed_logout_urls : null
  web_origins         = length(local.web_origins) > 0 ? local.web_origins : null
  allowed_origins     = length(local.allowed_origins) > 0 ? local.allowed_origins : null

  # OAuth configuration
  grant_types     = length(local.grant_types) > 0 ? local.grant_types : null
  oidc_conformant = local.oidc_conformant
  is_first_party  = local.is_first_party

  # Cross-origin settings
  cross_origin_auth = local.cross_origin_authentication
  cross_origin_loc  = local.cross_origin_loc

  # SSO settings
  sso          = local.sso
  sso_disabled = local.sso_disabled

  # Custom login page
  custom_login_page    = local.custom_login_page
  custom_login_page_on = local.custom_login_page_on
  initiate_login_uri   = local.initiate_login_uri

  # Organization settings
  organization_usage           = local.organization_usage
  organization_require_behavior = local.organization_require_behavior

  # Client metadata
  client_metadata = length(local.client_metadata) > 0 ? local.client_metadata : null
  client_aliases  = length(local.client_aliases) > 0 ? local.client_aliases : null

  # Additional settings
  is_token_endpoint_ip_header_trusted = local.is_token_endpoint_ip_header_trusted

  # JWT configuration
  dynamic "jwt_configuration" {
    for_each = local.jwt_configuration != null ? [local.jwt_configuration] : []
    content {
      lifetime_in_seconds = jwt_configuration.value.lifetime_in_seconds
      alg                 = jwt_configuration.value.alg
      secret_encoded      = jwt_configuration.value.secret_encoded
      scopes              = length(jwt_configuration.value.scopes) > 0 ? jwt_configuration.value.scopes : null
    }
  }

  # Refresh token configuration
  dynamic "refresh_token" {
    for_each = local.refresh_token != null ? [local.refresh_token] : []
    content {
      rotation_type                = refresh_token.value.rotation_type
      expiration_type              = refresh_token.value.expiration_type
      token_lifetime               = refresh_token.value.token_lifetime
      idle_token_lifetime          = refresh_token.value.idle_token_lifetime
      infinite_token_lifetime      = refresh_token.value.infinite_token_lifetime
      infinite_idle_token_lifetime = refresh_token.value.infinite_idle_token_lifetime
      leeway                       = refresh_token.value.leeway
    }
  }

  # Native social login configuration (for native apps)
  dynamic "native_social_login" {
    for_each = local.native_social_login != null ? [local.native_social_login] : []
    content {
      dynamic "apple" {
        for_each = native_social_login.value.apple != null ? [native_social_login.value.apple] : []
        content {
          enabled = apple.value.enabled
        }
      }
      dynamic "facebook" {
        for_each = native_social_login.value.facebook != null ? [native_social_login.value.facebook] : []
        content {
          enabled = facebook.value.enabled
        }
      }
    }
  }

  # Mobile configuration (for native apps)
  dynamic "mobile" {
    for_each = local.mobile != null ? [local.mobile] : []
    content {
      dynamic "android" {
        for_each = mobile.value.android != null ? [mobile.value.android] : []
        content {
          app_package_name         = android.value.app_package_name
          sha256_cert_fingerprints = length(android.value.sha256_cert_fingerprints) > 0 ? android.value.sha256_cert_fingerprints : null
        }
      }
      dynamic "ios" {
        for_each = mobile.value.ios != null ? [mobile.value.ios] : []
        content {
          team_id               = ios.value.team_id
          app_bundle_identifier = ios.value.app_bundle_identifier
        }
      }
    }
  }

  # Note: OIDC backchannel logout configuration may not be supported by all provider versions.
  # If your provider version supports it, you can configure it using:
  # oidc_backchannel_logout {
  #   backchannel_logout_urls = local.oidc_backchannel_logout.backchannel_logout_urls
  # }
}

# Auth0 Client Grants - Authorize API access for this client
# Each grant authorizes the client to call a specific API with specified scopes.
# Essential for M2M applications that need to access APIs.
resource "auth0_client_grant" "api_grants" {
  for_each = {
    for idx, grant in local.api_grants : idx => grant
    if grant.audience != null && grant.audience != ""
  }

  client_id = auth0_client.this.id
  audience  = each.value.audience
  scopes    = coalesce(each.value.scopes, [])

  # Organization settings (optional)
  organization_usage    = each.value.organization_usage
  allow_any_organization = coalesce(each.value.allow_any_organization, false)
}

# Data source to read back computed-only attributes (client_secret, token_endpoint_auth_method)
# that the auth0_client resource does not expose in its schema.
# Requires the read:client_keys scope on the Auth0 management API client.
data "auth0_client" "this" {
  client_id = auth0_client.this.client_id
}
