variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCS bucket"
  type = object({
    # The ID of the GCP project where the storage bucket will be created.
    # Uses StringValueOrRef pattern for cross-resource references.
    gcp_project_id = object({
      value = string
    })

    # The location for the bucket. Can be a region (e.g., "us-east1"), dual-region (e.g., "NAM4"),
    # or multi-region (e.g., "US", "EU", "ASIA").
    location = string

    # Enable Uniform Bucket-Level Access (UBLA) for simplified IAM-only access control.
    uniform_bucket_level_access_enabled = optional(bool, true)

    # Storage class for the bucket: STANDARD, NEARLINE, COLDLINE, or ARCHIVE.
    storage_class = optional(string, "STANDARD")

    # Enable object versioning to protect against accidental deletion/overwrite.
    versioning_enabled = optional(bool, false)

    # Lifecycle rules for automatic object management.
    lifecycle_rules = optional(list(object({
      action = object({
        type          = string
        storage_class = optional(string)
      })
      condition = object({
        age_days               = optional(number)
        created_before         = optional(string)
        is_live                = optional(bool)
        num_newer_versions     = optional(number)
        matches_storage_class  = optional(list(string))
      })
    })), [])

    # IAM bindings for bucket-level access control.
    iam_bindings = optional(list(object({
      role      = string
      members   = list(string)
      condition = optional(string)
    })), [])

    # Encryption configuration using Customer-Managed Encryption Keys (CMEK).
    encryption = optional(object({
      kms_key_name = string
    }))

    # CORS rules for cross-origin browser access.
    cors_rules = optional(list(object({
      methods          = list(string)
      origins          = list(string)
      response_headers = optional(list(string))
      max_age_seconds  = optional(number)
    })), [])

    # Website configuration for static website hosting.
    website = optional(object({
      main_page_suffix = optional(string)
      not_found_page   = optional(string)
    }))

    # Retention policy for WORM (Write Once, Read Many) compliance.
    retention_policy = optional(object({
      retention_period_seconds = number
      is_locked                = optional(bool, false)
    }))

    # Enable requester pays mode.
    requester_pays = optional(bool, false)

    # Logging configuration for legacy access logs.
    logging = optional(object({
      log_bucket        = string
      log_object_prefix = optional(string)
    }))

    # Public access prevention policy: "inherited" or "enforced".
    public_access_prevention = optional(string)

    # Custom labels for the bucket (cost tracking, governance, compliance).
    gcp_labels = optional(map(string), {})

    # Name of the GCS bucket to create in GCP.
    bucket_name = string
  })

  validation {
    condition     = can(regex("^[a-z0-9]([a-z0-9-._]*[a-z0-9])?$", var.spec.bucket_name)) && length(var.spec.bucket_name) >= 3 && length(var.spec.bucket_name) <= 63
    error_message = "Bucket name must be 3-63 characters, globally unique, lowercase letters, numbers, hyphens, or dots, starting and ending with a letter or number."
  }
}
