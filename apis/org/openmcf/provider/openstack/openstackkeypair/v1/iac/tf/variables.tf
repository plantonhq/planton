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
  description = "OpenStackKeypairSpec defines the configuration for an SSH keypair"
  type = object({
    # (Optional) SSH public key in OpenSSH authorized_keys format.
    # If omitted, OpenStack generates a new keypair.
    public_key = optional(string, "")

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
