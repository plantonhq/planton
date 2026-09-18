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
  description = "AwsNatGateway specification"
  type = object({
    # AWS region the NAT gateway is created in. Must match the region of the
    # subnet it lives in. Example: "us-west-2", "eu-west-1". This drives provider
    # construction, so it is required even though the gateway logically inherits
    # the subnet's region.
    region = string

    # How the gateway connects: "public" (default) or "private".
    #
    # - "public": the gateway is placed in a public subnet and assigned an Elastic
    #   IP (allocation_id is required); it provides outbound internet access for
    #   private subnets that route to it.
    # - "private": the gateway has no Elastic IP and provides outbound access only
    #   to other private networks (peered/transit/VPN), never the internet.
    #
    # AWS's own default for a new NAT gateway is public; this field is required so
    # the choice is always explicit (an empty value is rejected) and the two IaC
    # engines never need to invent a default. ForceNew: changing it replaces the
    # gateway.
    connectivity_type = optional(string, "")

    # The subnet the NAT gateway is created in (zonal mode only -- required
    # there, forbidden for a regional gateway, which spans the whole VPC).
    # Supply a literal subnet-id or reference an AwsSubnet and the platform
    # resolves its subnet_id output. For a public gateway this must be a PUBLIC
    # subnet (one whose route table reaches an internet gateway); for a private
    # gateway any subnet works. Immutable: changing the subnet replaces the
    # gateway.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_id = optional(string, "")

    # The Elastic IP allocation that gives a zonal public gateway its stable
    # outbound address. Supply a literal eipalloc-id or reference an
    # AwsElasticIp and the platform resolves its allocation_id output. Required
    # for a zonal public gateway; must be empty when connectivity_type is
    # private, and empty for a regional gateway (regional addressing lives in
    # availability_zone_addresses or is AWS-managed). ForceNew: changing it
    # replaces the gateway.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    allocation_id = optional(string, "")

    # The private IPv4 address to assign to a private gateway from the subnet's
    # range. Only valid when connectivity_type is private; leave empty to let AWS
    # choose. ForceNew: changing it replaces the gateway.
    private_ip = optional(string, "")

    # Additional Elastic IP allocations to attach to a public gateway, increasing
    # the number of available source ports for very high-throughput egress. Each
    # entry is a literal eipalloc-id or a reference to an AwsElasticIp. Only valid
    # when connectivity_type is public.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    secondary_allocation_ids = optional(list(string), [])

    # Additional private IPv4 addresses to assign to a private gateway, increasing
    # the number of available source ports. Only valid when connectivity_type is
    # private. Mutually exclusive with secondary_private_ip_address_count.
    secondary_private_ip_addresses = optional(list(string), [])

    # Number of additional private IPv4 addresses to let AWS assign to a private
    # gateway (an alternative to listing them explicitly). Only valid when
    # connectivity_type is private. Mutually exclusive with
    # secondary_private_ip_addresses.
    secondary_private_ip_address_count = optional(number, 0)

    # Placement model: "zonal" (default -- the classic single-subnet,
    # single-AZ gateway) or "regional" (one AWS-managed gateway spanning every
    # AZ of the VPC). Leave empty for zonal. ForceNew: changing the mode
    # replaces the gateway.
    availability_mode = optional(string, "")

    # The VPC a REGIONAL gateway spans (required for regional mode, forbidden
    # for zonal, where the subnet implies the VPC). Supply a literal vpc-id or
    # reference an AwsVpc and the platform resolves its vpc_id output.
    # Immutable: changing the VPC replaces the gateway.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = optional(string, "")

    # Explicit per-AZ layout for a regional gateway. Leave empty to let AWS
    # choose the zones and manage addresses automatically (auto mode); list
    # entries to pin the zones -- and, for a public regional gateway, the
    # Elastic IPs -- yourself (manual mode). Switching a regional gateway
    # between auto and manual replaces it (verified provider behavior).
    availability_zone_addresses = optional(list(object({
      # The availability zone, by name (e.g. "us-west-2a"). At least one of
      # availability_zone or availability_zone_id is required per entry; both
      # may be set (AWS reports both).
      availability_zone = optional(string, "")

      # The availability zone, by ID (e.g. "usw2-az1") -- stable across AWS
      # accounts, unlike zone names. At least one of availability_zone or
      # availability_zone_id is required per entry.
      availability_zone_id = optional(string, "")

      # Elastic IP allocations serving this zone (public regional gateways
      # only -- a private gateway carries no Elastic IPs). Each entry is a
      # literal eipalloc-id or a reference to an AwsElasticIp.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      allocation_ids = optional(list(string), [])
    })), [])
  })
}
