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
  description = "AwsTransitGateway specification"
  type = object({
    # The AWS region where the Transit Gateway will be created. All attached
    # VPCs must be in this region (cross-region connectivity uses TGW peering,
    # a separate surface).
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Human-readable description for the Transit Gateway. Appears in the AWS
    # console and CLI output.
    description = optional(string, "")

    # Private Autonomous System Number (ASN) for the Amazon side of BGP
    # sessions. Used when connecting VPNs or Direct Connect gateways; pick a
    # number that does not collide with your on-premises ASNs. Changing it
    # after creation replaces the gateway.
    #
    # Valid ranges: 64512-65534 (16-bit private) or 4200000000-4294967294
    # (32-bit private). When omitted, AWS assigns the default 64512.
    amazon_side_asn = optional(number, 0)

    # Automatically associate new attachments with the gateway's default route
    # table. When enabled (the default when omitted), every new attachment is
    # associated without manual intervention -- the full-mesh starting point.
    #
    # Disable for segmented topologies where every attachment is associated
    # explicitly with an AwsTransitGatewayRouteTable. AWS QUIRK: flipping this
    # dial from disabled back to enabled REPLACES the gateway (and with it,
    # every attachment); disabling an enabled gateway updates in place.
    default_route_table_association = optional(bool)

    # Automatically propagate routes from new attachments to the default route
    # table. When enabled (the default when omitted), every attached VPC's
    # CIDR blocks are advertised into the default table, creating full-mesh
    # reachability.
    #
    # Disable for isolated routing domains where propagations are declared
    # per-route-table on AwsTransitGatewayRouteTable resources. AWS QUIRK:
    # like the association dial, flipping from disabled back to enabled
    # REPLACES the gateway; disabling updates in place.
    default_route_table_propagation = optional(bool)

    # Enable DNS resolution for instances in attached VPCs. When enabled (the
    # default when omitted), queries to public DNS hostnames of instances in
    # other attached VPCs resolve to their private IP addresses.
    dns_support = optional(bool)

    # Enable Equal Cost Multi-Path (ECMP) routing for VPN connections. When
    # enabled (the default when omitted) and multiple VPN tunnels advertise
    # the same routes, traffic is distributed across all tunnels for higher
    # aggregate throughput.
    vpn_ecmp_support = optional(bool)

    # Automatically accept cross-account attachment requests shared via AWS
    # Resource Access Manager (RAM). Disabled by default: shared attachments
    # require explicit acceptance, which is the safer posture for a hub that
    # other accounts can request to join.
    auto_accept_shared_attachments = optional(bool, false)

    # Enable cross-VPC security group referencing. When enabled, security
    # group rules in one attached VPC can reference security groups in
    # another VPC connected through this gateway, replacing broad CIDR-based
    # rules with precise group-to-group rules. Individual attachments can
    # override this dial.
    security_group_referencing_support = optional(bool, false)

    # Enable multicast traffic routing through the Transit Gateway. This is
    # create-time immutable: changing it replaces the entire gateway. Only
    # enable for a clear multicast use case (financial market data feeds,
    # media streaming); the multicast domain surface itself is a separate
    # resource family.
    multicast_support = optional(bool, false)

    # Enforce encryption of in-transit traffic through the Transit Gateway.
    # When left unset, AWS applies its own default and the effective value is
    # computed after creation. Only set this when you explicitly need to pin
    # the encryption posture -- the tri-state (unset / enabled / disabled) is
    # sent to AWS only when present.
    encryption_support = optional(bool)

    # CIDR blocks to associate with the Transit Gateway. Used for advanced
    # features like TGW Connect (SD-WAN/third-party appliance integration
    # over GRE) where the gateway itself needs routable addresses.
    #
    # Most deployments do not need TGW CIDR blocks -- leave empty unless you
    # are using TGW Connect. Maximum 5 blocks. IPv4 blocks must be /24 or
    # larger (a numerically smaller prefix length), IPv6 blocks /64 or
    # larger, and the link-local range 169.254.0.0/16 is rejected by AWS.
    transit_gateway_cidr_blocks = optional(list(string), [])
  })
}
