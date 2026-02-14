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
  description = "Azure Container App specification"
  type = object({
    # The Azure Resource Group where the Container App will be created.
    resource_group = string

    # The name of the Container App.
    name = string

    # The Container App Environment resource ID.
    container_app_environment_id = string

    # Revision operating mode: "Single" or "Multiple". Default: "Single".
    revision_mode = optional(string, "Single")

    # Workload profile name. Omit to use the default Consumption profile.
    workload_profile_name = optional(string)

    # Maximum number of inactive revisions to retain (0-100).
    max_inactive_revisions = optional(number)

    # --- Template section (creates revisions when changed) ---

    # Main containers. At least one is required.
    containers = list(object({
      # Container name. Unique within the app.
      name = string
      # Container image in repository:tag format.
      image = string
      # CPU allocation in vCPU cores.
      cpu = number
      # Memory allocation in Gi format (e.g. "0.5Gi", "1Gi").
      memory = string

      # Environment variables.
      env = optional(list(object({
        name        = string
        value       = optional(string)
        secret_name = optional(string)
      })), [])

      # Container command (entrypoint override).
      command = optional(list(string), [])

      # Container arguments.
      args = optional(list(string), [])

      # Liveness probe.
      liveness_probe = optional(object({
        transport                = string
        port                     = number
        path                     = optional(string)
        host                     = optional(string)
        headers = optional(list(object({
          name  = string
          value = string
        })), [])
        initial_delay_in_seconds = optional(number, 0)
        interval_seconds         = optional(number, 10)
        timeout_seconds          = optional(number, 1)
        failure_count_threshold  = optional(number, 3)
        success_count_threshold  = optional(number, 3)
      }))

      # Readiness probe.
      readiness_probe = optional(object({
        transport                = string
        port                     = number
        path                     = optional(string)
        host                     = optional(string)
        headers = optional(list(object({
          name  = string
          value = string
        })), [])
        initial_delay_in_seconds = optional(number, 0)
        interval_seconds         = optional(number, 10)
        timeout_seconds          = optional(number, 1)
        failure_count_threshold  = optional(number, 3)
        success_count_threshold  = optional(number, 3)
      }))

      # Startup probe.
      startup_probe = optional(object({
        transport                = string
        port                     = number
        path                     = optional(string)
        host                     = optional(string)
        headers = optional(list(object({
          name  = string
          value = string
        })), [])
        initial_delay_in_seconds = optional(number, 0)
        interval_seconds         = optional(number, 10)
        timeout_seconds          = optional(number, 1)
        failure_count_threshold  = optional(number, 3)
        success_count_threshold  = optional(number, 3)
      }))

      # Volume mounts.
      volume_mounts = optional(list(object({
        name     = string
        path     = string
        sub_path = optional(string)
      })), [])
    }))

    # Init containers. Run to completion before main containers start.
    init_containers = optional(list(object({
      name   = string
      image  = string
      cpu    = optional(number)
      memory = optional(string)

      env = optional(list(object({
        name        = string
        value       = optional(string)
        secret_name = optional(string)
      })), [])

      command = optional(list(string), [])
      args    = optional(list(string), [])

      volume_mounts = optional(list(object({
        name     = string
        path     = string
        sub_path = optional(string)
      })), [])
    })), [])

    # Volumes available to containers.
    volumes = optional(list(object({
      name          = string
      storage_type  = optional(string, "EmptyDir")
      storage_name  = optional(string)
      mount_options = optional(string)
    })), [])

    # --- Scale configuration ---

    # Minimum number of replicas. Default: 0 (scale-to-zero).
    min_replicas = optional(number, 0)

    # Maximum number of replicas. Default: 10.
    max_replicas = optional(number, 10)

    # Scale cooldown period in seconds. Default: 300.
    cooldown_period_in_seconds = optional(number, 300)

    # KEDA polling interval in seconds. Default: 30.
    polling_interval_in_seconds = optional(number, 30)

    # Revision suffix for named revisions.
    revision_suffix = optional(string)

    # Termination grace period in seconds. Default: 0.
    termination_grace_period_seconds = optional(number, 0)

    # --- Scale rules ---

    # HTTP scale rules.
    http_scale_rules = optional(list(object({
      name                = string
      concurrent_requests = string
      authentication = optional(list(object({
        secret_name       = string
        trigger_parameter = string
      })), [])
    })), [])

    # TCP scale rules.
    tcp_scale_rules = optional(list(object({
      name                = string
      concurrent_requests = string
      authentication = optional(list(object({
        secret_name       = string
        trigger_parameter = string
      })), [])
    })), [])

    # Azure Queue scale rules.
    azure_queue_scale_rules = optional(list(object({
      name         = string
      queue_name   = string
      queue_length = number
      authentication = list(object({
        secret_name       = string
        trigger_parameter = string
      }))
    })), [])

    # Custom KEDA scale rules.
    custom_scale_rules = optional(list(object({
      name             = string
      custom_rule_type = string
      metadata         = optional(map(string), {})
      authentication = optional(list(object({
        secret_name       = string
        trigger_parameter = string
      })), [])
    })), [])

    # --- App-level configuration ---

    # Secrets available to the app.
    secrets = optional(list(object({
      name                = string
      value               = optional(string)
      key_vault_secret_id = optional(string)
      identity            = optional(string)
    })), [])

    # Private container registry credentials.
    registries = optional(list(object({
      server               = string
      username             = optional(string)
      password_secret_name = optional(string)
      identity             = optional(string)
    })), [])

    # Ingress configuration.
    ingress = optional(object({
      external_enabled           = optional(bool, false)
      target_port                = number
      exposed_port               = optional(number)
      transport                  = optional(string, "auto")
      allow_insecure_connections = optional(bool, false)
      client_certificate_mode    = optional(string)

      # Traffic weight distribution across revisions.
      traffic_weight = list(object({
        latest_revision = optional(bool, false)
        revision_suffix = optional(string)
        percentage      = number
        label           = optional(string)
      }))

      # IP security restrictions.
      ip_security_restrictions = optional(list(object({
        name             = string
        action           = string
        ip_address_range = string
        description      = optional(string)
      })), [])

      # CORS policy.
      cors_policy = optional(object({
        allowed_origins           = list(string)
        allowed_headers           = optional(list(string), [])
        allowed_methods           = optional(list(string), [])
        exposed_headers           = optional(list(string), [])
        max_age_in_seconds        = optional(number)
        allow_credentials_enabled = optional(bool, false)
      }))
    }))

    # Dapr sidecar configuration.
    dapr = optional(object({
      app_id       = string
      app_port     = optional(number)
      app_protocol = optional(string, "http")
    }))

    # Managed identity configuration.
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
  })
}
