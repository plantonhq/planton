# Auth0TenantSettings Main Resources
#
# auth0_tenant manages the settings of the EXISTING tenant the provider's
# credential belongs to; the Management API cannot create or delete a tenant.
# Every argument is null and every block absent unless the spec sets it, so the
# provider never sends what the spec leaves unset and the tenant keeps its
# value; a declared block carries only the fields the spec sets inside it. A
# spec that manages only the default domain sends no tenant setting.
#
# The provider itself still writes six settings the spec leaves unset on the
# first deploy (and after an import): default_redirection_uri,
# skip_non_verifiable_callback_uri_confirmation_prompt, mtls, error_page,
# default_token_quota and country_codes -- see the spec's comments. Nothing
# here can prevent that; the manifest declares the live values instead.
#
# Destroy drops the resource from state and leaves the last-applied values in
# place: Auth0 has no delete for tenant settings. The Pulumi module's
# applyTenantSettings (iac/pulumi/module/tenant.go) is its twin.
resource "auth0_tenant" "this" {
  count = local.manages_tenant_settings ? 1 : 0

  # Identity.
  friendly_name   = local.friendly_name
  picture_url     = local.picture_url
  support_email   = local.support_email
  support_url     = local.support_url
  enabled_locales = local.enabled_locales
  sandbox_version = var.spec.sandbox_version

  # Sessions.
  session_lifetime                = var.spec.session_lifetime
  idle_session_lifetime           = var.spec.idle_session_lifetime
  ephemeral_session_lifetime      = var.spec.ephemeral_session_lifetime
  idle_ephemeral_session_lifetime = var.spec.idle_ephemeral_session_lifetime

  dynamic "session_cookie" {
    for_each = var.spec.session_cookie != null ? [var.spec.session_cookie] : []
    content {
      mode = session_cookie.value.mode
    }
  }

  # The anonymous block is sent only when declared; the provider removes the
  # tenant's anonymous-session settings when sessions is declared without it.
  dynamic "sessions" {
    for_each = var.spec.sessions != null ? [var.spec.sessions] : []
    content {
      oidc_logout_prompt_enabled = sessions.value.oidc_logout_prompt_enabled

      dynamic "anonymous" {
        for_each = sessions.value.anonymous != null ? [sessions.value.anonymous] : []
        content {
          activate_cookie     = anonymous.value.activate_cookie
          lifetime_in_minutes = anonymous.value.lifetime_in_minutes
        }
      }
    }
  }

  # OAuth and OpenID Connect.
  client_id_metadata_document_supported                = var.spec.client_id_metadata_document_supported
  resource_parameter_profile                           = var.spec.resource_parameter_profile
  dynamic_client_registration_security_mode            = var.spec.dynamic_client_registration_security_mode
  pushed_authorization_requests_supported              = var.spec.pushed_authorization_requests_supported
  acr_values_supported                                 = local.acr_values_supported
  disable_acr_values_supported                         = var.spec.disable_acr_values_supported
  allow_organization_name_in_authentication_api        = var.spec.allow_organization_name_in_authentication_api
  default_redirection_uri                              = var.spec.default_redirection_uri
  allowed_logout_urls                                  = local.allowed_logout_urls
  skip_non_verifiable_callback_uri_confirmation_prompt = local.skip_non_verifiable_callback_uri_confirmation_prompt

  dynamic "oidc_logout" {
    for_each = var.spec.oidc_logout != null ? [var.spec.oidc_logout] : []
    content {
      rp_logout_end_session_endpoint_discovery = oidc_logout.value.rp_logout_end_session_endpoint_discovery
    }
  }

  dynamic "mtls" {
    for_each = var.spec.mtls != null ? [var.spec.mtls] : []
    content {
      disable                 = mtls.value.disable
      enable_endpoint_aliases = mtls.value.enable_endpoint_aliases
    }
  }

  # Defaults.
  default_audience  = local.default_audience
  default_directory = local.default_directory

  # A declared quota with neither clients nor organizations is still sent
  # (empty), which the provider turns into "remove the tenant's default
  # quotas".
  dynamic "default_token_quota" {
    for_each = var.spec.default_token_quota != null ? [var.spec.default_token_quota] : []
    content {
      dynamic "clients" {
        for_each = default_token_quota.value.clients != null ? [default_token_quota.value.clients] : []
        content {
          client_credentials {
            enforce  = clients.value.client_credentials.enforce
            per_day  = clients.value.client_credentials.per_day
            per_hour = clients.value.client_credentials.per_hour
          }
        }
      }

      dynamic "organizations" {
        for_each = default_token_quota.value.organizations != null ? [default_token_quota.value.organizations] : []
        content {
          client_credentials {
            enforce  = organizations.value.client_credentials.enforce
            per_day  = organizations.value.client_credentials.per_day
            per_hour = organizations.value.client_credentials.per_hour
          }
        }
      }
    }
  }

  # Sign-in and MFA.
  customize_mfa_in_postlogin_action = var.spec.customize_mfa_in_postlogin_action
  phone_consolidated_experience     = var.spec.phone_consolidated_experience

  dynamic "country_codes" {
    for_each = var.spec.country_codes != null ? [var.spec.country_codes] : []
    content {
      list = country_codes.value.list
      mode = country_codes.value.mode
    }
  }

  # The error page. A declared page with no html, url or show_log_link true
  # returns the tenant to Auth0's default page.
  dynamic "error_page" {
    for_each = var.spec.error_page != null ? [var.spec.error_page] : []
    content {
      html          = error_page.value.html
      show_log_link = error_page.value.show_log_link
      url           = error_page.value.url
    }
  }

  # Behavior flags: only the flags the spec sets are sent; every other flag
  # keeps the tenant's value. The provider sends enable_sso only when it
  # changes.
  dynamic "flags" {
    for_each = var.spec.flags != null ? [var.spec.flags] : []
    content {
      allow_legacy_delegation_grant_types    = flags.value.allow_legacy_delegation_grant_types
      allow_legacy_ro_grant_types            = flags.value.allow_legacy_ro_grant_types
      allow_legacy_tokeninfo_endpoint        = flags.value.allow_legacy_tokeninfo_endpoint
      dashboard_insights_view                = flags.value.dashboard_insights_view
      dashboard_log_streams_next             = flags.value.dashboard_log_streams_next
      disable_clickjack_protection_headers   = flags.value.disable_clickjack_protection_headers
      disable_fields_map_fix                 = flags.value.disable_fields_map_fix
      disable_management_api_sms_obfuscation = flags.value.disable_management_api_sms_obfuscation
      enable_adfs_waad_email_verification    = flags.value.enable_adfs_waad_email_verification
      enable_apis_section                    = flags.value.enable_apis_section
      enable_client_connections              = flags.value.enable_client_connections
      enable_custom_domain_in_emails         = flags.value.enable_custom_domain_in_emails
      enable_dynamic_client_registration     = flags.value.enable_dynamic_client_registration
      enable_idtoken_api2                    = flags.value.enable_idtoken_api2
      enable_legacy_logs_search_v2           = flags.value.enable_legacy_logs_search_v2
      enable_legacy_profile                  = flags.value.enable_legacy_profile
      enable_pipeline2                       = flags.value.enable_pipeline2
      enable_public_signup_user_exists_error = flags.value.enable_public_signup_user_exists_error
      enable_sso                             = flags.value.enable_sso
      mfa_show_factor_list_on_enrollment     = flags.value.mfa_show_factor_list_on_enrollment
      no_disclose_enterprise_connections     = flags.value.no_disclose_enterprise_connections
      remove_alg_from_jwks                   = flags.value.remove_alg_from_jwks
      revoke_refresh_token_grant             = flags.value.revoke_refresh_token_grant
      use_scope_descriptions_for_consent     = flags.value.use_scope_descriptions_for_consent
    }
  }
}

# A tenant installed before the resource was counted keeps its state: the same
# object, now at index 0.
moved {
  from = auth0_tenant.this
  to   = auth0_tenant.this[0]
}

# The tenant's settings, read without writing, when the spec declares none of
# them (the default domain alone, or nothing).
data "auth0_tenant" "current" {
  count = local.manages_tenant_settings ? 0 : 1
}

# The tenant's default domain -- the one its email links and Management API
# notifications use -- declared only when the spec manages it. Auth0 has no way
# to unset a default, so the resource's delete only forgets it: destroy leaves
# the last-applied default in place.
resource "auth0_custom_domain_default" "this" {
  count  = local.default_custom_domain != null ? 1 : 0
  domain = local.default_custom_domain
}
