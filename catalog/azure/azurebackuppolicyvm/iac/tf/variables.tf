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
  description = "AzureBackupPolicyVm specification"
  type = object({
    resource_group      = string
    recovery_vault_name = string
    name                = string
    policy_type         = optional(string)
    backup = object({
      frequency     = string
      time          = string
      weekdays      = optional(list(string), [])
      hour_interval = optional(number)
      hour_duration = optional(number)
    })
    instant_restore_retention_days = optional(number)
    instant_restore_resource_group = optional(object({
      prefix = string
      suffix = optional(string, "")
    }))
    tiering_policy = optional(object({
      archived_restore_point = object({
        mode          = string
        duration      = optional(number)
        duration_type = optional(string, "")
      })
    }))
    consistency_type = optional(string, "")
    timezone         = optional(string)
    retention_daily = optional(object({
      count = number
    }))
    retention_weekly = optional(object({
      count    = number
      weekdays = list(string)
    }))
    retention_monthly = optional(object({
      count             = number
      weeks             = optional(list(string), [])
      weekdays          = optional(list(string), [])
      days              = optional(list(number), [])
      include_last_days = optional(bool, false)
    }))
    retention_yearly = optional(object({
      count             = number
      months            = list(string)
      weeks             = optional(list(string), [])
      weekdays          = optional(list(string), [])
      days              = optional(list(number), [])
      include_last_days = optional(bool, false)
    }))
  })
}