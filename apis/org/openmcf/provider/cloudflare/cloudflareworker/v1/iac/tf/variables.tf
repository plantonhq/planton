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
  description = "Specification for Cloudflare Worker"
  type = object({
    # Cloudflare account ID (32 hex characters)
    account_id = string

    # Worker name (1-63 characters)
    worker_name = string

    # Worker script bundle configuration (R2 object reference)
    script_bundle = object({
      bucket = string  # R2 bucket name
      path   = string  # Path to bundle in R2
    })

    # Optional KV namespace bindings
    kv_bindings = optional(list(object({
      name       = string
      field_path = string  # Namespace ID reference
    })), [])

    # Optional DNS configuration for custom domain
    dns = optional(object({
      enabled = optional(bool, false)
      zone_id = optional(object({
        value = optional(string, "")
      }))
      hostname      = optional(string, "")
      route_pattern = optional(string, "")
    }))

    # Compatibility date (YYYY-MM-DD format)
    compatibility_date = optional(string, "")

    # Usage model: 0=BUNDLED, 1=UNBOUND
    usage_model = optional(number, 0)

    # Environment variables and secrets
    env = optional(object({
      variables = optional(map(string), {})
      secrets   = optional(map(string), {})
    }))
  })
}
