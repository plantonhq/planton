variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id   = optional(string, "")
    org  = optional(string, "")
    env  = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "OciSubnet specification"
  type = object({
    compartment_id = object({
      value = string
    })

    vcn_id = object({
      value = string
    })

    cidr_block = string

    display_name = optional(string, "")

    dns_label = optional(string, "")

    availability_domain = optional(string, "")

    prohibit_public_ip_on_vnic = optional(bool, false)

    prohibit_internet_ingress = optional(bool, false)

    dhcp_options_id = optional(object({
      value = string
    }), null)

    route_table_id = optional(object({
      value = string
    }), null)

    security_list_ids = optional(list(object({
      value = string
    })), [])

    ipv6_cidr_block = optional(string, "")

    route_rules = optional(list(object({
      destination      = string
      destination_type = optional(string, "cidr_block")
      network_entity_id = object({
        value = string
      })
      description = optional(string, "")
    })), [])
  })
}
