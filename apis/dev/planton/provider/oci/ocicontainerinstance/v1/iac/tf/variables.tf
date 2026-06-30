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
  description = "OciContainerInstance specification"
  type = object({
    compartment_id = object({
      value = string
    })

    availability_domain = string
    display_name        = optional(string, "")
    shape               = string

    shape_config = object({
      ocpus         = number
      memory_in_gbs = optional(number, 0)
    })

    containers = list(object({
      image_url                      = string
      display_name                   = optional(string, "")
      command                        = optional(list(string), [])
      arguments                      = optional(list(string), [])
      environment_variables          = optional(map(string), {})
      working_directory              = optional(string, "")
      is_resource_principal_disabled = optional(bool, false)

      resource_config = optional(object({
        memory_limit_in_gbs = optional(number, 0)
        vcpus_limit         = optional(number, 0)
      }), null)

      health_checks = optional(list(object({
        health_check_type        = string
        port                     = number
        name                     = optional(string, "")
        path                     = optional(string, "")
        failure_action           = optional(string, "")
        failure_threshold        = optional(number, 0)
        success_threshold        = optional(number, 0)
        initial_delay_in_seconds = optional(number, 0)
        interval_in_seconds      = optional(number, 0)
        timeout_in_seconds       = optional(number, 0)
        headers = optional(list(object({
          name  = string
          value = string
        })), [])
      })), [])

      security_context = optional(object({
        is_non_root_user_check_enabled = optional(bool, false)
        is_root_file_system_readonly   = optional(bool, false)
        run_as_user                    = optional(number, 0)
        run_as_group                   = optional(number, 0)
        capabilities = optional(object({
          add_capabilities  = optional(list(string), [])
          drop_capabilities = optional(list(string), [])
        }), null)
      }), null)

      volume_mounts = optional(list(object({
        mount_path  = string
        volume_name = string
        is_read_only = optional(bool, false)
        partition    = optional(number, 0)
        sub_path     = optional(string, "")
      })), [])
    }))

    vnics = list(object({
      subnet_id = object({
        value = string
      })
      display_name           = optional(string, "")
      hostname_label         = optional(string, "")
      is_public_ip_assigned  = optional(bool, null)
      nsg_ids = optional(list(object({
        value = string
      })), [])
      private_ip             = optional(string, "")
      skip_source_dest_check = optional(bool, false)
    }))

    container_restart_policy             = optional(string, "")
    fault_domain                         = optional(string, "")
    graceful_shutdown_timeout_in_seconds = optional(number, 0)

    dns_config = optional(object({
      nameservers = optional(list(string), [])
      options     = optional(list(string), [])
      searches    = optional(list(string), [])
    }), null)

    image_pull_secrets = optional(list(object({
      registry_endpoint = string
      secret_type       = string
      username          = optional(string, "")
      password          = optional(string, "")
      secret_id = optional(object({
        value = string
      }), null)
    })), [])

    volumes = optional(list(object({
      name          = string
      volume_type   = string
      backing_store = optional(string, "")
      configs = optional(list(object({
        data      = optional(string, "")
        file_name = optional(string, "")
        path      = optional(string, "")
      })), [])
    })), [])
  })
}
