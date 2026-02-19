variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "OciNetworkSecurityGroup specification"
  type = object({
    compartment_id = object({
      value = string
    })

    vcn_id = object({
      value = string
    })

    display_name = optional(string, "")

    ingress_rules = optional(list(object({
      source      = string
      source_type = optional(string, "cidr_block")
      protocol    = string
      description = optional(string, "")
      stateless   = optional(bool, false)

      tcp_options = optional(object({
        destination_port_range = optional(object({
          min = number
          max = number
        }), null)
        source_port_range = optional(object({
          min = number
          max = number
        }), null)
      }), null)

      udp_options = optional(object({
        destination_port_range = optional(object({
          min = number
          max = number
        }), null)
        source_port_range = optional(object({
          min = number
          max = number
        }), null)
      }), null)

      icmp_options = optional(object({
        type = number
        code = optional(number, null)
      }), null)
    })), [])

    egress_rules = optional(list(object({
      destination      = string
      destination_type = optional(string, "cidr_block")
      protocol         = string
      description      = optional(string, "")
      stateless        = optional(bool, false)

      tcp_options = optional(object({
        destination_port_range = optional(object({
          min = number
          max = number
        }), null)
        source_port_range = optional(object({
          min = number
          max = number
        }), null)
      }), null)

      udp_options = optional(object({
        destination_port_range = optional(object({
          min = number
          max = number
        }), null)
        source_port_range = optional(object({
          min = number
          max = number
        }), null)
      }), null)

      icmp_options = optional(object({
        type = number
        code = optional(number, null)
      }), null)
    })), [])
  })
}
