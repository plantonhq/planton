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
  description = "CloudflareCustomHostnameFallbackOriginSpec defines a zone's fallback origin"
  type = object({
    # (Required) The SaaS Zone ID. StringValueOrRef is flattened to a plain string.
    zone_id = optional(string)

    # (Required) The fallback origin hostname. StringValueOrRef -> plain string.
    origin = optional(string)
  })
}
