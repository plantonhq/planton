variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Function App specification"
  type = object({
    # The Azure region where the Function App will be created.
    region = string

    # The Azure Resource Group name.
    resource_group = string

    # The name of the Function App (globally unique).
    name = string

    # The App Service Plan resource ID.
    service_plan_id = string

    # The Azure Storage Account name for Function App runtime state.
    storage_account_name = string

    # Access key for the storage account. Conflicts with storage_uses_managed_identity.
    storage_account_access_key = optional(string)

    # Use managed identity for storage access instead of an access key.
    storage_uses_managed_identity = optional(bool, false)

    # Azure Functions runtime version (e.g., "~4").
    functions_extension_version = optional(string, "~4")

    # Site configuration including runtime, scaling, security, and networking.
    site_config = object({
      # Application stack (runtime selection).
      application_stack = optional(object({
        # .NET runtime version: "3.1", "6.0", "7.0", "8.0", "9.0", "10.0"
        dotnet_version = optional(string)

        # Use the .NET isolated worker runtime model.
        use_dotnet_isolated_runtime = optional(bool, false)

        # Node.js runtime version: "12", "14", "16", "18", "20", "22", "24"
        node_version = optional(string)

        # Python runtime version: "3.8", "3.9", "3.10", "3.11", "3.12", "3.13", "3.14"
        python_version = optional(string)

        # Java runtime version: "8", "11", "17", "21"
        java_version = optional(string)

        # PowerShell Core runtime version: "7", "7.2", "7.4"
        powershell_core_version = optional(string)

        # Docker container configuration.
        docker = optional(object({
          registry_url      = string
          image_name        = string
          image_tag         = string
          registry_username = optional(string)
          registry_password = optional(string)
        }))

        # Use a custom handler runtime.
        use_custom_runtime = optional(bool)
      }))

      # Keep the Function App always loaded in memory.
      always_on = optional(bool)

      # Custom startup command.
      app_command_line = optional(string)

      # Health check endpoint path.
      health_check_path = optional(string)

      # Minimum TLS version for incoming requests.
      minimum_tls_version = optional(string, "1.2")

      # Minimum TLS version for SCM (Kudu) site.
      scm_minimum_tls_version = optional(string, "1.2")

      # Maximum number of workers for scale-out (Consumption/Elastic Premium).
      app_scale_limit = optional(number)

      # Minimum pre-warmed instances (Elastic Premium only).
      elastic_instance_minimum = optional(number)

      # Number of pre-warmed instances beyond the minimum.
      pre_warmed_instance_count = optional(number)

      # Number of worker instances (Dedicated plans).
      worker_count = optional(number)

      # Enable HTTP/2 protocol.
      http2_enabled = optional(bool, false)

      # Enable WebSocket connections.
      websockets_enabled = optional(bool, false)

      # Use a 32-bit worker process.
      use_32_bit_worker = optional(bool, false)

      # Route all outbound traffic through VNet.
      vnet_route_all_enabled = optional(bool, false)

      # FTPS state: "AllAllowed", "FtpsOnly", "Disabled".
      ftps_state = optional(string, "Disabled")

      # Load balancing mode.
      load_balancing_mode = optional(string, "LeastRequests")

      # Enable runtime scale monitoring for KEDA-based triggers.
      runtime_scale_monitoring_enabled = optional(bool)

      # CORS configuration.
      cors = optional(object({
        allowed_origins     = list(string)
        support_credentials = optional(bool, false)
      }))

      # IP restriction rules for the main site.
      ip_restrictions = optional(list(object({
        name                      = optional(string)
        priority                  = optional(number)
        action                    = optional(string, "Allow")
        ip_address                = optional(string)
        service_tag               = optional(string)
        virtual_network_subnet_id = optional(string)
        description               = optional(string)
        headers = optional(object({
          x_forwarded_for   = optional(list(string), [])
          x_forwarded_host  = optional(list(string), [])
          x_azure_fdid      = optional(list(string), [])
          x_fd_health_probe = optional(list(string), [])
        }))
      })), [])

      # Default action for main site IP restrictions: "Allow" or "Deny".
      ip_restriction_default_action = optional(string, "Allow")

      # Use main site IP restrictions for SCM site.
      scm_use_main_ip_restriction = optional(bool, false)

      # IP restriction rules for the SCM (Kudu) site.
      scm_ip_restrictions = optional(list(object({
        name                      = optional(string)
        priority                  = optional(number)
        action                    = optional(string, "Allow")
        ip_address                = optional(string)
        service_tag               = optional(string)
        virtual_network_subnet_id = optional(string)
        description               = optional(string)
        headers = optional(object({
          x_forwarded_for   = optional(list(string), [])
          x_forwarded_host  = optional(list(string), [])
          x_azure_fdid      = optional(list(string), [])
          x_fd_health_probe = optional(list(string), [])
        }))
      })), [])

      # Default action for SCM site IP restrictions: "Allow" or "Deny".
      scm_ip_restriction_default_action = optional(string, "Allow")

      # App Service logging configuration.
      app_service_logs = optional(object({
        disk_quota_mb         = optional(number, 35)
        retention_period_days = optional(number)
      }))

      # Default documents list for the web server.
      default_documents = optional(list(string))

      # Use managed identity for ACR image pulls.
      container_registry_use_managed_identity = optional(bool, false)

      # Client ID of the managed identity for ACR pulls.
      container_registry_managed_identity_client_id = optional(string)

      # Application Insights instrumentation key (classic).
      application_insights_key = optional(string)
    })

    # Application settings (environment variables) as key-value pairs.
    app_settings = optional(map(string), {})

    # Named connection strings.
    connection_strings = optional(list(object({
      name  = string
      type  = string
      value = string
    })), [])

    # Application Insights connection string for APM telemetry.
    application_insights_connection_string = optional(string)

    # Enforce HTTPS-only access.
    https_only = optional(bool, true)

    # Enable public network access.
    public_network_access_enabled = optional(bool, true)

    # Enable built-in logging via AzureWebJobsDashboard.
    builtin_logging_enabled = optional(bool, true)

    # Subnet ID for VNet integration.
    virtual_network_subnet_id = optional(string)

    # Managed identity configuration.
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))

    # User Assigned Identity ID for Key Vault references.
    key_vault_reference_identity_id = optional(string)

    # Enable client certificate authentication (mTLS).
    client_certificate_enabled = optional(bool, false)

    # Client certificate mode: "Required", "Optional", "OptionalInteractiveUser".
    client_certificate_mode = optional(string, "Optional")

    # Paths excluded from client certificate validation (semicolon-separated).
    client_certificate_exclusion_paths = optional(string)

    # Force disable the Azure Files content share.
    content_share_force_disabled = optional(bool, false)

    # Azure Storage Account mounts.
    storage_mounts = optional(list(object({
      name         = string
      type         = string
      account_name = string
      share_name   = string
      access_key   = string
      mount_path   = optional(string)
    })), [])
  })
}
