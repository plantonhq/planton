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
  description = "CloudflareHyperdriveConfigSpec defines a Cloudflare Hyperdrive connection pooler + cache"
  type = object({
    # (Required) The Cloudflare account ID that owns this Hyperdrive config.
    account_id = string

    # (Required) Human-readable name for the Hyperdrive config.
    name = string

    # (Required) Origin database connection details.
    origin = object({
      database  = string
      scheme    = string
      user      = string
      host      = optional(string)
      port      = optional(number, 0)
      # StringValueOrRef secrets are flattened to plain strings by the tfvars converter.
      password             = string
      access_client_id     = optional(string, "")
      access_client_secret = optional(string, "")
      # Workers VPC Service to egress through (mutually exclusive with mtls).
      service_id = optional(string, "")
    })

    # (Optional) Query-result caching behavior.
    caching = optional(object({
      disabled               = optional(bool, false)
      max_age                = optional(number, 0)
      stale_while_revalidate = optional(number, 0)
    }))

    # (Optional) Mutual-TLS configuration for the origin connection.
    mtls = optional(object({
      ca_certificate_id   = optional(string, "")
      mtls_certificate_id = optional(string, "")
      sslmode             = optional(string, "")
    }))

    # (Optional) Maximum pooled connections to the origin (0 = plan default).
    origin_connection_limit = optional(number, 0)
  })
}
