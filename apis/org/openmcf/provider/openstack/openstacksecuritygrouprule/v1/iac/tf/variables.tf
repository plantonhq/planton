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
  description = "OpenStackSecurityGroupRuleSpec defines a standalone security group rule"
  type = object({
    # (Required) The ID of the security group this rule belongs to.
    # Supports StringValueOrRef pattern - use {value: "sg-id"} for literal values.
    security_group_id = object({
      value = string
    })

    # (Required) Direction of the rule: "ingress" or "egress".
    direction = string

    # (Required) Layer-3 protocol type: "IPv4" or "IPv6".
    ethertype = string

    # (Optional) IP protocol: "tcp", "udp", "icmp", etc. Empty = all protocols.
    protocol = optional(string, "")

    # (Optional) Lower port bound (or ICMP type).
    port_range_min = optional(number)

    # (Optional) Upper port bound (or ICMP code).
    port_range_max = optional(number)

    # (Optional) CIDR to restrict traffic. Mutually exclusive with remote_group_id.
    remote_ip_prefix = optional(string, "")

    # (Optional) Security group ID for remote group filtering.
    # Supports StringValueOrRef pattern - use {value: "sg-id"} for literal values.
    remote_group_id = optional(object({
      value = string
    }))

    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
