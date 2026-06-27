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
  description = "AwsEcsCluster specification"
  type = object({
    region = string
    enable_container_insights = optional(bool, false)
    capacity_providers = optional(list(string), [])
    default_capacity_provider_strategy = optional(list(object({
      capacity_provider = optional(string, "")
      base = optional(number, 0)
      weight = optional(number, 0)
    })), [])
    execute_command_configuration = optional(object({
      logging = optional(string, "")
      log_configuration = optional(object({
        cloud_watch_log_group_name = optional(string, "")
        cloud_watch_encryption_enabled = optional(bool, false)
        s3_bucket_name = optional(string, "")
        s3_key_prefix = optional(string, "")
        s3_encryption_enabled = optional(bool, false)
      }))
      kms_key_id = optional(string, "")
    }))
  })
}
