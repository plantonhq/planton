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
  description = "AzureMonitorAutoscaleSetting specification"
  type = object({
    resource_group     = string
    name               = string
    region             = string
    target_resource_id = string
    enabled            = optional(bool)
    predictive = optional(object({
      scale_mode      = string
      look_ahead_time = optional(string, "")
    }))
    profiles = list(object({
      name = string
      capacity = object({
        minimum = number
        maximum = number
        default = number
      })
      rules = optional(list(object({
        metric_trigger = object({
          metric_name              = string
          metric_resource_id       = string
          time_grain               = string
          statistic                = string
          time_window              = string
          time_aggregation         = string
          operator                 = string
          threshold                = number
          metric_namespace         = optional(string, "")
          divide_by_instance_count = optional(bool, false)
          dimensions = optional(list(object({
            name     = string
            operator = string
            values   = list(string)
          })), [])
        })
        scale_action = object({
          direction = string
          type      = string
          value     = number
          cooldown  = string
        })
      })), [])
      fixed_date = optional(object({
        timezone = optional(string)
        start    = string
        end      = string
      }))
      recurrence = optional(object({
        timezone = optional(string)
        days     = list(string)
        hour     = number
        minute   = number
      }))
    }))
    notification = optional(object({
      email = optional(object({
        custom_emails = optional(list(string), [])
      }))
      webhooks = optional(list(object({
        service_uri = string
        properties  = optional(map(string), {})
      })), [])
    }))
    tags = optional(map(string), {})
  })
}
