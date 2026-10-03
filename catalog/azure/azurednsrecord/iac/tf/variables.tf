variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AzureDnsRecord specification"
  type = object({
    resource_group = string
    zone_name      = string
    name           = string
    ttl_seconds    = optional(number)
    tags           = optional(map(string), {})
    a = optional(object({
      addresses          = optional(list(string), [])
      target_resource_id = optional(string, "")
    }))
    aaaa = optional(object({
      addresses          = optional(list(string), [])
      target_resource_id = optional(string, "")
    }))
    cname = optional(object({
      value              = optional(string, "")
      target_resource_id = optional(string, "")
    }))
    mx = optional(list(object({
      preference = number
      exchange   = string
    })), [])
    srv = optional(list(object({
      priority = number
      weight   = number
      port     = number
      target   = string
    })), [])
    caa = optional(list(object({
      flags = number
      tag   = string
      value = string
    })), [])
    txt = optional(list(string), [])
    ns  = optional(list(string), [])
    ptr = optional(list(string), [])
  })
}