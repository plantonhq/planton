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
  description = "Specification for the Cloudflare Workers KV namespace"
  type = object({
    # Human-readable name (title) for the KV namespace.
    namespace_name = string

    # Cloudflare account ID (32 hex characters) that owns the namespace.
    account_id = string

    # Default time-to-live for entries, in seconds. Not represented on the
    # Cloudflare KV namespace resource itself; kept for spec completeness.
    ttl_seconds = optional(number)

    # Free-form description. Not represented on the Cloudflare KV namespace
    # resource itself; kept for spec completeness.
    description = optional(string)
  })
}
