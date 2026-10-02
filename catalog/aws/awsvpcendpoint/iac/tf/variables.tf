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
  description = "AwsVpcEndpoint specification"
  type = object({
    # The AWS region the endpoint is created in. Must match the VPC's
    # region. Example: "us-west-2", "eu-west-1".
    region = string

    # The VPC the endpoint lives in. Reference an AwsVpc's vpc_id output
    # or pass a literal VPC id. Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = string

    # The endpoint type. "Gateway" (S3/DynamoDB via route tables),
    # "Interface" (ENI-based PrivateLink -- most AWS services and every
    # third-party PrivateLink service), "GatewayLoadBalancer" (fronts a
    # GWLB appliance fleet), "Resource" and "ServiceNetwork" (VPC
    # Lattice). Empty defaults to "Gateway" -- AWS's own default.
    # Create-only in AWS: changing the type replaces the endpoint.
    endpoint_type = optional(string, "")

    # The AWS service to connect to, e.g. "com.amazonaws.us-west-2.s3"
    # for S3 in us-west-2 or a PrivateLink provider's
    # "com.amazonaws.vpce.<region>.vpce-svc-..." name. Exactly one of
    # service_name, resource_configuration_arn, or service_network_arn
    # must be set. Create-only in AWS.
    service_name = optional(string, "")

    # The ARN of a VPC Lattice resource configuration to connect to --
    # the "Resource" endpoint type's target. Exactly one of the three
    # service-target fields must be set. Create-only in AWS.
    resource_configuration_arn = optional(string, "")

    # The ARN of a VPC Lattice service network to connect to -- the
    # "ServiceNetwork" endpoint type's target. Exactly one of the three
    # service-target fields must be set. Create-only in AWS.
    service_network_arn = optional(string, "")

    # Route tables a GATEWAY endpoint injects its prefix-list route into.
    # Traffic to the service from any subnet on these tables flows
    # through the endpoint. Reference an AwsSubnet's route_table_id
    # output when the subnet owns its table (inline routes), or the
    # AwsVpc's main_route_table_id / default_route_table_id outputs when
    # subnets ride the VPC main table; literals also work. Gateway
    # endpoints only. Updates in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    route_table_ids = optional(list(string), [])

    # Subnets an INTERFACE / GatewayLoadBalancer / Resource /
    # ServiceNetwork endpoint places its network interfaces in -- one ENI
    # per subnet, so spread across AZs for availability (each AZ is
    # billed separately for interface endpoints). Reference AwsSubnet
    # subnet_id outputs or pass literal subnet ids. Not applicable to
    # gateway endpoints. Updates in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Security groups attached to an INTERFACE endpoint's network
    # interfaces -- they must allow inbound traffic from the clients on
    # the service's port (443 for AWS APIs). Empty means AWS attaches
    # the VPC's DEFAULT security group. Reference AwsSecurityGroup
    # security_group_id outputs or pass literal ids. Interface endpoints
    # only. Updates in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Resolve the service's PUBLIC DNS name (e.g. sts.us-west-2.
    # amazonaws.com) to the endpoint's private IPs inside the VPC, via an
    # AWS-managed private hosted zone. Most interface-endpoint users want
    # this on -- clients keep their default SDK endpoints and privately
    # reach the service with zero code changes. Requires the VPC to have
    # BOTH DNS support and DNS hostnames enabled. Interface endpoints
    # only (AWS replaces the endpoint if this changes on other types --
    # on Interface it updates in place). Tri-state: unset lets AWS use
    # its default (off) and keeps an existing endpoint's current setting
    # (the provider attribute is Optional+Computed -- an omitted value is
    # never sent); true enables private DNS; an EXPLICIT false is the
    # only way to turn it back off once enabled.
    private_dns_enabled = optional(bool)

    # Fine-grained DNS behavior for the endpoint. Only meaningful when
    # the endpoint creates DNS records (interface endpoints, and the
    # Lattice types for the preference/domain fields).
    dns_options = optional(object({
      # Which record types the endpoint's DNS names resolve to: "ipv4",
      # "ipv6", "dualstack", or "service-defined" (the service picks).
      # Empty lets AWS choose based on the endpoint's ip_address_type.
      dns_record_ip_type = optional(string, "")

      # Route only INBOUND (on-premises / cross-VPC) resolver traffic
      # through this endpoint's private DNS, keeping in-VPC traffic on the
      # gateway endpoint. Applies to the S3 dual-stack pattern -- a service
      # with BOTH a gateway and an interface endpoint in the same VPC:
      # in-VPC S3 traffic rides the free gateway while on-premises clients
      # resolve to the interface endpoint. Requires private_dns_enabled.
      private_dns_only_for_inbound_resolver_endpoint = optional(bool, false)

      # Which private domains get a private hosted zone, for Resource /
      # ServiceNetwork endpoints: "ALL_DOMAINS", "VERIFIED_DOMAINS_ONLY",
      # "VERIFIED_DOMAINS_AND_SPECIFIED_DOMAINS", or
      # "SPECIFIED_DOMAINS_ONLY". Empty keeps the AWS default. Create-only
      # in AWS.
      private_dns_preference = optional(string, "")

      # The private domains to create hosted zones for -- required when
      # private_dns_preference includes specified domains
      # ("VERIFIED_DOMAINS_AND_SPECIFIED_DOMAINS" or
      # "SPECIFIED_DOMAINS_ONLY"), and must be empty otherwise. AWS allows
      # 1-10 domains. Create-only in AWS.
      private_dns_specified_domains = optional(list(string), [])
    }))

    # The IP address type of the endpoint: "ipv4", "dualstack", or
    # "ipv6". Empty lets AWS pick based on the service and subnets --
    # effectively ipv4 for nearly every service today. The service must
    # support the chosen type. Updates in place.
    ip_address_type = optional(string, "")

    # An IAM policy document controlling which principals may use the
    # endpoint to reach which resources -- e.g. "only this account's
    # buckets" on an S3 gateway endpoint, turning the endpoint into a
    # data-exfiltration control. Expressed as a structured document
    # (standard IAM policy shape: Version, Statement, ...) rather than an
    # embedded JSON string. Empty means full access (AWS's default). All
    # gateway and most interface endpoints support policies. Updates in
    # place.
    policy = optional(any)

    # Pin specific IPv4/IPv6 addresses for the endpoint's ENI in chosen
    # subnets -- for appliances or firewall rules that need stable
    # endpoint IPs. Every subnet_id listed here must also appear in
    # subnet_ids. Rarely needed: AWS assigns addresses automatically when
    # omitted. Updates in place.
    subnet_configurations = optional(list(object({
      # The subnet whose ENI addresses are being pinned. Reference an
      # AwsSubnet's subnet_id output or pass a literal subnet id -- and
      # list the same subnet in spec.subnet_ids.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_id = string

      # The private IPv4 address to assign to the ENI in this subnet. Must
      # be a free address inside the subnet's CIDR.
      ipv4 = optional(string, "")

      # The IPv6 address to assign to the ENI in this subnet. Requires a
      # dualstack or ipv6 endpoint and an IPv6-enabled subnet.
      ipv6 = optional(string, "")
    })), [])

    # Connect to a service in ANOTHER region (interface endpoints only),
    # e.g. reach us-east-1-only services from a us-west-2 VPC without
    # cross-region networking of your own. Empty means the endpoint's own
    # region. Create-only in AWS.
    service_region = optional(string, "")

    # Accept the endpoint connection automatically when the PrivateLink
    # service requires acceptance and lives in the SAME AWS account.
    # Cross-account services must accept on their side regardless.
    auto_accept = optional(bool, false)
  })
}
