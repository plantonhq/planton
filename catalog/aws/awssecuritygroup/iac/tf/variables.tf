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
  description = "AwsSecurityGroup specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # vpc_id is the ID of the VPC where this Security Group will be created.
    # Example: "vpc-12345abcde"
    # Required: every security group belongs to exactly one VPC.
    # ForceNew: changing the VPC forces group replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = string

    # description provides a short explanation of this Security Group's purpose.
    # Required by AWS. ForceNew: AWS does not allow editing a group description
    # in place, so changing it forces group replacement.
    # Example: "Allows inbound HTTP and SSH for web tier"
    description = string

    # ingress defines the inbound traffic rules for this Security Group.
    # If empty, inbound traffic is fully restricted (deny all).
    ingress = optional(list(object({
      # protocol indicates the protocol for the rule.
      # Common values: "tcp", "udp", "icmp", "icmpv6", or "-1" (all protocols).
      # IANA protocol numbers are also accepted.
      protocol = string

      # from_port is the starting port in the range. For single-port rules,
      # from_port == to_port. For ICMP/ICMPv6, from_port is the ICMP TYPE
      # (-1 means all types). For all-protocol rules (protocol "-1"), both ports
      # must be 0.
      from_port = optional(number, 0)

      # to_port is the ending port in the range. For single-port rules,
      # to_port == from_port. For ICMP/ICMPv6, to_port is the ICMP CODE
      # (-1 means all codes). For all-protocol rules (protocol "-1"), both ports
      # must be 0.
      to_port = optional(number, 0)

      # ipv4_cidrs is the list of IPv4 CIDR blocks allowed (ingress) or targeted (egress).
      # Examples: "10.0.0.0/16", "0.0.0.0/0"
      # If empty, no IPv4 CIDRs are included in this rule.
      ipv4_cidrs = optional(list(string), [])

      # ipv6_cidrs is the list of IPv6 CIDR blocks allowed or targeted.
      # Example: "::/0"
      # If empty, no IPv6 CIDRs are included in this rule.
      ipv6_cidrs = optional(list(string), [])

      # source_security_group_ids is the list of Security Group IDs that can send traffic (for ingress).
      # Typically used for internal traffic between resources. Can reference other AwsSecurityGroup resources.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_security_group_ids = optional(list(string), [])

      # destination_security_group_ids is the list of Security Group IDs that receive traffic (for egress).
      # Useful for restricting outbound traffic to specific groups. Can reference other AwsSecurityGroup resources.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination_security_group_ids = optional(list(string), [])

      # prefix_list_ids is the list of managed prefix list IDs allowed (ingress)
      # or targeted (egress). Example: "pl-63a5400a". Managed prefix lists name a
      # set of CIDRs by a stable ID: AWS-managed lists cover services like S3 and
      # DynamoDB gateway endpoints (so an egress rule can target "the S3 service"
      # instead of hardcoding its CIDRs), and customer-managed lists let network
      # teams maintain shared CIDR sets (office ranges, partner networks) that
      # many groups reference without copying.
      prefix_list_ids = optional(list(string), [])

      # self_reference indicates whether to allow traffic from/to the same Security Group.
      # This is equivalent to referencing the group's own ID -- the standard pattern
      # for intra-cluster traffic (nodes of one cluster talking to each other).
      self_reference = optional(bool, false)

      # description is an optional explanation of this specific rule,
      # aiding in clarity and maintenance. Max 255 chars.
      description = optional(string, "")
    })), [])

    # egress defines the outbound traffic rules for this Security Group.
    # If empty, ALL outbound traffic is denied: the module revokes the allow-all
    # egress rule AWS adds to every new group, so the manifest is the complete
    # statement of what the group permits. Add an explicit all-traffic egress
    # rule (protocol "-1", 0.0.0.0/0) to restore the AWS default behavior.
    egress = optional(list(object({
      # protocol indicates the protocol for the rule.
      # Common values: "tcp", "udp", "icmp", "icmpv6", or "-1" (all protocols).
      # IANA protocol numbers are also accepted.
      protocol = string

      # from_port is the starting port in the range. For single-port rules,
      # from_port == to_port. For ICMP/ICMPv6, from_port is the ICMP TYPE
      # (-1 means all types). For all-protocol rules (protocol "-1"), both ports
      # must be 0.
      from_port = optional(number, 0)

      # to_port is the ending port in the range. For single-port rules,
      # to_port == from_port. For ICMP/ICMPv6, to_port is the ICMP CODE
      # (-1 means all codes). For all-protocol rules (protocol "-1"), both ports
      # must be 0.
      to_port = optional(number, 0)

      # ipv4_cidrs is the list of IPv4 CIDR blocks allowed (ingress) or targeted (egress).
      # Examples: "10.0.0.0/16", "0.0.0.0/0"
      # If empty, no IPv4 CIDRs are included in this rule.
      ipv4_cidrs = optional(list(string), [])

      # ipv6_cidrs is the list of IPv6 CIDR blocks allowed or targeted.
      # Example: "::/0"
      # If empty, no IPv6 CIDRs are included in this rule.
      ipv6_cidrs = optional(list(string), [])

      # source_security_group_ids is the list of Security Group IDs that can send traffic (for ingress).
      # Typically used for internal traffic between resources. Can reference other AwsSecurityGroup resources.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_security_group_ids = optional(list(string), [])

      # destination_security_group_ids is the list of Security Group IDs that receive traffic (for egress).
      # Useful for restricting outbound traffic to specific groups. Can reference other AwsSecurityGroup resources.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination_security_group_ids = optional(list(string), [])

      # prefix_list_ids is the list of managed prefix list IDs allowed (ingress)
      # or targeted (egress). Example: "pl-63a5400a". Managed prefix lists name a
      # set of CIDRs by a stable ID: AWS-managed lists cover services like S3 and
      # DynamoDB gateway endpoints (so an egress rule can target "the S3 service"
      # instead of hardcoding its CIDRs), and customer-managed lists let network
      # teams maintain shared CIDR sets (office ranges, partner networks) that
      # many groups reference without copying.
      prefix_list_ids = optional(list(string), [])

      # self_reference indicates whether to allow traffic from/to the same Security Group.
      # This is equivalent to referencing the group's own ID -- the standard pattern
      # for intra-cluster traffic (nodes of one cluster talking to each other).
      self_reference = optional(bool, false)

      # description is an optional explanation of this specific rule,
      # aiding in clarity and maintenance. Max 255 chars.
      description = optional(string, "")
    })), [])

    # revoke_rules_on_delete forcibly revokes this group's rules (and rules in
    # OTHER groups that reference this one) before deleting the group. Without
    # it, deleting a group that is still referenced by another group's rules
    # fails with a DependencyViolation. Enable for groups that are referenced
    # cross-group (e.g. an app tier referenced by a database tier) so teardown
    # never requires manual rule surgery. Safe to toggle in place.
    revoke_rules_on_delete = optional(bool, false)

    # additional_vpc_ids shares this security group into OTHER VPCs beyond its
    # home vpc_id, so resources in those VPCs can attach the same group instead
    # of maintaining copied groups per VPC (one firewall definition, many VPCs).
    # Each entry must be a different VPC in the same account and region as the
    # group's own VPC, and must not repeat vpc_id or another entry — AWS rejects
    # duplicates at apply (entries may be references, so this is not validated
    # here). Reference AwsVpc vpc_id outputs or pass literal "vpc-..." ids.
    # Adding or removing an entry associates/disassociates in place; the group
    # itself is never replaced. Rules referencing security groups from a
    # DIFFERENT VPC than the one a packet traverses are ignored by AWS in that
    # VPC — prefer CIDR/prefix-list rules on multi-VPC groups.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    additional_vpc_ids = optional(list(string), [])
  })
}
