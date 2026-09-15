variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsSesConfigurationSet specification"
  type = object({
    region = string
    delivery_options = optional(object({
      tls_policy = optional(string)
      max_delivery_seconds = optional(number)
      sending_pool_name = optional(string, "")
    }))
    reputation_metrics_enabled = optional(bool, false)
    sending_enabled = optional(bool)
    suppressed_reasons = optional(list(string), [])
    tracking_options = optional(object({
      custom_redirect_domain = string
      https_policy = optional(string)
    }))
    vdm_options = optional(object({
      engagement_metrics_enabled = optional(bool)
      optimized_shared_delivery_enabled = optional(bool)
    }))
    event_destinations = optional(list(object({
      name = string
      enabled = optional(bool)
      matching_event_types = list(string)
      cloud_watch = optional(object({
        dimensions = list(object({
          name = string
          value_source = string
          default_value = string
        }))
      }))
      event_bus = optional(string, "")
      firehose = optional(object({
        delivery_stream = string
        iam_role = string
      }))
      sns_topic = optional(string, "")
      pinpoint_application_arn = optional(string, "")
    })), [])
  })
}
