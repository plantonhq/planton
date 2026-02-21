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
  description = "OciVcn specification"
  type = object({
    compartment_id = object({
      value = string
    })

    cidr_blocks = list(string)

    display_name = optional(string, "")

    dns_label = optional(string, "")

    is_ipv6_enabled = optional(bool, false)

    is_internet_gateway_enabled = optional(bool, false)

    is_nat_gateway_enabled = optional(bool, false)

    is_service_gateway_enabled = optional(bool, false)
  })
}
