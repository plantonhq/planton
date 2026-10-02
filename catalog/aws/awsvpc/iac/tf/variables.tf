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
  description = "AwsVpc specification"
  type = object({
    # The AWS region where the VPC will be created.
    # Example: "us-west-2", "eu-west-1", "ap-southeast-1".
    # For the list of regions, see:
    # https://aws.amazon.com/about-aws/global-infrastructure/regions_az/
    region = string

    # The primary IPv4 CIDR block for the VPC. This is the VPC's main address
    # range and is immutable: changing it replaces the VPC. The mask must be
    # between /16 and /28 (AWS limits), e.g. "10.0.0.0/16" yields 65,536
    # addresses. Leave empty only when allocating the primary CIDR from IPAM
    # (set ipv4_ipam_pool_id instead); exactly one of cidr_block or
    # ipv4_ipam_pool_id must be provided.
    cidr_block = optional(string, "")

    # Additional IPv4 address ranges to associate with the VPC beyond the
    # primary one. Each entry is associated as its own resource and can be
    # added or removed without recreating the VPC -- the standard way to grow
    # a network whose original range ran out of subnet space, or to carve a
    # distinct range (such as the 100.64.0.0/10 shared space) for specific
    # workloads. Each entry names an explicit CIDR or an IPAM allocation.
    secondary_ipv4_cidrs = optional(list(object({
      # An explicit IPv4 CIDR block to associate (e.g. "10.1.0.0/16" or
      # "100.64.0.0/16"). The mask must be between /16 and /28 and the range must
      # not overlap the VPC's other CIDRs. When ipam_pool_id is also set, this
      # pins a specific block from that pool. Immutable: changing it replaces the
      # association (the VPC itself is untouched).
      cidr_block = optional(string, "")

      # The IPAM pool to allocate this secondary CIDR from, instead of (or in
      # addition to, when pinning a specific block) cidr_block. Supply a literal
      # ipam-pool-id; there is no IPAM pool catalog kind yet. Immutable: changing
      # it replaces the association.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      ipam_pool_id = optional(string, "")

      # The netmask length of the CIDR to allocate from ipam_pool_id (e.g. 16 for
      # a /16). Must be between 16 and 28. Requires ipam_pool_id and is mutually
      # exclusive with cidr_block. Immutable: changing it replaces the
      # association.
      netmask_length = optional(number, 0)
    })), [])

    # The IPAM (IP Address Manager) pool to allocate the primary IPv4 CIDR from,
    # instead of specifying cidr_block directly. Pair with ipv4_netmask_length to
    # let IPAM choose a block of the requested size, or with cidr_block to take a
    # specific block from the pool. Supply a literal ipam-pool-id; there is no
    # IPAM pool catalog kind yet. Immutable: changing it replaces the VPC.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ipv4_ipam_pool_id = optional(string, "")

    # The netmask length of the primary IPv4 CIDR to allocate from
    # ipv4_ipam_pool_id (e.g. 16 for a /16). Must be between 16 and 28. Requires
    # ipv4_ipam_pool_id and is mutually exclusive with cidr_block. Immutable:
    # changing it replaces the VPC.
    ipv4_netmask_length = optional(number, 0)

    # The tenancy of instances launched into this VPC: "default" (instances may
    # run on shared hardware) or "dedicated" (every instance runs on
    # single-tenant hardware, at higher cost). Leave empty for the AWS default
    # ("default"). Note that AWS only supports changing "dedicated" -> "default"
    # in place; changing "default" -> "dedicated" replaces the VPC.
    instance_tenancy = optional(string, "")

    # Whether the Amazon-provided DNS server (the .2 resolver) answers DNS
    # queries from within the VPC. AWS enables this by default, and disabling it
    # breaks name resolution for most workloads, so when this field is unset the
    # VPC keeps DNS resolution ON. Set it explicitly to false only to deliberately
    # turn the Amazon resolver off.
    enable_dns_support = optional(bool)

    # Whether instances with public IP addresses receive public DNS hostnames.
    # AWS leaves this off by default; turn it on for VPCs whose instances should
    # be reachable by DNS name. See:
    # https://docs.aws.amazon.com/vpc/latest/userguide/vpc-dns.html#vpc-dns-hostnames
    enable_dns_hostnames = optional(bool, false)

    # Whether to enable Network Address Usage (NAU) metrics for the VPC. NAU is a
    # CloudWatch metric that tracks how the VPC's IP address space is consumed,
    # useful for capacity planning in large networks. Off by default.
    enable_network_address_usage_metrics = optional(bool, false)

    # Request an Amazon-provided IPv6 /56 CIDR block for the VPC. This is the
    # simplest way to make a VPC dual-stack. Mutually exclusive with the IPAM
    # IPv6 fields (ipv6_ipam_pool_id / ipv6_cidr_block / ipv6_netmask_length).
    assign_generated_ipv6_cidr_block = optional(bool, false)

    # An explicit IPv6 CIDR block to allocate from ipv6_ipam_pool_id (e.g.
    # "2600:1f18:abcd:1200::/56"). The prefix length must be one of /44, /48,
    # /52, /56, or /60 (the AWS VPC IPv6 sizes). Requires ipv6_ipam_pool_id (a
    # bring-your-own IPv6 block without IPAM is not supported on the VPC
    # resource) and is mutually exclusive with ipv6_netmask_length.
    ipv6_cidr_block = optional(string, "")

    # The network border group from which to advertise an Amazon-provided IPv6
    # CIDR (a Local Zone / Wavelength advertisement scope). Only valid together
    # with assign_generated_ipv6_cidr_block.
    ipv6_cidr_block_network_border_group = optional(string, "")

    # The IPAM pool to allocate the IPv6 CIDR from. Pair with ipv6_netmask_length
    # (the size to allocate) or ipv6_cidr_block (a specific block). Supply a
    # literal ipam-pool-id; there is no IPAM pool catalog kind yet. Mutually
    # exclusive with assign_generated_ipv6_cidr_block.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ipv6_ipam_pool_id = optional(string, "")

    # The netmask length of the IPv6 CIDR to allocate from ipv6_ipam_pool_id.
    # Must be one of 44, 48, 52, 56, or 60. Requires ipv6_ipam_pool_id and is
    # mutually exclusive with ipv6_cidr_block.
    ipv6_netmask_length = optional(number, 0)

    # Additional IPv6 address ranges to associate with the VPC beyond the one
    # configured above. Each entry is associated as its own resource and can be
    # added or removed without recreating the VPC. Each entry names exactly one
    # source: an Amazon-provided block, a BYOIP public pool, or an IPAM pool.
    secondary_ipv6_cidrs = optional(list(object({
      # Request an Amazon-provided IPv6 /56 block. Mutually exclusive with the
      # pool sources and with cidr_block (Amazon chooses the range).
      assign_generated = optional(bool, false)

      # The BYOIP public IPv6 pool (ipv6pool-ec2-...) to take the range from --
      # address space you brought to AWS outside of IPAM. Pair with cidr_block to
      # pin the specific block. Immutable: changing it replaces the association.
      ipv6_pool = optional(string, "")

      # The IPAM pool to allocate the range from. Pair with netmask_length (the
      # size to allocate) or cidr_block (a specific block). Supply a literal
      # ipam-pool-id; there is no IPAM pool catalog kind yet. Immutable: changing
      # it replaces the association.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      ipam_pool_id = optional(string, "")

      # An explicit IPv6 CIDR block, used with ipv6_pool or ipam_pool_id to pin
      # the exact range (e.g. "2600:1f18:abcd:1200::/56"). The prefix length must
      # be one of /44, /48, /52, /56, or /60. Mutually exclusive with
      # netmask_length and with assign_generated. Immutable: changing it replaces
      # the association.
      cidr_block = optional(string, "")

      # The netmask length to allocate from ipam_pool_id. Must be one of 44, 48,
      # 52, 56, or 60 (AWS only sizes IPAM allocations -- pool and Amazon blocks
      # pin their own size). Mutually exclusive with cidr_block. Immutable:
      # changing it replaces the association.
      netmask_length = optional(number, 0)
    })), [])

    # VPC Encryption Control: monitor or enforce encryption in transit for
    # traffic entering, leaving, and moving within the VPC. Omit to leave the
    # feature off (the AWS default). Start in "monitor" mode to observe, then
    # move to "enforce" once the exclusions below cover every path that cannot
    # encrypt.
    encryption_control = optional(object({
      # The control mode: "monitor" (observe and report unencrypted paths) or
      # "enforce" (block unencrypted traffic, honoring the exclusions below).
      mode = string

      # Exclude internet gateway traffic from enforcement.
      exclude_internet_gateway = optional(bool, false)

      # Exclude egress-only internet gateway traffic from enforcement.
      exclude_egress_only_internet_gateway = optional(bool, false)

      # Exclude NAT gateway traffic from enforcement.
      exclude_nat_gateway = optional(bool, false)

      # Exclude virtual private gateway (Site-to-Site VPN) traffic from
      # enforcement.
      exclude_virtual_private_gateway = optional(bool, false)

      # Exclude VPC peering traffic from enforcement.
      exclude_vpc_peering = optional(bool, false)

      # Exclude VPC Lattice traffic from enforcement.
      exclude_vpc_lattice = optional(bool, false)

      # Exclude Lambda function traffic from enforcement.
      exclude_lambda = optional(bool, false)

      # Exclude Elastic File System (EFS) traffic from enforcement.
      exclude_elastic_file_system = optional(bool, false)
    }))
  })
}
