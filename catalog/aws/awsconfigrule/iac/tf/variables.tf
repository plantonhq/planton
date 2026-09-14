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
  description = "AwsConfigRule specification"
  type = object({
    region                      = string
    description                 = optional(string, "")
    input_parameters            = optional(string, "")
    maximum_execution_frequency = optional(string, "")
    managed = optional(object({
      rule_identifier = string
    }))
    custom_lambda = optional(object({
      function_arn = string
      source_details = optional(list(object({
        message_type                = optional(string, "")
        maximum_execution_frequency = optional(string, "")
      })), [])
    }))
    custom_policy = optional(object({
      policy_runtime            = optional(string, "")
      policy_text               = string
      enable_debug_log_delivery = optional(bool, false)
    }))
    scope = optional(object({
      compliance_resource_id    = optional(string, "")
      compliance_resource_types = optional(list(string), [])
      tag_key                   = optional(string, "")
      tag_value                 = optional(string, "")
    }))
    evaluation_modes = optional(list(string), [])
    organization = optional(object({
      enabled                     = optional(bool)
      excluded_accounts           = optional(list(string), [])
      trigger_types               = optional(list(string), [])
      debug_log_delivery_accounts = optional(list(string), [])
    }))
    remediation = optional(object({
      automatic      = optional(bool, false)
      target_id      = string
      target_version = optional(string, "")
      resource_type  = optional(string, "")
      parameters = optional(list(object({
        name           = string
        resource_value = optional(string, "")
        static_value   = optional(string, "")
        static_values  = optional(list(string), [])
      })), [])
      maximum_automatic_attempts           = optional(number, 0)
      retry_attempt_seconds                = optional(number, 0)
      concurrent_execution_rate_percentage = optional(number, 0)
      error_percentage                     = optional(number, 0)
    }))
  })
}