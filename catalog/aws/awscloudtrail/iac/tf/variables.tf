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
  description = "AwsCloudTrail specification"
  type = object({
    region = string
    s3_bucket_name = string
    s3_key_prefix = optional(string, "")
    is_multi_region_trail = optional(bool, false)
    include_global_service_events = optional(bool)
    is_organization_trail = optional(bool, false)
    enable_logging = optional(bool)
    enable_log_file_validation = optional(bool, false)
    kms_key_id = optional(string, "")
    sns_topic_name = optional(string, "")
    cloudwatch_logs = optional(object({
      log_group_arn = string
      role_arn = string
    }))
    event_selectors = optional(list(object({
      read_write_type = optional(string, "")
      include_management_events = optional(bool)
      exclude_management_event_sources = optional(list(string), [])
      data_resources = optional(list(object({
        type = optional(string, "")
        values = list(string)
      })), [])
    })), [])
    advanced_event_selectors = optional(list(object({
      name = optional(string, "")
      field_selectors = list(object({
        field = optional(string, "")
        equals = optional(list(string), [])
        not_equals = optional(list(string), [])
        starts_with = optional(list(string), [])
        not_starts_with = optional(list(string), [])
        ends_with = optional(list(string), [])
        not_ends_with = optional(list(string), [])
      }))
    })), [])
    insight_types = optional(list(string), [])
    organization_delegated_admin_account_id = optional(string, "")
  })
}