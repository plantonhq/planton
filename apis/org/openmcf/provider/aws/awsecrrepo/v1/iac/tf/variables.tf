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
  description = "AwsEcrRepo specification"
  type = object({
    region = string
    repository_name = string
    image_immutable = optional(bool, false)
    encryption_type = optional(string, "")
    kms_key_id = optional(string, "")
    force_delete = optional(bool, false)
    scan_on_push = optional(bool, false)
    lifecycle_policy = optional(object({
      expire_untagged_after_days = optional(number, 0)
      max_image_count = optional(number, 0)
    }))
  })
}
