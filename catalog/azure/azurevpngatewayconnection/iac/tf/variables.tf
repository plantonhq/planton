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
  description = "AzureVpnGatewayConnection specification"
  type = object({
    name                      = string
    vpn_gateway_id            = string
    remote_vpn_site_id        = string
    internet_security_enabled = optional(bool, false)
    routing = optional(object({
      associated_route_table_id = string
      inbound_route_map_id      = optional(string, "")
      outbound_route_map_id     = optional(string, "")
      propagated_route_table = optional(object({
        route_table_ids = list(string)
        labels          = optional(list(string), [])
      }))
    }))
    vpn_links = list(object({
      name                                  = string
      vpn_site_link_id                      = string
      bandwidth_mbps                        = optional(number)
      protocol                              = optional(string, "")
      connection_mode                       = optional(string, "")
      route_weight                          = optional(number, 0)
      dpd_timeout_seconds                   = optional(number)
      shared_key                            = optional(string, "")
      bgp_enabled                           = optional(bool, false)
      ratelimit_enabled                     = optional(bool, false)
      local_azure_ip_address_enabled        = optional(bool, false)
      policy_based_traffic_selector_enabled = optional(bool, false)
      egress_nat_rule_ids                   = optional(list(string), [])
      ingress_nat_rule_ids                  = optional(list(string), [])
      ipsec_policies = optional(list(object({
        sa_lifetime_sec          = optional(number, 0)
        sa_data_size_kb          = optional(number, 0)
        encryption_algorithm     = optional(string, "")
        integrity_algorithm      = optional(string, "")
        ike_encryption_algorithm = optional(string, "")
        ike_integrity_algorithm  = optional(string, "")
        dh_group                 = optional(string, "")
        pfs_group                = optional(string, "")
      })), [])
      custom_bgp_addresses = optional(list(object({
        ip_address          = string
        ip_configuration_id = optional(string, "")
      })), [])
    }))
    traffic_selector_policies = optional(list(object({
      local_address_cidrs  = list(string)
      remote_address_cidrs = list(string)
    })), [])
  })
}
