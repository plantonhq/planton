variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name   = string
    id     = optional(string)
    org    = optional(string)
    env    = optional(string)
    labels = optional(map(string))
    tags   = optional(list(string))
  })
}

variable "spec" {
  description = "Alibaba Cloud VPN Gateway specification"
  type = object({
    region           = string
    vpc_id           = string
    vswitch_id       = string
    vpn_gateway_name = string
    description      = optional(string, "")
    bandwidth        = number
    payment_type     = optional(string, "PayAsYouGo")
    enable_ssl       = optional(bool, false)
    ssl_connections  = optional(number)
    tags             = optional(map(string), {})
    resource_group_id = optional(string, "")
    connections = optional(list(object({
      name                = string
      customer_gateway_ip = string
      customer_gateway_asn = optional(string, "")
      local_subnets       = list(string)
      remote_subnets      = list(string)
      enable_dpd          = optional(bool, true)
      enable_nat_traversal = optional(bool, true)
      effect_immediately  = optional(bool, true)
      ike_config = optional(object({
        psk          = optional(string, "")
        ike_version  = optional(string, "ikev2")
        ike_mode     = optional(string, "main")
        ike_enc_alg  = optional(string, "aes")
        ike_auth_alg = optional(string, "sha1")
        ike_pfs      = optional(string, "group2")
        ike_lifetime = optional(number, 86400)
      }))
      ipsec_config = optional(object({
        ipsec_enc_alg  = optional(string, "aes")
        ipsec_auth_alg = optional(string, "md5")
        ipsec_pfs      = optional(string, "group2")
        ipsec_lifetime = optional(number, 86400)
      }))
      health_check_config = optional(object({
        enable   = optional(bool, false)
        sip      = optional(string, "")
        dip      = optional(string, "")
        interval = optional(number, 3)
        retry    = optional(number, 3)
      }))
    })), [])
  })

  validation {
    condition     = length(var.spec.vpn_gateway_name) >= 2 && length(var.spec.vpn_gateway_name) <= 128
    error_message = "vpn_gateway_name must be between 2 and 128 characters."
  }

  validation {
    condition     = contains([5, 10, 20, 50, 100, 200, 500, 1000], var.spec.bandwidth)
    error_message = "bandwidth must be one of: 5, 10, 20, 50, 100, 200, 500, 1000 Mbps."
  }

  validation {
    condition     = contains(["PayAsYouGo", "Subscription"], var.spec.payment_type)
    error_message = "payment_type must be one of: PayAsYouGo, Subscription."
  }
}
