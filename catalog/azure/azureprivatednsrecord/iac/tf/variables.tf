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
  description = "AzurePrivateDnsRecord specification"
  type = object({
    private_dns_zone_id = string
    name                = string
    ttl_seconds         = optional(number)
    tags                = optional(map(string), {})
    a                   = optional(list(string), [])
    aaaa                = optional(list(string), [])
    cname               = optional(string, "")
    mx = optional(list(object({
      preference = number
      exchange   = string
    })), [])
    ptr = optional(list(string), [])
    srv = optional(list(object({
      priority = number
      weight   = number
      port     = number
      target   = string
    })), [])
    txt = optional(list(string), [])
  })
}
