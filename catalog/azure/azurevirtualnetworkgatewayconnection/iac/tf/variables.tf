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
  description = "AzureVirtualNetworkGatewayConnection specification"
  type = object({
    region                          = string
    resource_group                  = string
    name                            = string
    type                            = optional(string, "")
    virtual_network_gateway_id      = string
    local_network_gateway_id        = optional(string, "")
    peer_virtual_network_gateway_id = optional(string, "")
    express_route_circuit_id        = optional(string, "")
    shared_key                      = optional(string, "")
    authorization_key               = optional(string, "")
    bgp_enabled                     = optional(bool, false)
    custom_bgp_addresses = optional(object({
      primary   = string
      secondary = optional(string, "")
    }))
    dpd_timeout_seconds                = optional(number)
    connection_protocol                = optional(string, "")
    connection_mode                    = optional(string, "")
    routing_weight                     = optional(number)
    egress_nat_rule_ids                = optional(list(string), [])
    ingress_nat_rule_ids               = optional(list(string), [])
    use_policy_based_traffic_selectors = optional(bool, false)
    express_route_gateway_bypass       = optional(bool, false)
    private_link_fast_path_enabled     = optional(bool, false)
    local_azure_ip_address_enabled     = optional(bool, false)
    traffic_selector_policies = optional(list(object({
      local_address_cidrs  = list(string)
      remote_address_cidrs = list(string)
    })), [])
    ipsec_policy = optional(object({
      dh_group         = optional(string, "")
      ike_encryption   = optional(string, "")
      ike_integrity    = optional(string, "")
      ipsec_encryption = optional(string, "")
      ipsec_integrity  = optional(string, "")
      pfs_group        = optional(string, "")
      sa_datasize      = optional(number)
      sa_lifetime      = optional(number)
    }))
    tags = optional(map(string), {})
  })
}