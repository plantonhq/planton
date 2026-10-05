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
  description = "AwsCloudwatchLogDelivery specification"
  type = object({
    region = string
    vended = optional(object({
      source = optional(object({
        name = string
        log_type = string
        resource_arn = string
      }))
      destinations = optional(any, [])
      deliveries = optional(list(object({
        name = string
        destination_name = optional(string, "")
        destination_arn = optional(string, "")
        record_fields = optional(list(string), [])
        field_delimiter = optional(string, "")
        s3_configuration = optional(object({
          enable_hive_compatible_path = optional(bool, false)
          suffix_path = optional(string, "")
        }))
      })), [])
    }))
    cross_account_destination = optional(object({
      name = string
      role_arn = string
      target_arn = string
      access_policy = any
      force_update = optional(bool)
    }))
  })
}