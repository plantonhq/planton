variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsBackupPlan specification"
  type = object({
    region = string
    rules = list(object({
      name = string
      target_vault_name = string
      schedule = optional(string, "")
      schedule_expression_timezone = optional(string, "")
      start_window_minutes = optional(number, 0)
      completion_window_minutes = optional(number, 0)
      enable_continuous_backup = optional(bool, false)
      recovery_point_tags = optional(map(string), {})
      target_logically_air_gapped_backup_vault_arn = optional(string, "")
      lifecycle = optional(object({
        cold_storage_after_days = optional(number)
        delete_after_days = optional(number)
        opt_in_to_archive_for_supported_resources = optional(bool)
      }))
      copy_actions = optional(list(object({
        destination_vault_arn = string
        lifecycle = optional(object({
          cold_storage_after_days = optional(number)
          delete_after_days = optional(number)
          opt_in_to_archive_for_supported_resources = optional(bool)
        }))
      })), [])
      scan_actions = optional(list(object({
        malware_scanner = optional(string, "")
        scan_mode = optional(string, "")
      })), [])
    }))
    advanced_backup_settings = optional(list(object({
      resource_type = optional(string, "")
      backup_options = optional(map(string), {})
    })), [])
    scan_setting = optional(object({
      malware_scanner = optional(string, "")
      resource_types = list(string)
      scanner_role_arn = string
    }))
    selections = optional(list(object({
      name = string
      iam_role_arn = string
      resources = optional(list(string), [])
      not_resources = optional(list(string), [])
      selection_tags = optional(list(object({
        type = optional(string, "")
        key = string
        value = string
      })), [])
      conditions = optional(object({
        string_equals = optional(list(object({
          key = string
          value = string
        })), [])
        string_not_equals = optional(list(object({
          key = string
          value = string
        })), [])
        string_like = optional(list(object({
          key = string
          value = string
        })), [])
        string_not_like = optional(list(object({
          key = string
          value = string
        })), [])
      }))
    })), [])
  })
}