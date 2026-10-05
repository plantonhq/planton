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
  description = "AzureFunctionAppFlexConsumption specification"
  type = object({
    region                            = string
    resource_group                    = string
    function_app_name                 = string
    service_plan_id                   = string
    storage_container_endpoint        = string
    storage_authentication_type       = string
    storage_access_key                = optional(string, "")
    storage_user_assigned_identity_id = optional(string, "")
    runtime_name                      = string
    runtime_version                   = string
    instance_memory_in_mb             = optional(number)
    maximum_instance_count            = optional(number)
    http_concurrency                  = optional(number)
    always_ready = optional(list(object({
      name           = string
      instance_count = optional(number)
    })), [])
    site_config = object({
      api_management_api_id    = optional(string, "")
      api_definition_url       = optional(string, "")
      app_command_line         = optional(string, "")
      application_insights_key = optional(string, "")
      app_service_logs = optional(object({
        disk_quota_mb         = optional(number)
        retention_period_days = optional(number)
      }))
      container_registry_use_managed_identity = optional(bool)
      default_documents                       = optional(list(string), [])
      elastic_instance_minimum                = optional(number)
      http2_enabled                           = optional(bool)
      ip_restrictions = optional(list(object({
        name                      = optional(string, "")
        priority                  = optional(number)
        action                    = optional(string, "")
        ip_address                = optional(string, "")
        service_tag               = optional(string, "")
        virtual_network_subnet_id = optional(string, "")
        description               = optional(string, "")
        headers = optional(object({
          x_forwarded_for   = optional(list(string), [])
          x_forwarded_host  = optional(list(string), [])
          x_azure_fdid      = optional(list(string), [])
          x_fd_health_probe = optional(list(string), [])
        }))
      })), [])
      ip_restriction_default_action = optional(string, "")
      scm_use_main_ip_restriction   = optional(bool)
      scm_ip_restrictions = optional(list(object({
        name                      = optional(string, "")
        priority                  = optional(number)
        action                    = optional(string, "")
        ip_address                = optional(string, "")
        service_tag               = optional(string, "")
        virtual_network_subnet_id = optional(string, "")
        description               = optional(string, "")
        headers = optional(object({
          x_forwarded_for   = optional(list(string), [])
          x_forwarded_host  = optional(list(string), [])
          x_azure_fdid      = optional(list(string), [])
          x_fd_health_probe = optional(list(string), [])
        }))
      })), [])
      scm_ip_restriction_default_action = optional(string, "")
      load_balancing_mode               = optional(string, "")
      managed_pipeline_mode             = optional(string, "")
      remote_debugging_enabled          = optional(bool)
      remote_debugging_version          = optional(string, "")
      runtime_scale_monitoring_enabled  = optional(bool)
      websockets_enabled                = optional(bool)
      health_check_path                 = optional(string, "")
      health_check_eviction_time_in_min = optional(number)
      worker_count                      = optional(number)
      minimum_tls_version               = optional(string, "")
      scm_minimum_tls_version           = optional(string, "")
      cors = optional(object({
        allowed_origins     = list(string)
        support_credentials = optional(bool)
      }))
      vnet_route_all_enabled = optional(bool)
    })
    app_settings = optional(map(string), {})
    connection_strings = optional(list(object({
      name  = string
      type  = string
      value = string
    })), [])
    sticky_settings = optional(object({
      app_setting_names       = optional(list(string), [])
      connection_string_names = optional(list(string), [])
    }))
    application_insights_connection_string = optional(string, "")
    https_only                             = optional(bool)
    public_network_access_enabled          = optional(bool)
    enabled                                = optional(bool)
    client_certificate_enabled             = optional(bool)
    client_certificate_mode                = optional(string, "")
    client_certificate_exclusion_paths     = optional(string, "")
    virtual_network_subnet_id              = optional(string, "")
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    webdeploy_publish_basic_authentication_enabled = optional(bool)
    zip_deploy_file                                = optional(string, "")
    auth_settings_v2 = optional(object({
      auth_enabled                            = optional(bool)
      runtime_version                         = optional(string)
      config_file_path                        = optional(string, "")
      require_authentication                  = optional(bool)
      unauthenticated_action                  = optional(string, "")
      default_provider                        = optional(string, "")
      excluded_paths                          = optional(list(string), [])
      require_https                           = optional(bool)
      http_route_api_prefix                   = optional(string)
      forward_proxy_convention                = optional(string, "")
      forward_proxy_custom_host_header_name   = optional(string, "")
      forward_proxy_custom_scheme_header_name = optional(string, "")
      login = object({
        logout_endpoint                   = optional(string, "")
        token_store_enabled               = optional(bool)
        token_refresh_extension_time      = optional(number)
        token_store_path                  = optional(string, "")
        token_store_sas_setting_name      = optional(string, "")
        preserve_url_fragments_for_logins = optional(bool)
        allowed_external_redirect_urls    = optional(list(string), [])
        cookie_expiration_convention      = optional(string, "")
        cookie_expiration_time            = optional(string)
        validate_nonce                    = optional(bool)
        nonce_expiration_time             = optional(string)
      })
      apple_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
      }))
      active_directory_v2 = optional(object({
        client_id                            = string
        tenant_auth_endpoint                 = string
        client_secret_setting_name           = optional(string, "")
        client_secret_certificate_thumbprint = optional(string, "")
        login_parameters                     = optional(map(string), {})
        www_authentication_disabled          = optional(bool)
        jwt_allowed_groups                   = optional(list(string), [])
        jwt_allowed_client_applications      = optional(list(string), [])
        allowed_groups                       = optional(list(string), [])
        allowed_identities                   = optional(list(string), [])
        allowed_applications                 = optional(list(string), [])
        allowed_audiences                    = optional(list(string), [])
      }))
      azure_static_web_app_v2 = optional(object({
        client_id = string
      }))
      custom_oidc_v2 = optional(list(object({
        name                          = string
        client_id                     = string
        openid_configuration_endpoint = string
        name_claim_type               = optional(string, "")
        scopes                        = optional(list(string), [])
      })), [])
      facebook_v2 = optional(object({
        app_id                  = string
        app_secret_setting_name = string
        graph_api_version       = optional(string, "")
        login_scopes            = optional(list(string), [])
      }))
      github_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
        login_scopes               = optional(list(string), [])
      }))
      google_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
        allowed_audiences          = optional(list(string), [])
        login_scopes               = optional(list(string), [])
      }))
      microsoft_v2 = optional(object({
        client_id                  = string
        client_secret_setting_name = string
        allowed_audiences          = optional(list(string), [])
        login_scopes               = optional(list(string), [])
      }))
      twitter_v2 = optional(object({
        consumer_key                 = string
        consumer_secret_setting_name = string
      }))
    }))
    tags = optional(map(string), {})
  })
}
