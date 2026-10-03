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
  description = "AwsBackupVault specification"
  type = object({
    region = string
    standard = optional(object({
      kms_key_arn = optional(string, "")
      force_destroy = optional(bool, false)
      lock = optional(object({
        changeable_for_days = optional(number)
        min_retention_days = optional(number)
        max_retention_days = optional(number)
      }))
      policy = optional(any)
      notifications = optional(object({
        sns_topic_arn = string
        events = list(string)
      }))
    }))
    air_gapped = optional(object({
      min_retention_days = optional(number, 0)
      max_retention_days = optional(number, 0)
      encryption_key_arn = optional(string, "")
    }))
  })
}