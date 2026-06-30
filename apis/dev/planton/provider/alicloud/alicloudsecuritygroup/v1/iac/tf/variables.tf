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
  description = "Alibaba Cloud Security Group specification"
  type = object({
    region                = string
    vpc_id                = string
    security_group_name   = string
    description           = optional(string, "")
    inner_access_policy   = optional(string, "Accept")
    resource_group_id     = optional(string, "")
    tags                  = optional(map(string), {})
    rules = optional(list(object({
      type                      = string
      ip_protocol               = string
      port_range                = optional(string, "-1/-1")
      cidr_ip                   = optional(string, "")
      source_security_group_id  = optional(string, "")
      priority                  = optional(number, 1)
      policy                    = optional(string, "accept")
      description               = optional(string, "")
    })), [])
  })

  validation {
    condition     = length(var.spec.security_group_name) >= 2 && length(var.spec.security_group_name) <= 128
    error_message = "security_group_name must be between 2 and 128 characters."
  }

  validation {
    condition     = contains(["Accept", "Drop"], var.spec.inner_access_policy)
    error_message = "inner_access_policy must be one of: Accept, Drop."
  }

  validation {
    condition = alltrue([
      for r in var.spec.rules : contains(["ingress", "egress"], r.type)
    ])
    error_message = "Each rule type must be one of: ingress, egress."
  }

  validation {
    condition = alltrue([
      for r in var.spec.rules : contains(["tcp", "udp", "icmp", "gre", "all"], r.ip_protocol)
    ])
    error_message = "Each rule ip_protocol must be one of: tcp, udp, icmp, gre, all."
  }

  validation {
    condition = alltrue([
      for r in var.spec.rules : r.priority >= 1 && r.priority <= 100
    ])
    error_message = "Each rule priority must be between 1 and 100."
  }

  validation {
    condition = alltrue([
      for r in var.spec.rules : contains(["accept", "drop"], r.policy)
    ])
    error_message = "Each rule policy must be one of: accept, drop."
  }
}
