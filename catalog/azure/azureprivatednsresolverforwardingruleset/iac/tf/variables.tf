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
  description = "AzurePrivateDnsResolverForwardingRuleset specification"
  type = object({
    region                = string
    resource_group        = string
    name                  = string
    outbound_endpoint_ids = list(string)
    forwarding_rules = optional(list(object({
      name        = string
      domain_name = string
      target_dns_servers = list(object({
        ip_address = string
        port       = optional(number)
      }))
      enabled  = optional(bool)
      metadata = optional(map(string), {})
    })), [])
    tags = optional(map(string), {})
  })
}
