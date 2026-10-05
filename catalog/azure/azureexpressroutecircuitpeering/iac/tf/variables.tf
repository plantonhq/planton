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
  description = "AzureExpressRouteCircuitPeering specification"
  type = object({
    resource_group                = string
    express_route_circuit_name    = string
    peering_type                  = optional(string, "")
    vlan_id                       = optional(number, 0)
    primary_peer_address_prefix   = optional(string, "")
    secondary_peer_address_prefix = optional(string, "")
    ipv4_enabled                  = optional(bool)
    peer_asn                      = optional(number, 0)
    shared_key                    = optional(string, "")
    microsoft_peering_config = optional(object({
      advertised_public_prefixes = list(string)
      customer_asn               = optional(number, 0)
      routing_registry_name      = optional(string)
      advertised_communities     = optional(list(string), [])
    }))
    ipv6 = optional(object({
      primary_peer_address_prefix   = string
      secondary_peer_address_prefix = string
      enabled                       = optional(bool)
      route_filter_id               = optional(string, "")
      microsoft_peering = optional(object({
        advertised_public_prefixes = list(string)
        customer_asn               = optional(number, 0)
        routing_registry_name      = optional(string)
        advertised_communities     = optional(list(string), [])
      }))
    }))
    route_filter_id = optional(string, "")
    connections = optional(list(object({
      name                = string
      peer_peering_id     = string
      address_prefix_ipv4 = string
      address_prefix_ipv6 = optional(string, "")
      authorization_key   = optional(string, "")
    })), [])
  })
}