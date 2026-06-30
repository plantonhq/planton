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
  description = "HetznerCloudCertificate specification"
  type = object({
    uploaded = optional(object({
      certificate = string
      private_key = string
    }))
    managed = optional(object({
      domain_names = list(string)
    }))
  })

  validation {
    condition = (
      (var.spec.uploaded != null ? 1 : 0) +
      (var.spec.managed != null ? 1 : 0)
    ) == 1
    error_message = "Exactly one of 'uploaded' or 'managed' must be set."
  }
}

variable "hcloud_token" {
  description = "Hetzner Cloud API token for authentication"
  type        = string
  sensitive   = true
}
