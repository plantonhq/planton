variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud OSS bucket specification"
  type = object({
    region          = string
    bucket_name     = string
    acl             = optional(string, "private")
    storage_class   = optional(string, "Standard")
    redundancy_type = optional(string, "LRS")

    versioning_enabled = optional(bool, false)

    server_side_encryption = optional(object({
      sse_algorithm      = string
      kms_master_key_id  = optional(string, "")
    }))

    lifecycle_rules = optional(list(object({
      prefix                              = optional(string, "")
      enabled                             = bool
      expiration_days                     = optional(number, 0)
      transitions = optional(list(object({
        days          = number
        storage_class = string
      })), [])
      abort_multipart_upload_days         = optional(number, 0)
      noncurrent_version_expiration_days  = optional(number, 0)
    })), [])

    cors_rules = optional(list(object({
      allowed_origins = list(string)
      allowed_methods = list(string)
      allowed_headers = optional(list(string), [])
      expose_headers  = optional(list(string), [])
      max_age_seconds = optional(number, 0)
    })), [])

    logging = optional(object({
      target_bucket = string
      target_prefix = optional(string, "")
    }))

    force_destroy     = optional(bool, false)
    resource_group_id = optional(string, "")
    tags              = optional(map(string), {})
  })

  validation {
    condition     = length(var.spec.bucket_name) >= 3 && length(var.spec.bucket_name) <= 63
    error_message = "bucket_name must be between 3 and 63 characters."
  }

  validation {
    condition     = contains(["private", "public-read", "public-read-write"], var.spec.acl)
    error_message = "acl must be one of: private, public-read, public-read-write."
  }

  validation {
    condition     = contains(["Standard", "IA", "Archive", "ColdArchive", "DeepColdArchive"], var.spec.storage_class)
    error_message = "storage_class must be one of: Standard, IA, Archive, ColdArchive, DeepColdArchive."
  }

  validation {
    condition     = contains(["LRS", "ZRS"], var.spec.redundancy_type)
    error_message = "redundancy_type must be one of: LRS, ZRS."
  }
}
