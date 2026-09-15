variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "GcpIdentityPlatformConfig specification"
  type = object({
    project_id = optional(string, "")
    sign_in = optional(object({
      email = optional(object({
        enabled           = optional(bool)
        password_required = optional(bool, false)
      }))
      phone_number = optional(object({
        enabled            = optional(bool)
        test_phone_numbers = optional(map(string), {})
      }))
      anonymous = optional(object({
        enabled = optional(bool)
      }))
      allow_duplicate_emails = optional(bool, false)
    }))
    authorized_domains = optional(list(string), [])
    mfa = optional(object({
      state             = optional(string, "")
      enabled_providers = optional(list(string), [])
      provider_configs = optional(list(object({
        state = optional(string, "")
        totp_provider_config = optional(object({
          adjacent_intervals = optional(number, 0)
        }))
      })), [])
    }))
    blocking_functions = optional(object({
      triggers = list(object({
        event_type   = string
        function_uri = string
      }))
      forward_inbound_credentials = optional(object({
        access_token  = optional(bool, false)
        id_token      = optional(bool, false)
        refresh_token = optional(bool, false)
      }))
    }))
    sign_up_quota = optional(object({
      quota          = optional(number, 0)
      quota_duration = optional(string, "")
      start_time     = optional(string, "")
    }))
    sms_region_config = optional(object({
      allow_by_default = optional(object({
        disallowed_regions = optional(list(string), [])
      }))
      allowlist_only = optional(object({
        allowed_regions = optional(list(string), [])
      }))
    }))
    client_permissions = optional(object({
      disabled_user_signup   = optional(bool, false)
      disabled_user_deletion = optional(bool, false)
    }))
    request_logging_enabled = optional(bool)
    multi_tenant = optional(object({
      allow_tenants           = optional(bool, false)
      default_tenant_location = optional(string, "")
    }))
    autodelete_anonymous_users = optional(bool, false)
    default_supported_idps = optional(list(object({
      idp_id        = string
      client_id     = string
      client_secret = string
      enabled       = optional(bool)
    })), [])
    oauth_idp_configs = optional(list(object({
      name          = string
      display_name  = optional(string, "")
      issuer        = string
      client_id     = string
      client_secret = optional(string, "")
      enabled       = optional(bool)
      response_type = optional(object({
        code     = optional(bool, false)
        id_token = optional(bool, false)
      }))
    })), [])
    inbound_saml_configs = optional(list(object({
      name         = string
      display_name = string
      enabled      = optional(bool)
      idp_config = object({
        idp_entity_id = string
        sso_url       = string
        sign_request  = optional(bool, false)
        idp_certificates = optional(list(object({
          x509_certificate = optional(string, "")
        })), [])
      })
      sp_config = optional(object({
        callback_uri = optional(string, "")
        sp_entity_id = optional(string, "")
      }))
    })), [])
    deletion_policy = optional(string, "")
    adopt_existing  = optional(bool, false)
  })
}
