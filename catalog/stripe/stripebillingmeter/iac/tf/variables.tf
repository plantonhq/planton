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
  description = "StripeBillingMeter specification"
  type = object({
    display_name = string
    event_name   = string
    default_aggregation = object({
      formula = optional(string, "")
    })
    customer_mapping = optional(object({
      event_payload_key = string
    }))
    value_settings = optional(object({
      event_payload_key = string
    }))
    event_time_window = optional(string, "")
    alerts = optional(list(object({
      title      = string
      gte        = optional(number, 0)
      customer   = optional(string, "")
      recurrence = optional(string, "")
    })), [])
  })
}
