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
  description = "AzureBackupPolicyFileShare specification"
  type = object({
    resource_group      = string
    recovery_vault_name = string
    name                = string
    backup = object({
      frequency = string
      time      = optional(string, "")
      hourly = optional(object({
        interval        = number
        start_time      = string
        window_duration = number
      }))
    })
    backup_tier                = optional(string)
    snapshot_retention_in_days = optional(number)
    timezone                   = optional(string)
    retention_daily = object({
      count = number
    })
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