variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareZeroTrustAccessApplicationSpec defines the Access application"
  # NOTE: StringValueOrRef fields flatten to plain strings and enums to their string
  # names via the proto->tfvars converter; unset fields are omitted (filled null by
  # the optional() object types below).
  type = object({
    account_id = optional(string, "")
    zone_id    = optional(string, "")
    name       = string
    type       = optional(string, "")
    domain     = optional(string, "")

    policies = optional(list(object({
      policy     = string
      precedence = optional(number, 0)
    })), [])

    destinations = optional(list(object({
      type          = optional(string)
      uri           = optional(string)
      cidr          = optional(string)
      hostname      = optional(string)
      l4_protocol   = optional(string)
      port_range    = optional(string)
      vnet_id       = optional(string)
      mcp_server_id = optional(string)
    })), [])

    allowed_idps = optional(list(string), [])

    auto_redirect_to_identity    = optional(bool, false)
    session_duration             = optional(string, "")
    tags                         = optional(list(string), [])
    custom_pages                 = optional(list(string), [])
    app_launcher_visible         = optional(bool)
    skip_app_launcher_login_page = optional(bool, false)
    app_launcher_logo_url        = optional(string, "")
    bg_color                     = optional(string, "")
    header_bg_color              = optional(string, "")
    logo_url                     = optional(string, "")

    landing_page_design = optional(object({
      title             = optional(string)
      message           = optional(string)
      image_url         = optional(string)
      button_color      = optional(string)
      button_text_color = optional(string)
    }))

    footer_links = optional(list(object({
      name = string
      url  = string
    })), [])

    allow_authenticate_via_warp     = optional(bool, false)
    allow_iframe                    = optional(bool, false)
    options_preflight_bypass        = optional(bool, false)
    read_service_tokens_from_header = optional(string, "")
    same_site_cookie_attribute      = optional(string, "")
    service_auth_401_redirect       = optional(bool, false)
    skip_interstitial               = optional(bool, false)
    enable_binding_cookie           = optional(bool, false)
    http_only_cookie_attribute      = optional(bool)
    path_cookie_attribute           = optional(bool, false)
    custom_deny_message             = optional(string, "")
    custom_deny_url                 = optional(string, "")
    custom_non_identity_deny_url    = optional(string, "")

    cors_headers = optional(object({
      allow_all_headers  = optional(bool)
      allow_all_methods  = optional(bool)
      allow_all_origins  = optional(bool)
      allow_credentials  = optional(bool)
      allowed_headers    = optional(list(string))
      allowed_methods    = optional(list(string))
      allowed_origins    = optional(list(string))
      max_age            = optional(number)
    }))

    mfa_config = optional(object({
      allowed_authenticators = optional(list(string))
      mfa_disabled           = optional(bool)
      session_duration       = optional(string)
    }))

    oauth_configuration = optional(object({
      enabled = optional(bool)
      dynamic_client_registration = optional(object({
        enabled               = optional(bool)
        allow_any_on_localhost = optional(bool)
        allow_any_on_loopback  = optional(bool)
        allowed_uris          = optional(list(string))
      }))
      grant = optional(object({
        access_token_lifetime = optional(string)
        session_duration      = optional(string)
      }))
    }))

    target_criteria = optional(list(object({
      port     = number
      protocol = string
      target_attributes = list(object({
        name   = string
        values = list(string)
      }))
    })), [])

    saas_app = optional(object({
      auth_type                        = optional(string)
      consumer_service_url             = optional(string)
      sp_entity_id                     = optional(string)
      name_id_format                   = optional(string)
      name_id_transform_jsonata        = optional(string)
      saml_attribute_transform_jsonata = optional(string)
      default_relay_state              = optional(string)
      custom_attributes = optional(list(object({
        name          = optional(string)
        friendly_name = optional(string)
        name_format   = optional(string)
        required      = optional(bool)
        source = optional(object({
          name = optional(string)
          name_by_idp = optional(list(object({
            idp_id      = optional(string)
            source_name = optional(string)
          })))
        }))
      })))
      redirect_uris                    = optional(list(string))
      grant_types                      = optional(list(string))
      scopes                           = optional(list(string))
      group_filter_regex               = optional(string)
      app_launcher_url                 = optional(string)
      access_token_lifetime            = optional(string)
      allow_pkce_without_client_secret = optional(bool)
      custom_claims = optional(list(object({
        name     = optional(string)
        required = optional(bool)
        scope    = optional(string)
        source = optional(object({
          name        = optional(string)
          name_by_idp = optional(map(string))
        }))
      })))
      hybrid_and_implicit_options = optional(object({
        return_access_token_from_authorization_endpoint = optional(bool)
        return_id_token_from_authorization_endpoint     = optional(bool)
      }))
      refresh_token_options = optional(object({
        lifetime = optional(string)
      }))
    }))

    scim_config = optional(object({
      idp_uid              = string
      remote_uri           = string
      enabled              = optional(bool)
      deactivate_on_delete = optional(bool)
      authentication = optional(object({
        scheme            = string
        user              = optional(string)
        password          = optional(string)
        token             = optional(string)
        client_id         = optional(string)
        client_secret     = optional(string)
        authorization_url = optional(string)
        token_url         = optional(string)
        scopes            = optional(list(string))
      }))
      mappings = optional(list(object({
        schema            = string
        enabled           = optional(bool)
        filter            = optional(string)
        strictness        = optional(string)
        transform_jsonata = optional(string)
        operations = optional(object({
          create = optional(bool)
          update = optional(bool)
          delete = optional(bool)
        }))
      })))
    }))
  })
}
