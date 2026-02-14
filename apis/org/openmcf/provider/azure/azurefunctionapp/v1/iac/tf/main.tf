# Create the Azure Linux Function App.
resource "azurerm_linux_function_app" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group
  service_plan_id     = var.spec.service_plan_id

  storage_account_name          = var.spec.storage_account_name
  storage_account_access_key    = var.spec.storage_account_access_key
  storage_uses_managed_identity = var.spec.storage_uses_managed_identity

  functions_extension_version = var.spec.functions_extension_version

  https_only                    = var.spec.https_only
  public_network_access_enabled = var.spec.public_network_access_enabled
  builtin_logging_enabled       = var.spec.builtin_logging_enabled

  virtual_network_subnet_id          = var.spec.virtual_network_subnet_id
  key_vault_reference_identity_id    = var.spec.key_vault_reference_identity_id
  client_certificate_enabled         = var.spec.client_certificate_enabled
  client_certificate_mode            = var.spec.client_certificate_mode
  client_certificate_exclusion_paths = var.spec.client_certificate_exclusion_paths
  content_share_force_disabled       = var.spec.content_share_force_disabled

  app_settings = var.spec.app_settings

  tags = local.final_tags

  # ---------------------------------------------------------------------------
  # Site Config
  # ---------------------------------------------------------------------------
  site_config {
    always_on                                     = var.spec.site_config.always_on
    app_command_line                              = var.spec.site_config.app_command_line
    health_check_path                             = var.spec.site_config.health_check_path
    minimum_tls_version                           = var.spec.site_config.minimum_tls_version
    scm_minimum_tls_version                       = var.spec.site_config.scm_minimum_tls_version
    app_scale_limit                               = var.spec.site_config.app_scale_limit
    elastic_instance_minimum                      = var.spec.site_config.elastic_instance_minimum
    pre_warmed_instance_count                     = var.spec.site_config.pre_warmed_instance_count
    worker_count                                  = var.spec.site_config.worker_count
    http2_enabled                                 = var.spec.site_config.http2_enabled
    websockets_enabled                            = var.spec.site_config.websockets_enabled
    use_32_bit_worker                             = var.spec.site_config.use_32_bit_worker
    vnet_route_all_enabled                        = var.spec.site_config.vnet_route_all_enabled
    ftps_state                                    = var.spec.site_config.ftps_state
    load_balancing_mode                           = var.spec.site_config.load_balancing_mode
    runtime_scale_monitoring_enabled              = var.spec.site_config.runtime_scale_monitoring_enabled
    ip_restriction_default_action                 = var.spec.site_config.ip_restriction_default_action
    scm_use_main_ip_restriction                   = var.spec.site_config.scm_use_main_ip_restriction
    scm_ip_restriction_default_action             = var.spec.site_config.scm_ip_restriction_default_action
    default_documents                             = var.spec.site_config.default_documents
    container_registry_use_managed_identity       = var.spec.site_config.container_registry_use_managed_identity
    container_registry_managed_identity_client_id = var.spec.site_config.container_registry_managed_identity_client_id
    application_insights_key                      = var.spec.site_config.application_insights_key
    application_insights_connection_string        = var.spec.application_insights_connection_string

    # --- Application Stack ---
    dynamic "application_stack" {
      for_each = var.spec.site_config.application_stack != null ? [var.spec.site_config.application_stack] : []
      content {
        dotnet_version              = application_stack.value.dotnet_version
        use_dotnet_isolated_runtime = application_stack.value.use_dotnet_isolated_runtime
        node_version                = application_stack.value.node_version
        python_version              = application_stack.value.python_version
        java_version                = application_stack.value.java_version
        powershell_core_version     = application_stack.value.powershell_core_version
        use_custom_runtime          = application_stack.value.use_custom_runtime

        dynamic "docker" {
          for_each = application_stack.value.docker != null ? [application_stack.value.docker] : []
          content {
            registry_url      = docker.value.registry_url
            image_name        = docker.value.image_name
            image_tag         = docker.value.image_tag
            registry_username = docker.value.registry_username
            registry_password = docker.value.registry_password
          }
        }
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

    # --- App Service Logs ---
    dynamic "app_service_logs" {
      for_each = var.spec.site_config.app_service_logs != null ? [var.spec.site_config.app_service_logs] : []
      content {
        disk_quota_mb         = app_service_logs.value.disk_quota_mb
        retention_period_days = app_service_logs.value.retention_period_days
      }
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
