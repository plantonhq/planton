variable "metadata" {
  description = "metadata"
  type = object({

    # name of the resource
    name = string

    # id of the resource
    id = string

    # id of the organization to which the api-resource belongs to
    org = string

    # environment to which the resource belongs to
    env = string

    # labels for the resource
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # annotations for the resource
    annotations = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "spec"
  type = object({

    # vpc_id is the ID of the VPC where this Security Group will be created.
    # Example: "vpc-12345abcde"
    # This field is required because every Security Group must belong to one VPC.
    vpc_id = object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    })

    # description provides a short explanation of this Security Group’s purpose.
    # This field is required by AWS and cannot be modified once created without a replacement.
    # Example: "Allows inbound HTTP and SSH for web tier"
    description = string

    # ingress_rules define the inbound traffic rules for this Security Group.
    # If empty, inbound traffic is fully restricted (deny all).
    ingress = list(object({

      # protocol indicates the protocol for the rule.
      # Common values: "tcp", "udp", "icmp", or "-1" (all protocols).
      protocol = string

      # from_port is the starting port in the range.
      # For single-port rules, from_port == to_port.
      # Use 0 when specifying all ports (with protocol = -1) or for ICMP types.
      from_port = number

      # to_port is the ending port in the range.
      # For single-port rules, to_port == from_port.
      # Use 0 when specifying all ports (with protocol = -1) or for ICMP codes.
      to_port = number

      # ipv4_cidrs is the list of IPv4 CIDR blocks allowed (ingress) or targeted (egress).
      # Examples: "10.0.0.0/16", "0.0.0.0/0"
      # If empty, no IPv4 CIDRs are included in this rule.
      ipv4_cidrs = list(string)

      # ipv6_cidrs is the list of IPv6 CIDR blocks allowed or targeted.
      # Example: "::/0"
      # If empty, no IPv6 CIDRs are included in this rule.
      ipv6_cidrs = list(string)

      # source_security_group_ids is the list of Security Group IDs that can send traffic (for ingress) (StringValueOrRef).
      source_security_group_ids = list(object({ value = string }))

      # destination_security_group_ids is the list of Security Group IDs that receive traffic (for egress) (StringValueOrRef).
      destination_security_group_ids = list(object({ value = string }))

      # self_reference indicates whether to allow traffic from/to the same Security Group.
      # This is equivalent to referencing the group’s own ID.
      self_reference = bool

      # rule_description is an optional explanation of this specific rule,
      # aiding in clarity and maintenance. Max 255 chars recommended.
      description = string
    }))

    # egress_rules define the outbound traffic rules for this Security Group.
    # If empty, AWS defaults to allow all outbound traffic unless configured otherwise.
    egress = list(object({

      # protocol indicates the protocol for the rule.
      # Common values: "tcp", "udp", "icmp", or "-1" (all protocols).
      protocol = string

      # from_port is the starting port in the range.
      # For single-port rules, from_port == to_port.
      # Use 0 when specifying all ports (with protocol = -1) or for ICMP types.
      from_port = number

      # to_port is the ending port in the range.
      # For single-port rules, to_port == from_port.
      # Use 0 when specifying all ports (with protocol = -1) or for ICMP codes.
      to_port = number

      # ipv4_cidrs is the list of IPv4 CIDR blocks allowed (ingress) or targeted (egress).
      # Examples: "10.0.0.0/16", "0.0.0.0/0"
      # If empty, no IPv4 CIDRs are included in this rule.
      ipv4_cidrs = list(string)

      # ipv6_cidrs is the list of IPv6 CIDR blocks allowed or targeted.
      # Example: "::/0"
      # If empty, no IPv6 CIDRs are included in this rule.
      ipv6_cidrs = list(string)

      # source_security_group_ids is the list of Security Group IDs that can send traffic (for ingress) (StringValueOrRef).
      source_security_group_ids = list(object({ value = string }))

      # destination_security_group_ids is the list of Security Group IDs that receive traffic (for egress) (StringValueOrRef).
      destination_security_group_ids = list(object({ value = string }))

      # self_reference indicates whether to allow traffic from/to the same Security Group.
      # This is equivalent to referencing the group’s own ID.
      self_reference = bool

      # rule_description is an optional explanation of this specific rule,
      # aiding in clarity and maintenance. Max 255 chars recommended.
      description = string
    }))
  })
}