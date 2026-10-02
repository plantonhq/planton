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
  description = "GcpRouterNat specification"
  type = object({
    # GCP project ID where the Cloud Router and NAT will be created.
    # If not specified, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Cloud Router to create in GCP.
    # Must be 1-63 characters, lowercase letters, numbers, or hyphens.
    # Must start with a lowercase letter and end with a lowercase letter or number.
    # Immutable after creation.
    router_name = string

    # Name of the NAT configuration on the Cloud Router.
    # Must be 1-63 characters, lowercase letters, numbers, or hyphens.
    # Must start with a lowercase letter and end with a lowercase letter or number.
    # Immutable after creation.
    nat_name = string

    # GCP region for the Cloud Router and NAT (e.g., "us-central1").
    # Immutable after creation.
    region = string

    # The VPC network the router attaches to. Accepts a network self link.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_self_link = string

    # The router's BGP configuration — ASN, route advertisement, keepalive,
    # identifier range. Only needed when the router also serves BGP
    # sessions (Cloud VPN / Interconnect on the same router); NAT-only
    # routers leave this unset and GCP assigns an ASN.
    router_bgp = optional(object({
      # Local BGP Autonomous System Number. Must be an RFC6996 private ASN,
      # 16-bit (64512-65534) or 32-bit (4200000000-4294967294). Fixed for the
      # router's lifetime; every VPN tunnel and Interconnect attachment on
      # this router shares it.
      asn = optional(number, 0)

      # Route advertisement mode. DEFAULT (the value when empty): advertise
      # all subnets visible to the router. CUSTOM: advertise only what
      # advertised_groups and advertised_ip_ranges specify — the shape for
      # exposing a curated set of ranges over VPN/Interconnect.
      advertise_mode = optional(string, "")

      # Prefix groups to advertise in CUSTOM mode, in addition to
      # advertised_ip_ranges. The API accepts exactly one group value:
      # ALL_SUBNETS (re-adds the default subnet advertisement on top of the
      # custom ranges).
      advertised_groups = optional(list(string), [])

      # Individual IP ranges to advertise in CUSTOM mode, sent to all peers
      # of the router in addition to any advertised_groups.
      advertised_ip_ranges = optional(list(object({
        # The IP range to advertise, in CIDR form (e.g. "10.10.0.0/16").
        range = string

        # Human-readable description of this advertised range.
        description = optional(string, "")
      })), [])

      # Interval in seconds between BGP keepalive messages (20-60, default
      # 20). Hold time — how long a peer waits before declaring the session
      # dead — is three times this value.
      keepalive_interval = optional(number, 0)

      # Explicit range of valid BGP identifiers (what other vendors call the
      # router ID), as a link-local IPv4 CIDR from 169.254.0.0/16 of size at
      # least /30 (e.g. "169.254.8.0/30"). Must not overlap any IPv4 BGP
      # session range on the router. Leave empty to let GCP choose.
      identifier_range = optional(string, "")
    }))

    # Human-readable description of the Cloud Router's purpose.
    router_description = optional(string, "")

    # NAT gateway type. PUBLIC (default): NAT to the internet using external
    # IPs. PRIVATE: NAT between VPC networks (Network Connectivity Center
    # spokes) using subnetwork ranges — no external IPs involved.
    # Immutable after creation.
    type = optional(string, "")

    # Which subnetworks (and which of their IP ranges) get NAT.
    # ALL_SUBNETWORKS_ALL_IP_RANGES (default when empty and subnetworks is
    # empty): every subnetwork in the region, primary + secondary ranges.
    # ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES: every subnetwork, primary only.
    # LIST_OF_SUBNETWORKS (implied when subnetworks is non-empty): only the
    # subnetworks listed below.
    source_subnetwork_ip_ranges_to_nat = optional(string, "")

    # Per-subnetwork NAT scoping. Listing any subnetwork here selects
    # LIST_OF_SUBNETWORKS mode: only the listed subnetworks are NATed.
    subnetworks = optional(list(object({
      # The subnetwork to NAT. Accepts a subnetwork self link.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetwork = string

      # Which IP ranges of the subnetwork are NATed.
      # ALL_IP_RANGES: primary and all secondary ranges (the default when empty).
      # PRIMARY_IP_RANGE: only the primary range.
      # LIST_OF_SECONDARY_IP_RANGES: only the secondary ranges named in
      # secondary_ip_range_names — the shape for NATing GKE pod ranges without
      # exposing the node range, or vice versa.
      source_ip_ranges_to_nat = optional(list(string), [])

      # Names of the secondary ranges to NAT. Required when
      # source_ip_ranges_to_nat contains LIST_OF_SECONDARY_IP_RANGES.
      secondary_ip_range_names = optional(list(string), [])
    })), [])

    # Static external IPs the NAT translates through, each referencing a
    # GcpAddress reservation (EXTERNAL, in this region) by self link.
    # Non-empty selects MANUAL_ONLY allocation — the shape for stable,
    # allowlistable egress IPs. Empty selects AUTO_ONLY (GCP manages the
    # pool). Public NAT only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    nat_ips = optional(list(string), [])

    # IPs being drained: existing connections continue, new connections stop.
    # Each entry must already be present in nat_ips (the API enforces this and
    # rejects drain entries on a brand-new NAT). Public NAT only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    drain_nat_ips = optional(list(string), [])

    # Network tier for auto-allocated NAT IPs: PREMIUM (default) or STANDARD.
    # Only meaningful for public NAT with auto-allocation.
    auto_network_tier = optional(string, "")

    # Minimum ports reserved per VM (default: 64 static, 32 dynamic). Raise
    # when instances open many concurrent connections to the same
    # destination; must be a power of two when dynamic port allocation is
    # enabled.
    min_ports_per_vm = optional(number, 0)

    # Maximum ports per VM. Only valid with dynamic port allocation, where it
    # caps how far a busy VM's allocation can grow; must be a power of two.
    max_ports_per_vm = optional(number, 0)

    # Dynamic port allocation: VMs start at min_ports_per_vm and grow toward
    # max_ports_per_vm on demand — better pool utilization for mixed
    # workloads. Mutually exclusive with endpoint-independent mapping.
    enable_dynamic_port_allocation = optional(bool, false)

    # Endpoint-independent mapping: the same internal ip:port maps to the
    # same NAT ip:port regardless of destination (RFC 5128) — required by
    # some peer-to-peer and SIP workloads. Mutually exclusive with dynamic
    # port allocation.
    enable_endpoint_independent_mapping = optional(bool, false)

    # Which resource types this NAT serves: ENDPOINT_TYPE_VM (default),
    # ENDPOINT_TYPE_SWG (Secure Web Gateway), or ENDPOINT_TYPE_MANAGED_PROXY_LB
    # (regional load balancer proxies). A NAT serves exactly one endpoint
    # type. Immutable after creation.
    endpoint_types = optional(list(string), [])

    # Idle timeout for UDP connections, in seconds (default 30).
    udp_idle_timeout_sec = optional(number, 0)

    # Idle timeout for ICMP connections, in seconds (default 30).
    icmp_idle_timeout_sec = optional(number, 0)

    # Idle timeout for established TCP connections, in seconds (default 1200).
    tcp_established_idle_timeout_sec = optional(number, 0)

    # Idle timeout for transitory (half-open) TCP connections, in seconds
    # (default 30).
    tcp_transitory_idle_timeout_sec = optional(number, 0)

    # How long a TCP connection lingers in TIME_WAIT before its NAT port is
    # reusable, in seconds (default 120). Lowering this frees ports faster
    # for high-churn workloads at the cost of stricter RFC conformance.
    tcp_time_wait_timeout_sec = optional(number, 0)

    # NAT rules: route matching egress connections through dedicated NAT IPs
    # or ranges (e.g. a stable source IP for one partner's API).
    rules = optional(list(object({
      # Rule priority (0-65000). Lower numbers win when multiple rules match.
      rule_number = optional(number, 0)

      # CEL expression selecting the traffic this rule applies to, evaluated
      # against destination attributes. Example:
      # "destination.ip == '203.0.113.10' || destination.ip == '203.0.113.11'"
      match = string

      # Human-readable description of the rule's intent.
      description = optional(string, "")

      # The NAT IPs or ranges used for matching traffic.
      action = optional(object({
        # Static external IPs to NAT matching traffic through (public NAT only).
        # Each entry references a GcpAddress reservation by self link.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        source_nat_active_ips = optional(list(string), [])

        # IPs being drained out of this rule (public NAT only): existing
        # connections continue, no new connections are established. An IP must
        # already be in source_nat_active_ips before it can be drained.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        source_nat_drain_ips = optional(list(string), [])

        # Subnetworks whose primary ranges provide the NAT addresses for matching
        # traffic (private NAT only).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        source_nat_active_ranges = optional(list(string), [])

        # Subnetwork ranges being drained (private NAT only).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        source_nat_drain_ranges = optional(list(string), [])
      }))
    })), [])

    # Log filter for NAT translation logging.
    # **Default:** ERRORS_ONLY (recommended for production to detect port exhaustion and connection failures).
    # Use DISABLED for non-production environments to reduce costs.
    # Use ALL for security auditing or detailed troubleshooting (generates significant log volume).
    log_filter = optional(string)

    # Dedicate the router to encrypted VLAN attachments (HA VPN over Cloud
    # Interconnect). An encrypted-Interconnect router carries only
    # encrypted attachments and cannot be converted later. Immutable after
    # creation.
    encrypted_interconnect_router = optional(bool, false)

    # Resource Manager tags bound to the Cloud Router for org-policy and
    # IAM conditions. Keys in the form "tagKeys/{id}", values
    # "tagValues/{id}". Create-time only: changing them later replaces the
    # router.
    resource_manager_tags = optional(map(string), {})

    # Deletion policy for the router and NAT — what happens when this
    # resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the router and NAT are deleted
    #   "PREVENT" -- destroy FAILS; protects the egress path of a
    #                production fleet from accidental teardown
    #   "ABANDON" -- the router and NAT are removed from management but
    #                keep serving traffic in GCP
    deletion_policy = optional(string, "")

    # NAT64 scope — which subnetworks get IPv6-to-IPv4 translation.
    # ALL_IPV6_SUBNETWORKS: every subnetwork's IPv6 ranges are translated
    # (only ONE NAT per region in the network may claim this).
    # LIST_OF_IPV6_SUBNETWORKS: only the subnetworks in nat64_subnetworks.
    # Leave empty for no NAT64 (IPv4-only NAT).
    source_subnetwork_ip_ranges_to_nat64 = optional(string, "")

    # Subnetworks whose IPv6 traffic the NAT64 gateway translates. Each
    # entry references a dual-stack GcpSubnetwork by self link. Listing any
    # subnetwork here selects LIST_OF_IPV6_SUBNETWORKS mode (mirroring how
    # subnetworks implies LIST_OF_SUBNETWORKS).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    nat64_subnetworks = optional(list(string), [])
  })
}
