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
  description = "AwsConfigRecorder specification"
  type = object({
    region = string
    role_arn = string
    recording_enabled = optional(bool)
    recording_group = optional(object({
      all_supported = optional(bool)
      include_global_resource_types = optional(bool)
      resource_types = optional(list(string), [])
      exclusion_by_resource_types = optional(list(string), [])
      recording_strategy = optional(string, "")
    }))
    recording_mode = optional(object({
      recording_frequency = optional(string, "")
      override = optional(object({
        description = optional(string, "")
        recording_frequency = optional(string, "")
        resource_types = list(string)
      }))
    }))
    delivery_channel = optional(object({
      s3_bucket_name = string
      s3_key_prefix = optional(string, "")
      s3_kms_key_arn = optional(string, "")
      sns_topic_arn = optional(string, "")
      snapshot_delivery_frequency = optional(string, "")
    }))
    retention_period_in_days = optional(number, 0)
  })
}