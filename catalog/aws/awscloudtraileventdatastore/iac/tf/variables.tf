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
  description = "AwsCloudTrailEventDataStore specification"
  type = object({
    region = string
    billing_mode = optional(string, "")
    kms_key_id = optional(string, "")
    multi_region_enabled = optional(bool)
    organization_enabled = optional(bool, false)
    retention_period_days = optional(number, 0)
    termination_protection_enabled = optional(bool)
    suspend = optional(bool)
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
  })
}
