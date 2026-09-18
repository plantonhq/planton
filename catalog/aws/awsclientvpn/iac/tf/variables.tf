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
  description = "AwsClientVpn specification"
  type = object({
    # The AWS region where the Client VPN endpoint will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Human-friendly description shown in the AWS console.
    description = optional(string, "")

    # How clients prove who they are before a tunnel is established. One or
    # two options (a client passes if it satisfies ANY one); all ForceNew —
    # changing authentication replaces the endpoint. See
    # AwsClientVpnAuthenticationOption for the three types.
    authentication_options = list(object({
      # Authentication type. Values:
      #
      # - "certificate-authentication": mutual TLS — the client presents a
      #   certificate issued from the chain in `root_certificate_chain_arn`.
      #   Zero external identity infrastructure; certificate distribution and
      #   revocation are on you.
      #
      # - "directory-service-authentication": user/password against an AWS
      #   Directory Service directory (`active_directory_id`) — Managed
      #   Microsoft AD or AD Connector to on-prem AD.
      #
      # - "federated-authentication": SAML 2.0 single sign-on through an IAM
      #   SAML provider (`saml_provider_arn`) — Okta, Entra ID, and friends;
      #   the only type that supports the self-service portal.
      type = string

      # For "certificate-authentication": the ACM certificate whose CHAIN
      # client certificates must descend from — the client CA, distinct in
      # role from the endpoint's server certificate (for a self-signed setup
      # the same imported certificate may serve both roles, but never assume
      # it). Reference an imported CA certificate in ACM.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      root_certificate_chain_arn = optional(string, "")

      # For "directory-service-authentication": the AWS Directory Service
      # directory ID (e.g. "d-1234567890"). Directories have no Planton kind —
      # pass the literal ID.
      active_directory_id = optional(string, "")

      # For "federated-authentication": the ARN of the IAM SAML identity
      # provider that brokers sign-on (e.g.
      # "arn:aws:iam::123456789012:saml-provider/okta"). IAM SAML providers
      # have no Planton kind — pass the literal ARN (shape-checked here, since
      # no reference can supply it).
      saml_provider_arn = optional(string, "")

      # For "federated-authentication", optional: a second IAM SAML provider
      # used only by the self-service portal (when the portal needs a
      # different SAML app than the VPN itself).
      self_service_saml_provider_arn = optional(string, "")
    }))

    # The ACM certificate the VPN server presents to connecting clients — the
    # TLS identity of the endpoint itself, required for every authentication
    # type. Must live in ACM in the same region. Updates in place (rotating
    # the server certificate never replaces the endpoint).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    server_certificate_arn = string

    # The IPv4 range, in CIDR notation, from which connecting clients are
    # assigned addresses. Block size between /22 and /12 (AWS's documented
    # bounds — enforced here so the mistake fails at manifest time, not at
    # apply); must not overlap the VPC CIDR or any manually added route, and
    # cannot change after creation. Required — except when
    # `traffic_ip_address_type` is "ipv6", where AWS derives client
    # addressing and the field must be empty. Example: "10.100.0.0/22".
    client_cidr_block = optional(string, "")

    # Split-tunnel routing. When true, only traffic destined for the
    # endpoint's route table goes through the VPN and everything else stays
    # local to the client — the usual posture for corp-access VPNs. When
    # false (AWS's default, full tunnel), ALL client traffic enters the VPN;
    # pair that with a 0.0.0.0/0 route + authorization rule through a NAT-ed
    # subnet or clients lose internet access. Updates in place.
    split_tunnel = optional(bool, false)

    # Transport protocol for VPN sessions. Values: "udp" (AWS default —
    # lower latency, the standard OpenVPN choice), "tcp" (traverses
    # firewalls that block UDP). ForceNew.
    transport_protocol = optional(string, "")

    # Port the endpoint listens on. Values: 443 (default; blends with HTTPS
    # at the firewall) or 1194 (the traditional OpenVPN port). Either port
    # works with either transport protocol. Updates in place.
    vpn_port = optional(number)

    # IP address type of the ENDPOINT itself — which stacks clients can reach
    # it over. Values: "ipv4" (default), "ipv6", "dual-stack". ForceNew.
    endpoint_ip_address_type = optional(string, "")

    # IP address type of the traffic INSIDE the tunnel. Values: "ipv4"
    # (default), "ipv6", "dual-stack". With "ipv6", `client_cidr_block` must
    # be empty (AWS assigns client addressing). ForceNew.
    traffic_ip_address_type = optional(string, "")

    # The VPC whose security groups apply to this endpoint. Optional — when
    # omitted, AWS infers the VPC from the first associated subnet and
    # applies that VPC's default security group. Mutually exclusive with
    # `transit_gateway_configuration`. Updates in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = optional(string, "")

    # Security groups applied to the endpoint's network interfaces (max 5),
    # governing traffic between VPN clients and VPC resources. When omitted,
    # the VPC's default security group applies. Mutually exclusive with
    # `transit_gateway_configuration`. Updates in place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Subnets to associate as target networks (folded
    # network associations, one per subnet). Each association attaches the
    # endpoint to that subnet's AZ; associate subnets in two AZs for
    # resilience (each association bills hourly). All subnets must belong to
    # one VPC. May be empty — a zero-association endpoint is valid but routes
    # no traffic until a subnet is associated. Associations add/remove in
    # place, but each attach/detach takes AWS several minutes. Mutually
    # exclusive with `transit_gateway_configuration`.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Associate the endpoint with a transit gateway instead of a VPC — VPN
    # clients then reach every network the transit gateway routes to, without
    # per-subnet associations. Mutually exclusive with `vpc_id`,
    # `security_group_ids`, and `subnet_ids`. ForceNew.
    transit_gateway_configuration = optional(object({
      # The transit gateway to attach to.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      transit_gateway_id = string

      # Availability Zone NAMES the attachment spans (e.g. "us-west-2a").
      # Mutually exclusive with availability_zone_ids. Leave both empty for
      # AWS's default AZ selection.
      availability_zones = optional(list(string), [])

      # Availability Zone IDs the attachment spans (e.g. "usw2-az1") — the
      # account-independent form. Mutually exclusive with availability_zones.
      availability_zone_ids = optional(list(string), [])
    }))

    # Authorization rules — which clients may reach which destination CIDRs.
    # Without at least one rule, connected clients can reach NOTHING (the
    # endpoint authorizes no traffic by default). Rules add/remove in place.
    authorization_rules = optional(list(object({
      # The destination network being authorized, in CIDR notation — a VPC
      # CIDR, one subnet, an on-prem range, or "0.0.0.0/0" for internet access
      # (full-tunnel setups). Must be the canonical network address (host bits
      # zero: "10.0.1.0/24", never "10.0.1.5/24") — the provider rejects
      # non-canonical forms at plan time.
      target_network_cidr = string

      # Grant access only to one identity-provider group: an Active Directory
      # group SID (directory authentication) or a SAML group attribute value
      # (federated authentication). Exactly one of this and
      # `authorize_all_groups` must be set. Certificate-only endpoints have no
      # group concept — use `authorize_all_groups`. Must not contain a comma —
      # the provider uses the comma as its internal rule-ID separator, and a
      # comma here corrupts the rule's identity.
      access_group_id = optional(string, "")

      # Grant access to every authenticated client. Exactly one of this and
      # `access_group_id` must be set.
      authorize_all_groups = optional(bool, false)

      # Description shown in the AWS console.
      description = optional(string, "")
    })), [])

    # Additional routes in the endpoint's route table, beyond the route AWS
    # auto-creates for each associated subnet's VPC CIDR. The classic uses:
    # "0.0.0.0/0" through a NAT-ed subnet for full-tunnel internet egress, or
    # an on-premises CIDR through a subnet that reaches a VPN/transit
    # gateway. Each route's target subnet must be one of the associated
    # `subnet_ids`. Routes add/remove in place.
    routes = optional(list(object({
      # Destination network, in CIDR notation. "0.0.0.0/0" routes all client
      # internet traffic through the VPN (full tunnel) — the target subnet
      # then needs a NAT path or clients lose internet access. Must be the
      # canonical network address (host bits zero) — the provider rejects
      # non-canonical forms at plan time.
      destination_cidr_block = string

      # The associated subnet traffic to this destination egresses through.
      # Must be one of the endpoint's `subnet_ids` — AWS rejects a route whose
      # subnet is not (yet) associated.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_subnet_id = string

      # Description shown in the AWS console.
      description = optional(string, "")
    })), [])

    # Maximum VPN session duration in hours, after which clients must
    # re-authenticate. Values: 8, 10, 12, 24 (default 24). Updates in place.
    session_timeout_hours = optional(number)

    # When true, sessions are hard-disconnected at the session timeout and
    # the user is prompted to reconnect; when false (AWS default), the client
    # attempts to reconnect automatically. Updates in place.
    disconnect_on_session_timeout = optional(bool, false)

    # Enable the self-service portal, where users download their own client
    # configuration and reset their certificates. Only supported with a
    # federated (SAML) authentication option. Updates in place.
    self_service_portal_enabled = optional(bool, false)

    # Run a Lambda function on every new connection — allow/deny posture
    # checks beyond authentication (device compliance, source IP policy,
    # time-of-day rules). Presence enables the hook; removing the block
    # disables it (both engines send the explicit disabled state — AWS
    # otherwise keeps a once-enabled hook active forever). Updates in place.
    client_connect_options = optional(object({
      # The Lambda function AWS invokes for each connection attempt. AWS
      # requires the function name to start with "AWSClientVPN-". The function
      # receives connection metadata (user, device, source IP) and returns an
      # allow/deny decision, optionally with an error message shown to the
      # user.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      lambda_function_arn = string
    }))

    # Text banner displayed on AWS-provided clients when a session is
    # established (legal notices, acceptable-use reminders). Presence enables
    # the banner; removing the block disables it (both engines send the
    # explicit disabled state). Updates in place.
    client_login_banner = optional(object({
      # Banner text, up to 1400 UTF-8 characters — legal notices,
      # acceptable-use reminders, support contacts.
      banner_text = string
    }))

    # Enforce administrator-defined routes on connected devices, blocking
    # client-side route manipulation from bypassing the tunnel (a
    # security-posture hardening dial for managed fleets). Both engines send
    # the explicit value in both states, so flipping back to false genuinely
    # disables enforcement. Updates in place.
    client_route_enforcement_enabled = optional(bool, false)

    # Custom DNS server IPs pushed to connected clients (max 2). When empty,
    # clients keep their device DNS — set this to the VPC resolver (the VPC
    # CIDR base + 2, e.g. "10.0.0.2") so clients resolve private hosted
    # zones. Updates in place.
    dns_servers = optional(list(string), [])

    # Stream connection events (connect, disconnect, authentication failures)
    # to CloudWatch Logs. Presence enables logging — strongly recommended in
    # production; without it there is no record of who connected. Updates in
    # place.
    connection_log = optional(object({
      # The CloudWatch log group connection events are written to.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cloudwatch_log_group = string

      # Optional log stream within the group. When empty, AWS creates one.
      cloudwatch_log_stream = optional(string, "")
    }))
  })
}
