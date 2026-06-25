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
  description = "CloudflareEmailRoutingAddressSpec defines a verified destination address"
  type = object({
    # (Required) The Cloudflare account ID that owns the address.
    account_id = string

    # (Required) The destination email address. Immutable.
    email = string
  })
}
