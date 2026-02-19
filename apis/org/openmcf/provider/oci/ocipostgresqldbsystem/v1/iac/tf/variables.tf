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
  description = "OciPostgresqlDbSystem specification"
  type = object({
    compartment_id = object({
      value = string
    })

    display_name = optional(string, "")
    db_version   = string
    shape        = string

    instance_ocpu_count         = optional(number, 0)
    instance_memory_size_in_gbs = optional(number, 0)
    instance_count              = optional(number, 0)

    network_details = object({
      subnet_id = object({
        value = string
      })

      nsg_ids = optional(list(object({
        value = string
      })), [])

      is_reader_endpoint_enabled     = optional(bool, null)
      primary_db_endpoint_private_ip = optional(string, "")
    })

    storage_details = object({
      is_regionally_durable = bool
      availability_domain   = optional(string, "")
      iops                  = optional(number, 0)
    })

    credentials = optional(object({
      username = string

      password_details = object({
        password_type  = optional(string, "")
        password       = optional(string, "")
        secret_id = optional(object({
          value = string
        }), null)
        secret_version = optional(string, "")
      })
    }), null)

    management_policy = optional(object({
      backup_policy = optional(object({
        kind              = optional(string, "")
        backup_start      = optional(string, "")
        retention_days    = optional(number, 0)
        days_of_the_month = optional(list(number), [])
        days_of_the_week  = optional(list(string), [])
      }), null)

      maintenance_window_start = optional(string, "")
    }), null)

    config_id = optional(object({
      value = string
    }), null)

    description = optional(string, "")

    instances_details = optional(list(object({
      display_name = optional(string, "")
      description  = optional(string, "")
      private_ip   = optional(string, "")
    })), [])
  })
}
