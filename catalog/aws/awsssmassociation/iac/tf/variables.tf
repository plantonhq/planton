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
  description = "AwsSsmAssociation specification"
  type = object({
    region = string
    document_name = string
    association_name = optional(string, "")
    document_version = optional(string, "")
    parameters = optional(map(string), {})
    targets = optional(list(object({
      key = string
      values = list(string)
    })), [])
    schedule_expression = optional(string, "")
    apply_only_at_cron_interval = optional(bool, false)
    compliance_severity = optional(string, "")
    sync_compliance = optional(string, "")
    max_concurrency = optional(string, "")
    max_errors = optional(string, "")
    automation_target_parameter_name = optional(string, "")
    calendar_names = optional(list(string), [])
    output_location = optional(object({
      s3_bucket_name = string
      s3_key_prefix = optional(string, "")
      s3_region = optional(string, "")
    }))
    wait_for_success_timeout_seconds = optional(number, 0)
  })
}
