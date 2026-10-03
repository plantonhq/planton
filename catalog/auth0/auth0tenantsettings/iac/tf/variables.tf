variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "Auth0TenantSettings specification"
  type = object({
    friendly_name                   = optional(string, "")
    picture_url                     = optional(string, "")
    support_email                   = optional(string, "")
    support_url                     = optional(string, "")
    default_custom_domain           = optional(string, "")
    enabled_locales                 = optional(list(string), [])
    sandbox_version                 = optional(string)
    session_lifetime                = optional(number)
    idle_session_lifetime           = optional(number)
    ephemeral_session_lifetime      = optional(number)
    idle_ephemeral_session_lifetime = optional(number)
    session_cookie = optional(object({
      mode = optional(string)
    }))
    sessions = optional(object({
      oidc_logout_prompt_enabled = bool
      anonymous = optional(object({
        activate_cookie     = optional(bool)
        lifetime_in_minutes = optional(number)
      }))
    }))
    client_id_metadata_document_supported         = optional(bool)
    resource_parameter_profile                    = optional(string)
    dynamic_client_registration_security_mode     = optional(string)
    pushed_authorization_requests_supported       = optional(bool)
    acr_values_supported                          = optional(list(string), [])
    disable_acr_values_supported                  = optional(bool)
    allow_organization_name_in_authentication_api = optional(bool)
    default_redirection_uri                       = optional(string)
    allowed_logout_urls                           = optional(list(string), [])
    oidc_logout = optional(object({
      rp_logout_end_session_endpoint_discovery = bool
    }))
    mtls = optional(object({
      disable                 = optional(bool)
      enable_endpoint_aliases = optional(bool)
    }))
    skip_non_verifiable_callback_uri_confirmation_prompt = optional(bool)
    default_audience                                     = optional(string, "")
    default_directory                                    = optional(string, "")
    default_token_quota = optional(object({
      clients = optional(object({
        client_credentials = object({
          enforce  = optional(bool)
          per_day  = optional(number)
          per_hour = optional(number)
        })
      }))
      organizations = optional(object({
        client_credentials = object({
          enforce  = optional(bool)
          per_day  = optional(number)
          per_hour = optional(number)
        })
      }))
    }))
    customize_mfa_in_postlogin_action = optional(bool)
    phone_consolidated_experience     = optional(bool)
    country_codes = optional(object({
      list = list(string)
      mode = string
    }))
    error_page = optional(object({
      html          = optional(string)
      show_log_link = optional(bool)
      url           = optional(string)
    }))
    flags = optional(object({
      allow_legacy_delegation_grant_types    = optional(bool)
      allow_legacy_ro_grant_types            = optional(bool)
      allow_legacy_tokeninfo_endpoint        = optional(bool)
      dashboard_insights_view                = optional(bool)
      dashboard_log_streams_next             = optional(bool)
      disable_clickjack_protection_headers   = optional(bool)
      disable_fields_map_fix                 = optional(bool)
      disable_management_api_sms_obfuscation = optional(bool)
      enable_adfs_waad_email_verification    = optional(bool)
      enable_apis_section                    = optional(bool)
      enable_client_connections              = optional(bool)
      enable_custom_domain_in_emails         = optional(bool)
      enable_dynamic_client_registration     = optional(bool)
      enable_idtoken_api2                    = optional(bool)
      enable_legacy_logs_search_v2           = optional(bool)
      enable_legacy_profile                  = optional(bool)
      enable_pipeline2                       = optional(bool)
      enable_public_signup_user_exists_error = optional(bool)
      enable_sso                             = optional(bool)
      mfa_show_factor_list_on_enrollment     = optional(bool)
      no_disclose_enterprise_connections     = optional(bool)
      remove_alg_from_jwks                   = optional(bool)
      revoke_refresh_token_grant             = optional(bool)
      use_scope_descriptions_for_consent     = optional(bool)
    }))
  })
}
