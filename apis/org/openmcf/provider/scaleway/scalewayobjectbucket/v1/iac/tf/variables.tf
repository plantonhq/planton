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
  description = "Scaleway Object Bucket specification"
  type = object({
    # Region where the bucket will be created (e.g., "fr-par", "nl-ams", "pl-waw")
    region = string

    # Enable S3-compatible versioning
    versioning_enabled = optional(bool, false)

    # Enable S3 Object Lock (requires versioning, cannot be removed)
    object_lock_enabled = optional(bool, false)

    # Lifecycle rules for automated object management
    lifecycle_rules = optional(list(object({
      id      = string
      enabled = bool
      prefix  = optional(string, "")
      tags    = optional(map(string), {})

      expiration_days = optional(number, 0)

      transitions = optional(list(object({
        days          = number
        storage_class = string
      })), [])

      abort_incomplete_multipart_upload_days = optional(number, 0)
    })), [])

    # CORS rules for web application cross-origin access
    cors_rules = optional(list(object({
      allowed_methods = list(string)
      allowed_origins = list(string)
      allowed_headers = optional(list(string), [])
      expose_headers  = optional(list(string), [])
      max_age_seconds = optional(number, 0)
    })), [])

    # Allow bucket deletion even with objects inside
    force_destroy = optional(bool, false)
  })
}

variable "scaleway_access_key" {
  description = "Scaleway access key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_secret_key" {
  description = "Scaleway secret key for API authentication"
  type        = string
  sensitive   = true
}

variable "scaleway_project_id" {
  description = "Scaleway project ID (optional, defaults from provider)"
  type        = string
  default     = ""
}

variable "scaleway_organization_id" {
  description = "Scaleway organization ID (optional, defaults from provider)"
  type        = string
  default     = ""
}
