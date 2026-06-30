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
  description = "CloudflareListSpec defines an account-scoped Cloudflare List"
  type = object({
    # (Required) The Cloudflare account ID that owns the list.
    account_id = string

    # (Required) The list type: "ip", "redirect", "hostname", or "asn".
    # Immutable — changing it replaces the list.
    kind = string

    # (Required) The list name, used in rule expressions. Immutable.
    name = string

    # (Optional) Human-readable summary of the list's purpose.
    description = optional(string, "")
  })
}
