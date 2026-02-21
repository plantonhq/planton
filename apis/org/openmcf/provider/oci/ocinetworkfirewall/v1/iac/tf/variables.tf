variable "metadata" {
  type = object({
    name   = string
    id     = optional(string, "")
    org    = optional(string, "")
    env    = optional(string, "")
    labels = optional(map(string), {})
  })
}

variable "spec" {
  type = object({
    compartment_id = object({
      value = string
    })
    subnet_id = object({
      value = string
    })
    display_name        = optional(string, "")
    ipv4_address        = optional(string, "")
    ipv6_address        = optional(string, "")
    availability_domain = optional(string, "")
    network_security_group_ids = optional(list(object({
      value = string
    })), [])
    nat_configuration = optional(object({
      must_enable_private_nat = bool
    }))
    shape = optional(string, "")

    policy = object({
      display_name = optional(string, "")
      description  = optional(string, "")

      address_lists = optional(list(object({
        name        = string
        type        = string
        addresses   = list(string)
        description = optional(string, "")
      })), [])

      services = optional(list(object({
        name = string
        type = string
        port_ranges = list(object({
          minimum_port = number
          maximum_port = optional(number)
        }))
        description = optional(string, "")
      })), [])

      service_lists = optional(list(object({
        name        = string
        services    = list(string)
        description = optional(string, "")
      })), [])

      url_lists = optional(list(object({
        name = string
        urls = list(object({
          pattern = string
        }))
        description = optional(string, "")
      })), [])

      security_rules = optional(list(object({
        name       = string
        action     = string
        inspection = optional(string, "")
        condition = object({
          source_addresses      = optional(list(string), [])
          destination_addresses = optional(list(string), [])
          services              = optional(list(string), [])
          urls                  = optional(list(string), [])
        })
        description = optional(string, "")
      })), [])
    })
  })
}
