variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "DigitalOcean Spaces bucket specification"
  type = object({
    # Bucket name (DNS-compatible, 3-63 characters)
    bucket_name = string

    # DigitalOcean region for the bucket (e.g., nyc3, sfo3, ams3)
    region = string

    # Access control: PRIVATE (0) or PUBLIC_READ (1)
    # Default: PRIVATE (0)
    access_control = optional(number, 0)

    # Enable versioning for the bucket (cannot be disabled once enabled)
    # Default: false
    versioning_enabled = optional(bool, false)

    # Tags to apply to the bucket
    tags = optional(list(string), [])
  })

  validation {
    condition     = can(regex("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", var.spec.bucket_name))
    error_message = "bucket_name must be DNS-compatible: lowercase alphanumeric and hyphens only"
  }

  validation {
    condition     = length(var.spec.bucket_name) >= 3 && length(var.spec.bucket_name) <= 63
    error_message = "bucket_name must be between 3 and 63 characters"
  }

  validation {
    condition     = contains([0, 1], var.spec.access_control)
    error_message = "access_control must be 0 (PRIVATE) or 1 (PUBLIC_READ)"
  }
}

variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication"
  type        = string
  sensitive   = true
}

variable "spaces_access_id" {
  description = "DigitalOcean Spaces access key ID (for S3-compatible access)"
  type        = string
  sensitive   = true
  default     = null
}

variable "spaces_secret_key" {
  description = "DigitalOcean Spaces secret access key (for S3-compatible access)"
  type        = string
  sensitive   = true
  default     = null
}
