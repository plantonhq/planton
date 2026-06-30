variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "OpenStackSecurityGroupSpec defines the configuration for a Neutron security group"
  type = object({
    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Delete default egress rules after creation.
    delete_default_rules = optional(bool)

    # (Optional) Stateful or stateless security group.
    stateful = optional(bool)

    # (Optional) Inline security group rules.
    rules = optional(list(object({
      key              = string
      direction        = string
      ethertype        = string
      protocol         = optional(string, "")
      port_range_min   = optional(number)
      port_range_max   = optional(number)
      remote_ip_prefix = optional(string, "")
      remote_group_id  = optional(string, "")
      description      = optional(string, "")
    })), [])

    # (Optional) Tags to associate with the security group.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
