variable "metadata" {
  description = "metadata"
  type = object({
    # name of the resource
    name = string
    # id of the resource
    id = string
    # id of the organization to which the api-resource belongs to
    org = string
    # environment to which the resource belongs to
    env = string
    # labels for the resource
    labels = map(string)
    # annotations for the resource
    annotations = map(string)
    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "spec"
  type = object({
    # The AWS region where the S3 bucket will be created.
    region = string

    # Flag to indicate if the S3 bucket should have external (public) access
    is_public = optional(bool, false)

    # Enable versioning to protect against accidental deletions and overwrites
    versioning_enabled = optional(bool, false)

    # Encryption type for objects in the bucket (SSE_S3 or SSE_KMS)
    encryption_type = optional(string, "SSE_S3")

    # KMS key ID or ARN for SSE-KMS encryption (required when encryption_type is SSE_KMS) (StringValueOrRef)
    kms_key_id = optional(object({ value = string }), null)

    # Tags for resource governance, cost allocation, and organization
    tags = optional(map(string), {})

    # Lifecycle rules for automatic storage transitions and expiration
    lifecycle_rules = optional(list(object({
      id                                    = string
      enabled                               = optional(bool, false)
      prefix                                = optional(string, "")
      transition_days                       = optional(number, 0)
      transition_storage_class              = optional(string, "")
      expiration_days                       = optional(number, 0)
      noncurrent_version_expiration_days    = optional(number, 0)
      abort_incomplete_multipart_upload_days = optional(number, 0)
    })), [])

    # Replication configuration for disaster recovery or compliance
    replication = optional(object({
      enabled = bool
      role_arn = object({ value = string })
      destination = object({
        bucket_arn    = string
        storage_class = optional(string, "")
        account_id    = optional(string, "")
      })
      prefix   = optional(string, "")
      priority = optional(number, 0)
    }), null)

    # Server access logging configuration
    logging = optional(object({
      enabled       = bool
      target_bucket = string
      target_prefix = optional(string, "")
    }), null)

    # CORS configuration for web applications
    cors = optional(object({
      cors_rules = list(object({
        allowed_methods = list(string)
        allowed_origins = list(string)
        allowed_headers = optional(list(string), [])
        expose_headers  = optional(list(string), [])
        max_age_seconds = optional(number, 0)
      }))
    }), null)

    # Force destroy the bucket even if it contains objects
    force_destroy = optional(bool, false)
  })
}
