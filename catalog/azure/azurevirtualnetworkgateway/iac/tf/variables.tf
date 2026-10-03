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
  description = "AzureVirtualNetworkGateway specification"
  type = object({
    region         = string
    resource_group = string
    name           = string
    type           = optional(string, "")
    vpn_type       = optional(string, "")
    sku            = optional(string, "")
    generation     = optional(string, "")
    ip_configurations = list(object({
      name                          = optional(string, "")
      subnet_id                     = string
      public_ip_address_id          = optional(string, "")
      private_ip_address_allocation = optional(string, "")
    }))
    active_active              = optional(bool, false)
    private_ip_address_enabled = optional(bool, false)
    edge_zone                  = optional(string, "")
    bgp_enabled                = optional(bool, false)
    bgp_settings = optional(object({
      asn         = optional(number, 0)
      peer_weight = optional(number, 0)
      peering_addresses = optional(list(object({
        ip_configuration_name = optional(string, "")
        apipa_addresses       = list(string)
      })), [])
    }))
    custom_route_address_prefixes    = optional(list(string), [])
    default_local_network_gateway_id = optional(string, "")
    vpn_client_configuration = optional(object({
      address_spaces = list(string)
      aad_tenant     = optional(string, "")
      aad_audience   = optional(string, "")
      aad_issuer     = optional(string, "")
      root_certificates = optional(list(object({
        name             = string
        public_cert_data = string
      })), [])
      revoked_certificates = optional(list(object({
        name       = string
        thumbprint = string
      })), [])
      radius_server_address = optional(string, "")
      radius_server_secret  = optional(string, "")
      radius_servers = optional(list(object({
        address = string
        secret  = string
        score   = optional(number, 0)
      })), [])
      ipsec_policy = optional(object({
        dh_group               = optional(string, "")
        ike_encryption         = optional(string, "")
        ike_integrity          = optional(string, "")
        ipsec_encryption       = optional(string, "")
        ipsec_integrity        = optional(string, "")
        pfs_group              = optional(string, "")
        sa_lifetime_seconds    = optional(number, 0)
        sa_data_size_kilobytes = optional(number, 0)
      }))
      vpn_client_protocols = optional(list(string), [])
      vpn_auth_types       = optional(list(string), [])
      client_connections = optional(list(object({
        name               = string
        policy_group_names = list(string)
        address_prefixes   = list(string)
      })), [])
    }))
    policy_groups = optional(list(object({
      name = string
      policy_members = list(object({
        name  = string
        type  = optional(string, "")
        value = string
      }))
      is_default = optional(bool, false)
      priority   = optional(number, 0)
    })), [])
    bgp_route_translation_for_nat_enabled = optional(bool, false)
    dns_forwarding_enabled                = optional(bool, false)
    ip_sec_replay_protection_enabled      = optional(bool)
    minimum_scale_unit                    = optional(number)
    maximum_scale_unit                    = optional(number)
    remote_vnet_traffic_enabled           = optional(bool, false)
    virtual_wan_traffic_enabled           = optional(bool, false)
    nat_rules = optional(list(object({
      name = string
      mode = optional(string, "")
      type = optional(string, "")
      external_mappings = list(object({
        address_space = string
        port_range    = optional(string, "")
      }))
      internal_mappings = list(object({
        address_space = string
        port_range    = optional(string, "")
      }))
      ip_configuration_id = optional(string, "")
    })), [])
    tags = optional(map(string), {})
  })
}