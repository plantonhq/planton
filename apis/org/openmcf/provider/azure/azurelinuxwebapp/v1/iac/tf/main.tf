# Create the Azure Linux Web App.
resource "azurerm_linux_web_app" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  service_plan_id     = var.spec.service_plan_id

  enabled                 = var.spec.enabled
  client_affinity_enabled = var.spec.client_affinity_enabled

  https_only                    = var.spec.https_only
  public_network_access_enabled = var.spec.public_network_access_enabled

  virtual_network_subnet_id          = var.spec.virtual_network_subnet_id
  key_vault_reference_identity_id    = var.spec.key_vault_reference_identity_id
  client_certificate_enabled         = var.spec.client_certificate_enabled
  client_certificate_mode            = var.spec.client_certificate_mode
  client_certificate_exclusion_paths = var.spec.client_certificate_exclusion_paths

  app_settings = local.merged_app_settings

  tags = local.final_tags

  # ---------------------------------------------------------------------------
  # Site Config
  # ---------------------------------------------------------------------------
  site_config {
    always_on                                     = var.spec.site_config.always_on
    app_command_line                              = var.spec.site_config.app_command_line
    health_check_path                             = var.spec.site_config.health_check_path
    health_check_eviction_time_in_min             = var.spec.site_config.health_check_eviction_time_in_min
    minimum_tls_version                           = var.spec.site_config.minimum_tls_version
    scm_minimum_tls_version                       = var.spec.site_config.scm_minimum_tls_version
    worker_count                                  = var.spec.site_config.worker_count
    http2_enabled                                 = var.spec.site_config.http2_enabled
    websockets_enabled                            = var.spec.site_config.websockets_enabled
    use_32_bit_worker                             = var.spec.site_config.use_32_bit_worker
    vnet_route_all_enabled                        = var.spec.site_config.vnet_route_all_enabled
    ftps_state                                    = var.spec.site_config.ftps_state
    load_balancing_mode                           = var.spec.site_config.load_balancing_mode
    ip_restriction_default_action                 = var.spec.site_config.ip_restriction_default_action
    scm_use_main_ip_restriction                   = var.spec.site_config.scm_use_main_ip_restriction
    scm_ip_restriction_default_action             = var.spec.site_config.scm_ip_restriction_default_action
    container_registry_use_managed_identity       = var.spec.site_config.container_registry_use_managed_identity
    container_registry_managed_identity_client_id = var.spec.site_config.container_registry_managed_identity_client_id

    # --- Application Stack ---
    dynamic "application_stack" {
      for_each = var.spec.site_config.application_stack != null ? [var.spec.site_config.application_stack] : []
      content {
        dotnet_version              = application_stack.value.dotnet_version
        node_version                = application_stack.value.node_version
        python_version              = application_stack.value.python_version
        java_version                = application_stack.value.java_version
        java_server                 = application_stack.value.java_server
        java_server_version         = application_stack.value.java_server_version
        php_version                 = application_stack.value.php_version
        ruby_version                = application_stack.value.ruby_version
        go_version                  = application_stack.value.go_version
        docker_image_name           = application_stack.value.docker_image_name
        docker_registry_url         = application_stack.value.docker_registry_url
        docker_registry_username    = application_stack.value.docker_registry_username
        docker_registry_password    = application_stack.value.docker_registry_password
      }
    }

    # --- CORS ---
    dynamic "cors" {
      for_each = var.spec.site_config.cors != null ? [var.spec.site_config.cors] : []
      content {
        allowed_origins     = cors.value.allowed_origins
        support_credentials = cors.value.support_credentials
      }
    }

    # --- IP Restrictions (main site) ---
    dynamic "ip_restriction" {
      for_each = var.spec.site_config.ip_restrictions
      content {
        name                      = ip_restriction.value.name
        priority                  = ip_restriction.value.priority
        action                    = ip_restriction.value.action
        ip_address                = ip_restriction.value.ip_address
        service_tag               = ip_restriction.value.service_tag
        virtual_network_subnet_id = ip_restriction.value.virtual_network_subnet_id
        description               = ip_restriction.value.description

        dynamic "headers" {
          for_each = ip_restriction.value.headers != null ? [ip_restriction.value.headers] : []
          content {
            x_forwarded_for   = headers.value.x_forwarded_for
            x_forwarded_host  = headers.value.x_forwarded_host
            x_azure_fdid      = headers.value.x_azure_fdid
            x_fd_health_probe = headers.value.x_fd_health_probe
          }
        }
      }
    }

    # --- SCM IP Restrictions ---
    dynamic "scm_ip_restriction" {
      for_each = var.spec.site_config.scm_ip_restrictions
      content {
        name                      = scm_ip_restriction.value.name
        priority                  = scm_ip_restriction.value.priority
        action                    = scm_ip_restriction.value.action
        ip_address                = scm_ip_restriction.value.ip_address
        service_tag               = scm_ip_restriction.value.service_tag
        virtual_network_subnet_id = scm_ip_restriction.value.virtual_network_subnet_id
        description               = scm_ip_restriction.value.description

        dynamic "headers" {
          for_each = scm_ip_restriction.value.headers != null ? [scm_ip_restriction.value.headers] : []
          content {
            x_forwarded_for   = headers.value.x_forwarded_for
            x_forwarded_host  = headers.value.x_forwarded_host
            x_azure_fdid      = headers.value.x_azure_fdid
            x_fd_health_probe = headers.value.x_fd_health_probe
          }
        }
      }
    }

  }

  # ---------------------------------------------------------------------------
  # Logs (top-level block for Web App)
  # ---------------------------------------------------------------------------
  dynamic "logs" {
    for_each = var.spec.logs != null ? [var.spec.logs] : []
    content {
      dynamic "application_logs" {
        for_each = logs.value.application_logs != null ? [logs.value.application_logs] : []
        content {
          file_system_level = application_logs.value.file_system_level
        }
      }
      dynamic "http_logs" {
        for_each = logs.value.http_logs != null ? [logs.value.http_logs] : []
        content {
          file_system {
            retention_in_mb   = http_logs.value.retention_in_mb
            retention_in_days = http_logs.value.retention_in_days
          }
        }
      }
      failed_request_tracing  = logs.value.failed_request_tracing
      detailed_error_messages = logs.value.detailed_error_messages
    }
  }

  # ---------------------------------------------------------------------------
  # Connection Strings
  # ---------------------------------------------------------------------------
  dynamic "connection_string" {
    for_each = var.spec.connection_strings
    content {
      name  = connection_string.value.name
      type  = connection_string.value.type
      value = connection_string.value.value
    }
  }

  # ---------------------------------------------------------------------------
  # Storage Account Mounts
  # ---------------------------------------------------------------------------
  dynamic "storage_account" {
    for_each = var.spec.storage_mounts
    content {
      name         = storage_account.value.name
      type         = storage_account.value.type
      account_name = storage_account.value.account_name
      share_name   = storage_account.value.share_name
      access_key   = storage_account.value.access_key
      mount_path   = storage_account.value.mount_path
    }
  }

  # ---------------------------------------------------------------------------
  # Identity
  # ---------------------------------------------------------------------------
  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = identity.value.type
      identity_ids = length(identity.value.identity_ids) > 0 ? identity.value.identity_ids : null
    }
  }
}
