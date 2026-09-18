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
  description = "AwsSubnet specification"
  type = object({
    # AWS region the subnet is created in. Must match the region of the parent
    # VPC (a subnet cannot span regions). Example: "us-west-2", "eu-west-1".
    # This drives provider construction, so it is required even though the subnet
    # logically inherits the VPC's region.
    region = string

    # The VPC this subnet belongs to. Supply a literal vpc-id or reference an
    # AwsVpc and the platform resolves its vpc_id output. Immutable: changing the
    # VPC replaces the subnet.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = string

    # The availability zone the subnet lives in, by NAME (e.g. "us-west-2a").
    # AWS subnets are single-AZ; to span AZs, create one AwsSubnet per zone.
    # Exactly one of availability_zone or availability_zone_id must be set --
    # explicit placement is mandatory (AWS would otherwise pick a zone at
    # random, which is never what declarative infrastructure wants). Immutable:
    # changing the AZ replaces the subnet.
    availability_zone = optional(string, "")

    # IPv4 CIDR block for the subnet (e.g. "10.0.1.0/24"). Must fall within the
    # VPC's CIDR and not overlap any sibling subnet. Note that AWS reserves the
    # first four and the last IP address in every subnet, so a /28 yields 11
    # usable addresses. Exactly one IPv4 addressing method is required unless
    # ipv6_native is true: either this CIDR or an IPAM allocation
    # (ipv4_ipam_pool_id + ipv4_netmask_length). Immutable: changing the CIDR
    # replaces the subnet.
    cidr_block = optional(string, "")

    # When true, instances launched into this subnet receive a public IPv4
    # address by default. This is a convenience for public subnets; it does NOT
    # by itself make the subnet routable to the internet (that requires a route
    # to an internet gateway). Defaults to false (the AWS default).
    map_public_ip_on_launch = optional(bool, false)

    # When true, network interfaces created in this subnet are assigned an IPv6
    # address from the subnet's IPv6 CIDR on creation. Only meaningful when
    # ipv6_cidr_block is set. Defaults to false (the AWS default).
    assign_ipv6_address_on_creation = optional(bool, false)

    # IPv6 CIDR block for a dual-stack subnet (e.g. "2600:1f18:abcd:1200::/64").
    # Must be a /64 carved from an IPv6 CIDR associated with the parent VPC.
    # Leave empty for an IPv4-only subnet.
    ipv6_cidr_block = optional(string, "")

    # When true, enables DNS64 on the subnet so that instances can reach
    # IPv4-only destinations from an IPv6-only subnet via NAT64. Requires the
    # subnet to have an IPv6 CIDR. Defaults to false (the AWS default).
    enable_dns64 = optional(bool, false)

    # When true, instances launched into this subnet get a DNS A record for their
    # resource name. Used with private_dns_hostname_type_on_launch. Defaults to
    # false (the AWS default).
    enable_resource_name_dns_a_record_on_launch = optional(bool, false)

    # When true, instances launched into this subnet get a DNS AAAA record for
    # their resource name (IPv6). Defaults to false (the AWS default).
    enable_resource_name_dns_aaaa_record_on_launch = optional(bool, false)

    # The type of hostname assigned to instances at launch. One of "ip-name"
    # (hostname derived from the private IPv4 address) or "resource-name"
    # (hostname derived from the instance's resource id, required for IPv6-only
    # subnets). Leave empty to use the AWS default for the subnet's address
    # family.
    private_dns_hostname_type_on_launch = optional(string)

    # An existing route table to associate with this subnet. Mutually exclusive
    # with routes. If neither is set, the VPC's main route table is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    route_table_id = optional(string, "")

    # Inline route rules. When present, a dedicated route table is created, owned
    # by this subnet, populated with these rules, and associated with the subnet.
    # Mutually exclusive with route_table_id.
    routes = optional(list(object({
      # Destination IPv4 CIDR (e.g. "0.0.0.0/0" for the default route).
      destination_cidr_block = optional(string, "")

      # Destination IPv6 CIDR (e.g. "::/0" for the default IPv6 route).
      destination_ipv6_cidr_block = optional(string, "")

      # Destination managed-prefix-list id (e.g. "pl-0123abcd"), for routing to a
      # curated set of CIDRs such as an AWS service prefix list.
      destination_prefix_list_id = optional(string, "")

      # The kind of network entity this route targets. Determines which AWS route
      # attribute target_id maps to (gateway_id, nat_gateway_id, etc.).
      target_type = optional(string, "")

      # Identifier of the target. Supply a literal id, or reference the resource
      # that produces it (e.g. an AwsInternetGateway's internet_gateway_id or an
      # AwsNatGateway's nat_gateway_id). This single field is intentionally
      # polymorphic across all of target_type's kinds, so it carries no
      # default_kind -- the target's kind is given by target_type, and the
      # producing resource is referenced explicitly via value_from.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_id = string
    })), [])

    # The availability zone the subnet lives in, by ID (e.g. "usw2-az1"). AZ IDs
    # are stable across AWS accounts, unlike AZ names which are shuffled
    # per-account -- use the ID form when coordinating subnet placement across
    # accounts. Exactly one of availability_zone or availability_zone_id must be
    # set. Immutable: changing it replaces the subnet.
    availability_zone_id = optional(string, "")

    # The IPAM pool to allocate the subnet's IPv4 CIDR from, as an alternative
    # to declaring cidr_block explicitly. Requires ipv4_netmask_length. Supply a
    # literal ipam-pool-id; there is no IPAM pool catalog kind yet. Immutable:
    # changing it replaces the subnet.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ipv4_ipam_pool_id = optional(string, "")

    # Netmask length for the IPv4 CIDR allocated from ipv4_ipam_pool_id
    # (e.g. 24 for a /24). Valid range 16-28 (the AWS subnet netmask bounds).
    # Requires ipv4_ipam_pool_id; mutually exclusive with cidr_block.
    ipv4_netmask_length = optional(number)

    # The IPAM pool to allocate the subnet's IPv6 CIDR from, as an alternative
    # to declaring ipv6_cidr_block explicitly. Requires ipv6_netmask_length.
    # Supply a literal ipam-pool-id; there is no IPAM pool catalog kind yet.
    # Immutable: changing it replaces the subnet.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ipv6_ipam_pool_id = optional(string, "")

    # Netmask length for the IPv6 CIDR allocated from ipv6_ipam_pool_id. AWS
    # subnets accept /44, /48, /52, /56, /60, or /64 (a /64 is the norm --
    # anything shorter is for subnets that will be further subdivided by
    # longest-prefix-match routing). Requires ipv6_ipam_pool_id; mutually
    # exclusive with ipv6_cidr_block.
    ipv6_netmask_length = optional(number)

    # When true, creates an IPv6-only subnet: no IPv4 addressing at all
    # (cidr_block and ipv4_ipam_pool_id must both be empty), and an IPv6 CIDR
    # (ipv6_cidr_block or the IPv6 IPAM pair) is required. Instances in an
    # IPv6-only subnet need private_dns_hostname_type_on_launch =
    # "resource-name" and can reach IPv4-only destinations via DNS64 + NAT64
    # (see enable_dns64). Immutable: changing it replaces the subnet.
    ipv6_native = optional(bool, false)

    # Virtual private gateway IDs (vgw-...) whose routes propagate into the
    # subnet-owned route table -- the mechanism that pulls Site-to-Site VPN /
    # Direct Connect routes in automatically instead of declaring them one by
    # one. Only valid when this subnet OWNS its route table (inline routes
    # mode, or propagation-only with no inline routes); not allowed with an
    # external route_table_id, whose owner controls its own propagation.
    propagating_vgws = optional(list(string), [])
  })
}
