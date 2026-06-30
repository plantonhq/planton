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
  description = "Azure Linux Web App specification"
  type = object({
    # The Azure region where the Web App will be created.
    region = string

    # The Azure Resource Group name.
    resource_group = string

    # The name of the Web App (globally unique).
    name = string

    # The App Service Plan resource ID.
    service_plan_id = string

    # Whether the Web App is enabled. Setting to false stops the app without deleting it.
    enabled = optional(bool, true)

    # Enable client affinity (ARR session affinity cookies).
    client_affinity_enabled = optional(bool, false)

    # Site configuration including runtime, scaling, security, and networking.
    site_config = object({
      # Application stack (runtime selection).
      application_stack = optional(object({
        # .NET runtime version: "3.1", "6.0", "7.0", "8.0", "9.0", "10.0"
        dotnet_version = optional(string)

        # Node.js runtime version: "12-lts", "14-lts", "16-lts", "18-lts", "20-lts", "22-lts"
        node_version = optional(string)

        # Python runtime version: "3.8", "3.9", "3.10", "3.11", "3.12", "3.13", "3.14"
        python_version = optional(string)

        # Java runtime version: "8", "11", "17", "21"
        java_version = optional(string)

        # Java application server type: "JAVA", "TOMCAT", "JBOSSEAP"
        java_server = optional(string)

        # Java application server version (e.g., "9.0", "10.0" for Tomcat).
        java_server_version = optional(string)

        # PHP runtime version: "8.0", "8.1", "8.2", "8.3"
        php_version = optional(string)

        # Ruby runtime version: "2.7", "3.0", "3.1", "3.2"
        ruby_version = optional(string)

        # Go runtime version: "1.19", "1.20", "1.21"
        go_version = optional(string)

        # Docker image name (e.g., "myregistry.azurecr.io/myapp:latest").
        docker_image_name = optional(string)

        # Docker registry URL (e.g., "https://myregistry.azurecr.io").
        docker_registry_url = optional(string)

        # Docker registry username for private registries.
        docker_registry_username = optional(string)

        # Docker registry password for private registries.
        docker_registry_password = optional(string)
      }))

      # Keep the Web App always loaded in memory.
      always_on = optional(bool)

      # Custom startup command.
      app_command_line = optional(string)

      # Health check endpoint path.
      health_check_path = optional(string)

      # Time (in minutes) after which an unhealthy instance is evicted. 2-10.
      health_check_eviction_time_in_min = optional(number)

      # Minimum TLS version for incoming requests.
      minimum_tls_version = optional(string, "1.2")

      # Minimum TLS version for SCM (Kudu) site.
      scm_minimum_tls_version = optional(string, "1.2")

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

      # Use managed identity for ACR image pulls.
      container_registry_use_managed_identity = optional(bool, false)

      # Client ID of the managed identity for ACR pulls.
      container_registry_managed_identity_client_id = optional(string)
    })

    # Application settings (environment variables) as key-value pairs.
    app_settings = optional(map(string), {})

    # Named connection strings.
    connection_strings = optional(list(object({
      name  = string
      type  = string
      value = string
    })), [])

    # Application Insights connection string for APM telemetry (merged into app_settings).
    application_insights_connection_string = optional(string)

    # Enforce HTTPS-only access.
    https_only = optional(bool, true)

    # Enable public network access.
    public_network_access_enabled = optional(bool, true)

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

    # Azure Storage Account mounts.
    storage_mounts = optional(list(object({
      name         = string
      type         = string
      account_name = string
      share_name   = string
      access_key   = string
      mount_path   = optional(string)
    })), [])

    # Logging configuration for the Web App.
    logs = optional(object({
      # Application-level log settings.
      application_logs = optional(object({
        # Log level for file system logging: "Off", "Error", "Warning", "Information", "Verbose".
        file_system_level = string
      }))

      # HTTP request/response log settings.
      http_logs = optional(object({
        # Maximum size of log files in MB before rotation.
        retention_in_mb = optional(number, 35)
        # Number of days to retain log files.
        retention_in_days = optional(number, 0)
      }))

      # Enable failed request tracing.
      failed_request_tracing = optional(bool, false)

      # Enable detailed error messages in responses.
      detailed_error_messages = optional(bool, false)
    }))
  })
}
