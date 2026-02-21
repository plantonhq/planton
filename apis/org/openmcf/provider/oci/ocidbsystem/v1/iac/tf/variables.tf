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
  description = "OciDbSystem specification"
  type = object({
    compartment_id = object({
      value = string
    })

    display_name        = optional(string, "")
    availability_domain = string
    shape               = string

    subnet_id = object({
      value = string
    })

    ssh_public_keys = list(string)
    hostname        = string

    cpu_core_count         = optional(number, 0)
    database_edition       = optional(string, "")
    license_model          = optional(string, "")
    data_storage_size_in_gb = optional(number, 0)
    data_storage_percentage = optional(number, 0)
    disk_redundancy        = optional(string, "")
    node_count             = optional(number, 0)
    domain                 = optional(string, "")
    cluster_name           = optional(string, "")
    fault_domains          = optional(list(string), [])

    nsg_ids = optional(list(object({
      value = string
    })), [])

    backup_subnet_id = optional(object({
      value = string
    }), null)

    backup_network_nsg_ids = optional(list(object({
      value = string
    })), [])

    kms_key_id = optional(object({
      value = string
    }), null)
    kms_key_version_id = optional(string, "")

    time_zone        = optional(string, "")
    sparse_diskgroup = optional(bool, null)
    storage_volume_performance_mode = optional(string, "")
    private_ip = optional(string, "")

    data_collection_options = optional(object({
      is_diagnostics_events_enabled = optional(bool, null)
      is_health_monitoring_enabled  = optional(bool, null)
      is_incident_logs_enabled      = optional(bool, null)
    }), null)

    db_system_options = optional(object({
      storage_management = optional(string, "")
    }), null)

    maintenance_window_details = optional(object({
      preference                       = optional(string, "")
      patching_mode                    = optional(string, "")
      lead_time_in_weeks               = optional(number, 0)
      months                           = optional(list(string), [])
      weeks_of_month                   = optional(list(number), [])
      days_of_week                     = optional(list(string), [])
      hours_of_day                     = optional(list(number), [])
      custom_action_timeout_in_mins    = optional(number, 0)
      is_custom_action_timeout_enabled = optional(bool, null)
      is_monthly_patching_enabled      = optional(bool, null)
    }), null)

    db_home = object({
      db_version   = optional(string, "")
      display_name = optional(string, "")
      database_software_image_id = optional(object({
        value = string
      }), null)

      database = object({
        admin_password   = string
        db_name          = string
        character_set    = optional(string, "")
        ncharacter_set   = optional(string, "")
        pdb_name         = optional(string, "")
        db_domain        = optional(string, "")
        kms_key_id = optional(object({
          value = string
        }), null)
        kms_key_version_id = optional(string, "")
        vault_id = optional(object({
          value = string
        }), null)

        db_backup_config = optional(object({
          auto_backup_enabled    = optional(bool, null)
          auto_backup_window     = optional(string, "")
          recovery_window_in_days = optional(number, 0)
        }), null)
      })
    })
  })
}
