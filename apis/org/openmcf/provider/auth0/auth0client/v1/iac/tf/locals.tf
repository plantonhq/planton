# Local values for Auth0Client module
# These values are computed from the input variables

locals {
  # Core client configuration.
  # NOTE: optional spec fields with no variables.tf default arrive as null when the
  # caller omits them; `coalesce` rejects null AND empty values ("no non-null,
  # non-empty-string arguments"), so a client that omits e.g. description failed at
  # plan. Use a null-test instead so the empty default is honored without erroring.
  client_name      = var.metadata.name
  application_type = var.spec.application_type
  description      = var.spec.description != null ? var.spec.description : ""
  logo_uri         = var.spec.logo_uri

  # URLs configuration
  callbacks           = var.spec.callbacks != null ? var.spec.callbacks : []
  allowed_logout_urls = var.spec.allowed_logout_urls != null ? var.spec.allowed_logout_urls : []
  web_origins         = var.spec.web_origins != null ? var.spec.web_origins : []
  allowed_origins     = var.spec.allowed_origins != null ? var.spec.allowed_origins : []

  # OAuth configuration
  grant_types     = var.spec.grant_types != null ? var.spec.grant_types : []
  oidc_conformant = coalesce(var.spec.oidc_conformant, true)
  is_first_party  = coalesce(var.spec.is_first_party, true)

  # Cross-origin settings
  cross_origin_authentication = coalesce(var.spec.cross_origin_authentication, false)
  cross_origin_loc            = var.spec.cross_origin_loc

  # SSO settings
  sso          = coalesce(var.spec.sso, true)
  sso_disabled = coalesce(var.spec.sso_disabled, false)

  # Custom login page
  custom_login_page    = var.spec.custom_login_page
  custom_login_page_on = coalesce(var.spec.custom_login_page_on, false)
  initiate_login_uri   = var.spec.initiate_login_uri

  # Organization settings
  organization_usage           = var.spec.organization_usage
  organization_require_behavior = var.spec.organization_require_behavior

  # Client metadata
  client_metadata = var.spec.client_metadata != null ? var.spec.client_metadata : {}
  client_aliases  = var.spec.client_aliases != null ? var.spec.client_aliases : []

  # Additional settings
  is_token_endpoint_ip_header_trusted = coalesce(var.spec.is_token_endpoint_ip_header_trusted, false)

  # enabled_connections values are already flattened to plain strings by the tfvars generator.
  enabled_connections = [
    for conn in(var.spec.enabled_connections != null ? var.spec.enabled_connections : []) : conn
    if conn != null && conn != ""
  ]

  # JWT configuration with defaults
  jwt_configuration = var.spec.jwt_configuration != null ? {
    lifetime_in_seconds = var.spec.jwt_configuration.lifetime_in_seconds
    scopes              = var.spec.jwt_configuration.scopes != null ? var.spec.jwt_configuration.scopes : {}
    alg                 = var.spec.jwt_configuration.alg
    secret_encoded      = coalesce(var.spec.jwt_configuration.secret_encoded, false)
  } : null

  # Refresh token configuration with defaults
  refresh_token = var.spec.refresh_token != null ? {
    rotation_type                = var.spec.refresh_token.rotation_type
    expiration_type              = var.spec.refresh_token.expiration_type
    token_lifetime               = var.spec.refresh_token.token_lifetime
    idle_token_lifetime          = var.spec.refresh_token.idle_token_lifetime
    infinite_token_lifetime      = coalesce(var.spec.refresh_token.infinite_token_lifetime, false)
    infinite_idle_token_lifetime = coalesce(var.spec.refresh_token.infinite_idle_token_lifetime, false)
    leeway                       = var.spec.refresh_token.leeway
  } : null

  # Native social login configuration
  native_social_login = var.spec.native_social_login != null ? {
    apple = var.spec.native_social_login.apple != null ? {
      enabled = coalesce(var.spec.native_social_login.apple.enabled, false)
    } : null
    facebook = var.spec.native_social_login.facebook != null ? {
      enabled = coalesce(var.spec.native_social_login.facebook.enabled, false)
    } : null
  } : null

  # Mobile configuration
  mobile = var.spec.mobile != null ? {
    android = var.spec.mobile.android != null ? {
      app_package_name         = var.spec.mobile.android.app_package_name
      sha256_cert_fingerprints = var.spec.mobile.android.sha256_cert_fingerprints != null ? var.spec.mobile.android.sha256_cert_fingerprints : []
    } : null
    ios = var.spec.mobile.ios != null ? {
      team_id               = var.spec.mobile.ios.team_id
      app_bundle_identifier = var.spec.mobile.ios.app_bundle_identifier
    } : null
  } : null

  # OIDC backchannel logout
  oidc_backchannel_logout = var.spec.oidc_backchannel_logout != null ? {
    backchannel_logout_urls = var.spec.oidc_backchannel_logout.backchannel_logout_urls != null ? var.spec.oidc_backchannel_logout.backchannel_logout_urls : []
  } : null

  # API grants for authorizing API access
  # audience is already flattened to a plain string by the tfvars generator.
  api_grants = [
    for grant in(var.spec.api_grants != null ? var.spec.api_grants : []) : {
      audience               = grant.audience
      scopes                 = grant.scopes != null ? grant.scopes : []
      allow_any_organization = coalesce(grant.allow_any_organization, false)
      organization_usage     = grant.organization_usage
    }
    if grant != null && grant.audience != null && grant.audience != ""
  ]
}


