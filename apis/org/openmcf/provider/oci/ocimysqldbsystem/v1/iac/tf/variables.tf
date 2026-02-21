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
  description = "OciMysqlDbSystem specification"
  type = object({
    compartment_id = object({
      value = string
    })

    display_name        = optional(string, "")
    availability_domain = string
    shape_name          = string

    subnet_id = object({
      value = string
    })

    admin_username      = optional(string, "")
    admin_password      = optional(string, "")
    mysql_version       = optional(string, "")

    configuration_id = optional(object({
      value = string
    }), null)

    is_highly_available = optional(bool, null)
    hostname_label      = optional(string, "")
    ip_address          = optional(string, "")
    fault_domain        = optional(string, "")
    port                = optional(number, 0)
    port_x              = optional(number, 0)
    description         = optional(string, "")
    crash_recovery      = optional(string, "")
    database_management = optional(string, "")

    nsg_ids = optional(list(object({
      value = string
    })), [])

    data_storage = optional(object({
      data_storage_size_in_gb        = optional(number, 0)
      is_auto_expand_storage_enabled = optional(bool, null)
      max_storage_size_in_gbs        = optional(number, 0)
    }), null)

    backup_policy = optional(object({
      is_enabled       = optional(bool, null)
      retention_in_days = optional(number, 0)
      window_start_time = optional(string, "")
      pitr_policy = optional(object({
        is_enabled = optional(bool, null)
      }), null)
    }), null)

    maintenance = optional(object({
      window_start_time        = string
      maintenance_schedule_type = optional(string, "")
      version_preference        = optional(string, "")
      version_track_preference  = optional(string, "")
    }), null)

    deletion_policy = optional(object({
      automatic_backup_retention = optional(string, "")
      final_backup               = optional(string, "")
      is_delete_protected        = optional(bool, null)
    }), null)

    encrypt_data = optional(object({
      key_generation_type = optional(string, "")
      key_id = optional(object({
        value = string
      }), null)
    }), null)

    secure_connections = optional(object({
      certificate_generation_type = optional(string, "")
      certificate_id = optional(object({
        value = string
      }), null)
    }), null)

    customer_contacts = optional(list(object({
      email = string
    })), [])

    read_endpoint = optional(object({
      is_enabled                  = optional(bool, null)
      exclude_ips                 = optional(list(string), [])
      read_endpoint_hostname_label = optional(string, "")
      read_endpoint_ip_address     = optional(string, "")
    }), null)

    database_console = optional(object({
      status = optional(string, "")
      port   = optional(number, 0)
    }), null)

    rest = optional(object({
      configuration = optional(string, "")
      port          = optional(number, 0)
    }), null)
  })
}
