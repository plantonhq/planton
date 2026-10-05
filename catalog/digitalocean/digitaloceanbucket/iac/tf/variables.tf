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
  description = "DigitalOceanBucket specification"
  type = object({
    bucket_name = string
    region = optional(string, "")
    access_control = optional(string, "")
    versioning_enabled = optional(bool, false)
    force_destroy = optional(bool, false)
    lifecycle_rules = optional(list(object({
      id = optional(string, "")
      prefix = optional(string, "")
      enabled = bool
      abort_incomplete_multipart_upload_days = optional(number, 0)
      expiration = optional(object({
        date = optional(string, "")
        days = optional(number, 0)
        expired_object_delete_marker = optional(bool, false)
      }))
      noncurrent_version_expiration = optional(object({
        days = number
      }))
    })), [])
    cors_rules = optional(list(object({
      allowed_methods = list(string)
      allowed_origins = list(string)
      allowed_headers = optional(list(string), [])
      expose_headers = optional(list(string), [])
      id = optional(string, "")
      max_age_seconds = optional(number, 0)
    })), [])
    policy = optional(string, "")
    logging = optional(object({
      target_bucket = string
      target_prefix = string
    }))
  })
}