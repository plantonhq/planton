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
  description = "DigitalOcean DNS Record specification"
  type = object({
    domain = object({
      value      = optional(string)
      value_from = optional(object({
        kind       = optional(string)
        env        = optional(string)
        name       = optional(string)
        field_path = optional(string)
      }))
    })
    name = string
    type = string
    value = object({
      value      = optional(string)
      value_from = optional(object({
        kind       = optional(string)
        env        = optional(string)
        name       = optional(string)
        field_path = optional(string)
      }))
    })
    ttl_seconds = optional(number)
    priority    = optional(number)
    weight      = optional(number)
    port        = optional(number)
    flags       = optional(number)
    tag         = optional(string)
  })
}

variable "digitalocean_token" {
  description = "DigitalOcean API token for authentication"
  type        = string
  sensitive   = true
}
