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
  description = "AwsDlmLifecyclePolicy specification"
  type = object({
    region = string
    description = optional(string, "")
    execution_role_arn = string
    disabled = optional(bool, false)
    default_policy = optional(object({
      resource_type = optional(string, "")
      create_interval_days = optional(number, 0)
      retain_interval_days = optional(number, 0)
      copy_tags = optional(bool, false)
      extend_deletion = optional(bool, false)
      exclusions = optional(object({
        exclude_boot_volumes = optional(bool, false)
        exclude_tags = optional(map(string), {})
        exclude_volume_types = optional(list(string), [])
      }))
    }))
    custom_policy = optional(object({
      policy_type = optional(string, "")
      resource_types = optional(list(string), [])
      resource_locations = optional(list(string), [])
      target_tags = optional(map(string), {})
      parameters = optional(object({
        exclude_boot_volume = optional(bool, false)
        no_reboot = optional(bool, false)
      }))
      schedules = optional(list(object({
        name = string
        copy_tags = optional(bool, false)
        tags_to_add = optional(map(string), {})
        variable_tags = optional(map(string), {})
        create_rule = object({
          interval_hours = optional(number, 0)
          times = optional(list(string), [])
          cron_expression = optional(string, "")
          location = optional(string, "")
          scripts = optional(object({
            execution_handler = string
            stages = optional(list(string), [])
            execution_handler_service = optional(string, "")
            execute_operation_on_script_failure = optional(bool, false)
            execution_timeout_seconds = optional(number, 0)
            maximum_retry_count = optional(number, 0)
          }))
        })
        retain_rule = object({
          count = optional(number, 0)
          interval = optional(number, 0)
          interval_unit = optional(string, "")
        })
        archive_rule = optional(object({
          count = optional(number, 0)
          interval = optional(number, 0)
          interval_unit = optional(string, "")
        }))
        cross_region_copy_rules = optional(list(object({
          target_region = string
          encrypted = optional(bool, false)
          cmk_arn = optional(string, "")
          copy_tags = optional(bool, false)
          retain_rule = optional(object({
            interval = optional(number, 0)
            interval_unit = optional(string, "")
          }))
          deprecate_rule = optional(object({
            interval = optional(number, 0)
            interval_unit = optional(string, "")
          }))
        })), [])
        deprecate_rule = optional(object({
          count = optional(number, 0)
          interval = optional(number, 0)
          interval_unit = optional(string, "")
        }))
        fast_restore_rule = optional(object({
          availability_zones = list(string)
          count = optional(number, 0)
          interval = optional(number, 0)
          interval_unit = optional(string, "")
        }))
        share_rule = optional(object({
          target_accounts = list(string)
          unshare_interval = optional(number, 0)
          unshare_interval_unit = optional(string, "")
        }))
      })), [])
      event_source = optional(object({
        event_type = optional(string, "")
        description_regex = string
        snapshot_owners = list(string)
      }))
      action = optional(object({
        name = string
        cross_region_copies = list(object({
          target = string
          encrypted = optional(bool, false)
          cmk_arn = optional(string, "")
          retain_rule = optional(object({
            interval = optional(number, 0)
            interval_unit = optional(string, "")
          }))
        }))
      }))
    }))
  })
}