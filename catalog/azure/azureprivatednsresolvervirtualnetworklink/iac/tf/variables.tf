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
  description = "AzurePrivateDnsResolverVirtualNetworkLink specification"
  type = object({
    name                      = string
    dns_forwarding_ruleset_id = string
    virtual_network_id        = string
    metadata                  = optional(map(string), {})
  })
}
