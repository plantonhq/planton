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
  description = "AzureDataProtectionBackupPolicy specification"
  type = object({
    vault_id = string
    name     = string
    blob_storage = optional(object({
      operational_default_retention_duration = optional(string, "")
      vault_default_retention_duration       = optional(string, "")
      backup_repeating_time_intervals        = optional(list(string), [])
      time_zone                              = optional(string, "")
      retention_rules = optional(list(object({
        name = string
        criteria = object({
          absolute_criteria      = optional(string, "")
          days_of_month          = optional(list(number), [])
          days_of_week           = optional(list(string), [])
          months_of_year         = optional(list(string), [])
          scheduled_backup_times = optional(list(string), [])
          weeks_of_month         = optional(list(string), [])
        })
        life_cycle = object({
          data_store_type = string
          duration        = string
        })
        priority = number
      })), [])
    }))
    disk = optional(object({
      backup_repeating_time_intervals = list(string)
      default_retention_duration      = string
      retention_rules = optional(list(object({
        name     = string
        duration = string
        criteria = object({
          absolute_criteria = optional(string, "")
        })
        priority = number
      })), [])
      time_zone = optional(string, "")
    }))
    kubernetes_cluster = optional(object({
      backup_repeating_time_intervals = list(string)
      default_retention_rule = object({
        life_cycles = list(object({
          data_store_type = string
          duration        = string
        }))
      })
      retention_rules = optional(list(object({
        name = string
        criteria = object({
          absolute_criteria      = optional(string, "")
          days_of_week           = optional(list(string), [])
          months_of_year         = optional(list(string), [])
          scheduled_backup_times = optional(list(string), [])
          weeks_of_month         = optional(list(string), [])
        })
        life_cycles = list(object({
          data_store_type = string
          duration        = string
        }))
        priority = number
      })), [])
      time_zone = optional(string, "")
    }))
    mysql_flexible_server = optional(object({
      backup_repeating_time_intervals = list(string)
      default_retention_rule = object({
        life_cycles = list(object({
          data_store_type = string
          duration        = string
        }))
      })
      retention_rules = optional(list(object({
        name = string
        criteria = object({
          absolute_criteria      = optional(string, "")
          days_of_week           = optional(list(string), [])
          months_of_year         = optional(list(string), [])
          scheduled_backup_times = optional(list(string), [])
          weeks_of_month         = optional(list(string), [])
        })
        life_cycles = list(object({
          data_store_type = string
          duration        = string
        }))
        priority = number
      })), [])
      time_zone = optional(string, "")
    }))
    postgresql_flexible_server = optional(object({
      backup_repeating_time_intervals = list(string)
      default_retention_rule = object({
        life_cycles = list(object({
          data_store_type = string
          duration        = string
        }))
      })
      retention_rules = optional(list(object({
        name = string
        criteria = object({
          absolute_criteria      = optional(string, "")
          days_of_week           = optional(list(string), [])
          months_of_year         = optional(list(string), [])
          scheduled_backup_times = optional(list(string), [])
          weeks_of_month         = optional(list(string), [])
        })
        life_cycles = list(object({
          data_store_type = string
          duration        = string
        }))
        priority = number
      })), [])
      time_zone = optional(string, "")
    }))
    data_lake_storage = optional(object({
      backup_schedule            = list(string)
      default_retention_duration = string
      retention_rules = optional(list(object({
        name                   = string
        duration               = string
        absolute_criteria      = optional(string, "")
        days_of_week           = optional(list(string), [])
        months_of_year         = optional(list(string), [])
        scheduled_backup_times = optional(list(string), [])
        weeks_of_month         = optional(list(string), [])
      })), [])
      time_zone = optional(string, "")
    }))
  })
}
