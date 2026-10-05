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
  description = "AzureContainerInstance specification"
  type = object({
    resource_group              = string
    name                        = string
    region                      = string
    os_type                     = string
    restart_policy              = optional(string, "")
    sku                         = optional(string, "")
    priority                    = optional(string, "")
    ip_address_type             = optional(string, "")
    dns_name_label              = optional(string, "")
    dns_name_label_reuse_policy = optional(string, "")
    subnet_id                   = optional(string, "")
    zones                       = optional(list(string), [])
    exposed_ports = optional(list(object({
      port     = optional(number, 0)
      protocol = optional(string, "")
    })), [])
    containers = list(object({
      name         = string
      image        = string
      cpu          = number
      memory       = number
      cpu_limit    = optional(number)
      memory_limit = optional(number)
      ports = optional(list(object({
        port     = optional(number, 0)
        protocol = optional(string, "")
      })), [])
      environment_variables        = optional(map(string), {})
      secure_environment_variables = optional(map(string), {})
      commands                     = optional(list(string), [])
      volumes = optional(list(object({
        name       = string
        mount_path = string
        read_only  = optional(bool, false)
        azure_file = optional(object({
          share_name           = string
          storage_account_name = string
          storage_account_key  = string
        }))
        empty_dir = optional(bool, false)
        git_repo = optional(object({
          url       = string
          directory = optional(string, "")
          revision  = optional(string, "")
        }))
        secret = optional(map(string), {})
      })), [])
      security = optional(object({
        privilege_enabled = optional(bool, false)
      }))
      liveness_probe = optional(object({
        exec = optional(list(string), [])
        http_get = optional(object({
          path         = optional(string, "")
          port         = optional(number, 0)
          scheme       = optional(string, "")
          http_headers = optional(map(string), {})
        }))
        initial_delay_seconds = optional(number, 0)
        period_seconds        = optional(number, 0)
        failure_threshold     = optional(number, 0)
        success_threshold     = optional(number, 0)
        timeout_seconds       = optional(number, 0)
      }))
      readiness_probe = optional(object({
        exec = optional(list(string), [])
        http_get = optional(object({
          path         = optional(string, "")
          port         = optional(number, 0)
          scheme       = optional(string, "")
          http_headers = optional(map(string), {})
        }))
        initial_delay_seconds = optional(number, 0)
        period_seconds        = optional(number, 0)
        failure_threshold     = optional(number, 0)
        success_threshold     = optional(number, 0)
        timeout_seconds       = optional(number, 0)
      }))
    }))
    init_containers = optional(list(object({
      name                         = string
      image                        = string
      environment_variables        = optional(map(string), {})
      secure_environment_variables = optional(map(string), {})
      commands                     = optional(list(string), [])
      volumes = optional(list(object({
        name       = string
        mount_path = string
        read_only  = optional(bool, false)
        azure_file = optional(object({
          share_name           = string
          storage_account_name = string
          storage_account_key  = string
        }))
        empty_dir = optional(bool, false)
        git_repo = optional(object({
          url       = string
          directory = optional(string, "")
          revision  = optional(string, "")
        }))
        secret = optional(map(string), {})
      })), [])
      security = optional(object({
        privilege_enabled = optional(bool, false)
      }))
    })), [])
    image_registry_credentials = optional(list(object({
      server                    = string
      username                  = optional(string, "")
      password                  = optional(string, "")
      user_assigned_identity_id = optional(string, "")
    })), [])
    identity = optional(object({
      type         = string
      identity_ids = optional(list(string), [])
    }))
    diagnostics_log_analytics = optional(object({
      workspace_id  = string
      workspace_key = string
      log_type      = optional(string, "")
      metadata      = optional(map(string), {})
    }))
    dns_config = optional(object({
      nameservers    = list(string)
      search_domains = optional(list(string), [])
      options        = optional(list(string), [])
    }))
    key_vault_key_id                    = optional(string, "")
    key_vault_user_assigned_identity_id = optional(string, "")
    tags                                = optional(map(string), {})
  })
}
