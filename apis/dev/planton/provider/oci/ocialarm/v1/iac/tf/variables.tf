variable "metadata" {
  type = object({
    name   = string
    id     = optional(string, "")
    org    = optional(string, "")
    env    = optional(string, "")
    labels = optional(map(string), {})
  })
}

variable "spec" {
  type = object({
    compartment_id = object({
      value = string
    })
    metric_compartment_id = object({
      value = string
    })
    namespace    = string
    query        = string
    severity     = string
    destinations = list(string)
    is_enabled   = optional(bool, false)

    body                          = optional(string, "")
    alarm_summary                 = optional(string, "")
    notification_title            = optional(string, "")
    pending_duration              = optional(string, "")
    evaluation_slack_duration     = optional(string, "")
    repeat_notification_duration  = optional(string, "")
    message_format                = optional(string, "raw")
    metric_compartment_id_in_subtree                  = optional(bool)
    is_notifications_per_metric_dimension_enabled     = optional(bool)
    resource_group                = optional(string, "")
    notification_version          = optional(string, "")
    rule_name                     = optional(string, "")

    overrides = optional(list(object({
      rule_name        = string
      query            = optional(string, "")
      severity         = optional(string, "")
      body             = optional(string, "")
      pending_duration = optional(string, "")
    })), [])
  })
}
