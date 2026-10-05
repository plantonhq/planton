variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareZeroTrustDeviceDefaultProfile specification"
  type = object({
    account_id = string
    allow_mode_switch = optional(bool)
    allow_updates = optional(bool)
    allowed_to_leave = optional(bool)
    auto_connect = optional(number)
    captive_portal = optional(number)
    disable_auto_fallback = optional(bool)
    exclude_office_ips = optional(bool)
    register_interface_ip_with_dns = optional(bool)
    sccm_vpn_boundary_support = optional(bool)
    support_url = optional(string, "")
    switch_locked = optional(bool)
    tunnel_protocol = optional(string, "")
    lan_allow_minutes = optional(number)
    lan_allow_subnet_size = optional(number)
    exclude = optional(list(object({
      address = optional(string, "")
      host = optional(string, "")
      description = optional(string, "")
    })), [])
    include = optional(list(object({
      address = optional(string, "")
      host = optional(string, "")
      description = optional(string, "")
    })), [])
    service_mode_v2 = optional(object({
      mode = string
      port = optional(number)
    }))
    virtual_networks = optional(object({
      allowed = list(string)
      default_virtual_network_id = string
    }))
    dns_search_suffixes = optional(list(object({
      suffix = string
      description = optional(string, "")
    })), [])
    fallback_domains = optional(list(object({
      suffix = string
      description = optional(string, "")
      dns_server = optional(list(string), [])
    })), [])
    zone_certificates = optional(object({
      zone_id = string
      enabled = bool
    }))
  })
}