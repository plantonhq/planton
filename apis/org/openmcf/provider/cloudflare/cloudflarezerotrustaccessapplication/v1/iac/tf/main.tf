# Cloudflare Zero Trust Access application: the protected resource (self-hosted web
# app, SaaS app, SSH/VNC/RDP target, app launcher, MCP endpoint, ...) guarded by
# Cloudflare Access. Policies are attached by reference.
resource "cloudflare_zero_trust_access_application" "main" {
  account_id = var.spec.account_id != "" ? var.spec.account_id : null
  zone_id    = var.spec.zone_id != "" ? var.spec.zone_id : null

  name   = var.spec.name
  type   = local.app_type
  domain = var.spec.domain != "" ? var.spec.domain : null

  policies     = length(local.policies) > 0 ? local.policies : null
  destinations = length(try(var.spec.destinations, [])) > 0 ? var.spec.destinations : null
  allowed_idps = length(try(var.spec.allowed_idps, [])) > 0 ? var.spec.allowed_idps : null

  # Type-restricted toggles: the provider rejects these (even when false) on
  # incompatible application types, so send them only when enabled (false == the
  # provider default == omitted).
  auto_redirect_to_identity    = var.spec.auto_redirect_to_identity ? true : null
  session_duration             = var.spec.session_duration != "" ? var.spec.session_duration : null
  tags                         = length(try(var.spec.tags, [])) > 0 ? var.spec.tags : null
  custom_pages                 = length(try(var.spec.custom_pages, [])) > 0 ? var.spec.custom_pages : null
  app_launcher_visible         = var.spec.app_launcher_visible
  skip_app_launcher_login_page = var.spec.skip_app_launcher_login_page ? true : null
  app_launcher_logo_url        = var.spec.app_launcher_logo_url != "" ? var.spec.app_launcher_logo_url : null
  bg_color                     = var.spec.bg_color != "" ? var.spec.bg_color : null
  header_bg_color              = var.spec.header_bg_color != "" ? var.spec.header_bg_color : null
  logo_url                     = var.spec.logo_url != "" ? var.spec.logo_url : null
  landing_page_design          = var.spec.landing_page_design
  footer_links                 = length(try(var.spec.footer_links, [])) > 0 ? var.spec.footer_links : null

  allow_authenticate_via_warp     = var.spec.allow_authenticate_via_warp ? true : null
  allow_iframe                    = var.spec.allow_iframe ? true : null
  options_preflight_bypass        = var.spec.options_preflight_bypass ? true : null
  read_service_tokens_from_header = var.spec.read_service_tokens_from_header != "" ? var.spec.read_service_tokens_from_header : null
  same_site_cookie_attribute      = var.spec.same_site_cookie_attribute != "" ? var.spec.same_site_cookie_attribute : null
  service_auth_401_redirect       = var.spec.service_auth_401_redirect ? true : null
  skip_interstitial               = var.spec.skip_interstitial ? true : null
  enable_binding_cookie           = var.spec.enable_binding_cookie ? true : null
  http_only_cookie_attribute      = var.spec.http_only_cookie_attribute
  path_cookie_attribute           = var.spec.path_cookie_attribute ? true : null
  custom_deny_message             = var.spec.custom_deny_message != "" ? var.spec.custom_deny_message : null
  custom_deny_url                 = var.spec.custom_deny_url != "" ? var.spec.custom_deny_url : null
  custom_non_identity_deny_url    = var.spec.custom_non_identity_deny_url != "" ? var.spec.custom_non_identity_deny_url : null

  cors_headers        = var.spec.cors_headers
  mfa_config          = var.spec.mfa_config
  oauth_configuration = var.spec.oauth_configuration
  target_criteria     = length(local.target_criteria) > 0 ? local.target_criteria : null
  saas_app            = var.spec.saas_app
  scim_config         = var.spec.scim_config
}
