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
  description = "AzureLocalNetworkGateway specification"
  type = object({
    region          = string
    resource_group  = string
    name            = string
    gateway_address = optional(string, "")
    gateway_fqdn    = optional(string, "")
    address_spaces  = optional(list(string), [])
    bgp_settings = optional(object({
      asn                 = optional(number, 0)
      bgp_peering_address = string
      peer_weight         = optional(number, 0)
    }))
    tags = optional(map(string), {})
  })
}