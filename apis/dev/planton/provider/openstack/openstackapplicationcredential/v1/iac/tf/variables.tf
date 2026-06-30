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
  description = "OpenStackApplicationCredentialSpec defines the configuration for an application credential"
  type = object({
    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Allow sub-credential creation. Default: false.
    unrestricted = optional(bool, false)

    # (Optional) User-provided secret. Auto-generated if omitted.
    secret = optional(string, "")

    # (Optional) Role names to scope the credential.
    roles = optional(list(string), [])

    # (Optional) Fine-grained API access rules.
    access_rules = optional(list(object({
      path    = string
      method  = string
      service = string
    })), [])

    # (Optional) Expiration timestamp in RFC3339 format.
    expires_at = optional(string, "")

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
