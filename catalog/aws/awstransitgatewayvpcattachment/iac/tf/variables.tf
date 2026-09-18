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
  description = "AwsTransitGatewayVpcAttachment specification"
  type = object({
    # The AWS region where the attachment will be created. Must match the
    # region of both the Transit Gateway and the VPC.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The Transit Gateway to attach to. Create-time immutable: changing it
    # replaces the attachment. Reference an AwsTransitGateway's
    # transit_gateway_id output or pass a literal ID.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    transit_gateway_id = string

    # The VPC to attach. Create-time immutable: changing it replaces the
    # attachment. Reference an AwsVpc's vpc_id output or pass a literal ID.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = string

    # Subnets in which AWS provisions the attachment's network interfaces --
    # at most one per Availability Zone, and all in the VPC being attached.
    # The gateway routes traffic only to/from AZs it has an ENI in, so cover
    # every AZ your workloads run in. Updatable in place (adding/removing an
    # AZ does not replace the attachment). Accepts direct subnet IDs or
    # references to AwsSubnet resources.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Enable DNS resolution for this attachment. When enabled (the default
    # when omitted), DNS queries from the attached VPC to public hostnames of
    # instances in other attached VPCs resolve to private IPs. Overrides the
    # gateway-level dial for this attachment only.
    dns_support = optional(bool)

    # Enable IPv6 traffic over this attachment. Requires the VPC and the
    # chosen subnets to carry IPv6 CIDRs. Disabled by default.
    ipv6_support = optional(bool, false)

    # Enable appliance mode for this attachment. Required when routing
    # traffic through a stateful virtual appliance (firewall, IDS/IPS) hosted
    # in the attached VPC: appliance mode keeps a flow's return traffic in
    # the same Availability Zone as the original flow, preserving symmetric
    # routing for stateful inspection. Only enable on the shared-services VPC
    # that hosts the appliances.
    appliance_mode_support = optional(bool, false)

    # Enable cross-VPC security group referencing for this attachment,
    # overriding the gateway-level dial. When left unset, the attachment
    # inherits the gateway's setting (AWS computes the effective value); set
    # it only to pin this attachment's posture explicitly.
    security_group_referencing_support = optional(bool)

    # Associate this attachment with the gateway's default route table. When
    # left unset, the gateway's own default-association dial decides (AWS
    # computes the effective value). Set to false when this attachment is
    # associated with a custom AwsTransitGatewayRouteTable instead -- an
    # attachment can be associated with at most ONE route table, so a default
    # association and a custom association conflict at the AWS API.
    default_route_table_association = optional(bool)

    # Propagate this attachment's VPC CIDRs into the gateway's default route
    # table. When left unset, the gateway's own default-propagation dial
    # decides (AWS computes the effective value). Set to false for isolated
    # routing domains where propagations are declared on custom
    # AwsTransitGatewayRouteTable resources (an attachment CAN propagate to
    # many tables, so this is about keeping the default table clean, not a
    # hard conflict).
    default_route_table_propagation = optional(bool)
  })
}
