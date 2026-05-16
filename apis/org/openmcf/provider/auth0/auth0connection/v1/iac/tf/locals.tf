# Local values for Auth0Connection module
# These values are computed from the input variables

locals {
  # Core connection configuration
  connection_name     = var.metadata.name
  strategy            = var.spec.strategy
  display_name        = coalesce(var.spec.display_name, var.metadata.name)
  is_domain_connection = coalesce(var.spec.is_domain_connection, false)
  show_as_button      = coalesce(var.spec.show_as_button, true)

  # enabled_clients values are already flattened to plain strings by the tfvars generator.
  enabled_clients = var.spec.enabled_clients != null ? [
    for client in var.spec.enabled_clients : client
    if client != null && client != ""
  ] : []

  # Realms - default to empty list if not specified
  realms = coalesce(var.spec.realms, [])

  # Connection metadata
  connection_metadata = coalesce(var.spec.metadata, {})

  # Determine which options block to use based on strategy
  is_database_strategy = var.spec.strategy == "auth0"
  is_social_strategy = contains([
    "google-oauth2", "facebook", "github", "linkedin",
    "twitter", "microsoft-account", "apple"
  ], var.spec.strategy)
  is_saml_strategy   = var.spec.strategy == "samlp"
  is_oidc_strategy   = var.spec.strategy == "oidc"
  is_waad_strategy   = var.spec.strategy == "waad"

  # Database options with defaults
  database_options = var.spec.database_options != null ? {
    password_policy           = coalesce(var.spec.database_options.password_policy, "good")
    requires_username         = coalesce(var.spec.database_options.requires_username, false)
    disable_signup            = coalesce(var.spec.database_options.disable_signup, false)
    brute_force_protection    = coalesce(var.spec.database_options.brute_force_protection, true)
    password_history_size     = coalesce(var.spec.database_options.password_history_size, 5)
    password_no_personal_info = coalesce(var.spec.database_options.password_no_personal_info, true)
    password_dictionary       = coalesce(var.spec.database_options.password_dictionary, true)
    mfa_enabled               = coalesce(var.spec.database_options.mfa_enabled, false)
  } : null

  # Social options
  social_options = var.spec.social_options != null ? {
    client_id         = var.spec.social_options.client_id
    client_secret     = var.spec.social_options.client_secret
    scopes            = coalesce(var.spec.social_options.scopes, [])
    allowed_audiences = coalesce(var.spec.social_options.allowed_audiences, [])
    upstream_params   = coalesce(var.spec.social_options.upstream_params, {})
  } : null

  # SAML options
  saml_options = var.spec.saml_options != null ? {
    sign_in_endpoint    = var.spec.saml_options.sign_in_endpoint
    signing_cert        = var.spec.saml_options.signing_cert
    sign_out_endpoint   = var.spec.saml_options.sign_out_endpoint
    entity_id           = var.spec.saml_options.entity_id
    protocol_binding    = var.spec.saml_options.protocol_binding
    user_id_attribute   = var.spec.saml_options.user_id_attribute
    sign_request        = coalesce(var.spec.saml_options.sign_request, false)
    signature_algorithm = coalesce(var.spec.saml_options.signature_algorithm, "rsa-sha256")
    digest_algorithm    = coalesce(var.spec.saml_options.digest_algorithm, "sha256")
    attribute_mappings  = coalesce(var.spec.saml_options.attribute_mappings, {})
  } : null

  # OIDC options
  oidc_options = var.spec.oidc_options != null ? {
    issuer                 = var.spec.oidc_options.issuer
    client_id              = var.spec.oidc_options.client_id
    client_secret          = var.spec.oidc_options.client_secret
    scopes                 = coalesce(var.spec.oidc_options.scopes, ["openid", "profile", "email"])
    type                   = coalesce(var.spec.oidc_options.type, "front_channel")
    authorization_endpoint = var.spec.oidc_options.authorization_endpoint
    token_endpoint         = var.spec.oidc_options.token_endpoint
    userinfo_endpoint      = var.spec.oidc_options.userinfo_endpoint
    jwks_uri               = var.spec.oidc_options.jwks_uri
    attribute_mappings     = coalesce(var.spec.oidc_options.attribute_mappings, {})
  } : null

  # Azure AD options
  azure_ad_options = var.spec.azure_ad_options != null ? {
    client_id                   = var.spec.azure_ad_options.client_id
    client_secret               = var.spec.azure_ad_options.client_secret
    domain                      = var.spec.azure_ad_options.domain
    tenant_id                   = var.spec.azure_ad_options.tenant_id
    use_common_endpoint         = coalesce(var.spec.azure_ad_options.use_common_endpoint, false)
    max_groups_to_retrieve      = coalesce(var.spec.azure_ad_options.max_groups_to_retrieve, 50)
    should_trust_email_verified = coalesce(var.spec.azure_ad_options.should_trust_email_verified, true)
    api_enable_users            = coalesce(var.spec.azure_ad_options.api_enable_users, false)
  } : null
}

