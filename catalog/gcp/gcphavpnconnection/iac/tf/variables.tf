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
  description = "GcpHaVpnConnection specification"
  type = object({
    # The GCP project the tunnels and router interfaces are created in: a
    # reference to a GcpProject or the project ID as a literal. Must be the
    # gateway's project. Empty means the provider's default project.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The HA VPN gateway the tunnels leave from: a reference to a
    # GcpHaVpnGateway (its gateway_self_link output) or the gateway's self
    # link as a literal. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    gateway = string

    # The Cloud Router the BGP sessions run on -- the gateway's router: a
    # reference to the same GcpHaVpnGateway (its router_name output) or the
    # router's name as a literal. Must be in the gateway's region and
    # network. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    router = string

    # The region of the gateway and router -- a reference to the same
    # GcpHaVpnGateway (its region output) or the region as a literal, so a
    # connection can never name a region its gateway is not in. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    region = string

    # What the tunnels connect to: an external device or another Google
    # Cloud HA VPN gateway. Required.
    peer = object({
      # The other side is an on-premises appliance or another cloud's VPN
      # endpoint, described by its public addresses. This connection creates
      # the external VPN gateway resource for it.
      external_gateway = optional(object({
        # Name of the external VPN gateway resource in GCP. Defaults to the
        # connection's metadata.name when empty. 1-63 characters, lowercase
        # letters, digits, and hyphens, starting with a letter. Immutable.
        name = optional(string, "")

        # How many public addresses the device exposes, which fixes how many
        # interfaces it has and which HA topology the tunnels form:
        #   "SINGLE_IP_INTERNALLY_REDUNDANT" -- one address (the device is
        #       redundant behind it); two tunnels from the two gateway
        #       interfaces to that one address earn 99.9%
        #   "TWO_IPS_REDUNDANCY"  -- two addresses (two devices or one with two
        #       uplinks); one tunnel per gateway interface to each address --
        #       two tunnels -- earn Google's 99.99% SLA. The recommended shape.
        #   "FOUR_IPS_REDUNDANCY" -- four addresses; two tunnels per gateway
        #       interface, four in all, also 99.99%
        # Immutable.
        redundancy_type = string

        # The device's public addresses, one per interface, ids 0..n-1. The
        # count must match redundancy_type. Immutable.
        interfaces = list(object({
          # The interface's numeric id. Which ids are legal depends on the
          # gateway's redundancy_type: 0 for SINGLE_IP_INTERNALLY_REDUNDANT; 0 and
          # 1 for TWO_IPS_REDUNDANCY; 0-3 for FOUR_IPS_REDUNDANCY. Each id appears
          # once.
          id = optional(number, 0)

          # The public IPv4 address of the device's interface. Cannot be a Google
          # Compute Engine address (use peer.gcp_gateway for Google-to-Google).
          # Set this or ipv6_address, matching the gateway's gateway_ip_version.
          ip_address = optional(string, "")

          # The public IPv6 address of the device's interface, in any RFC 4291
          # form (e.g. 2001:db8::2d9:51:0:0). Cannot be a Compute Engine address.
          ipv6_address = optional(string, "")
        }))

        # Human-readable description of the device. Immutable.
        description = optional(string, "")

        # User labels on the external gateway resource, merged with Planton's
        # platform labels (which win on key conflicts). Mutable.
        labels = optional(map(string), {})
      }))

      # The other side is another Google Cloud HA VPN gateway -- VPN between
      # two VPCs, in the same or different projects or organizations: a
      # reference to a GcpHaVpnGateway (its gateway_self_link output) or the
      # gateway's self link as a literal. Google pairs interfaces
      # automatically (tunnel on interface 0 here connects to interface 0
      # there), so tunnels leave peer_external_gateway_interface empty. The
      # other VPC declares its own GcpHaVpnConnection pointing back at this
      # side's gateway; the two together form the link.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      gcp_gateway = optional(string, "")
    })

    # The tunnels, one to four, each with its BGP session. Two -- one per
    # gateway interface -- is the recommended HA shape.
    tunnels = list(object({
      # Name of the tunnel in GCP. Required; unique in the region. 1-63
      # characters, lowercase letters, digits, and hyphens, starting with a
      # letter. Also the default name of the tunnel's BGP session. Immutable.
      name = string

      # Which of the Google gateway's two interfaces this tunnel leaves from:
      # 0 or 1. For the 99.99% SLA put at least one tunnel on each. Immutable.
      vpn_gateway_interface = optional(number, 0)

      # Which interface of the EXTERNAL gateway this tunnel lands on (its id
      # in peer.external_gateway.interfaces). Set when the peer is an external
      # device; leave empty when the peer is a Google Cloud gateway (Google
      # pairs interfaces itself). Sent only when set. Immutable.
      peer_external_gateway_interface = optional(number)

      # The IKE pre-shared key both ends authenticate with. Sensitive: never
      # logged or exported; in Terraform state only its hash is kept. No
      # content rule, because sensitive fields hold a managed-secret reference
      # on consuming platforms and a content-shape rule would reject every
      # reference. Google accepts 1-63 printable characters. Immutable: a
      # rotation recreates the tunnel.
      shared_secret = string

      # IKE protocol version: 1 or 2. 2 when empty (Google's default and the
      # only version that supports IPv6, BFD-friendly rekeying, and modern
      # ciphers). Use 1 only for a legacy device. Always sent. Immutable.
      ike_version = optional(number)

      # Local traffic selectors -- the IPv4 CIDRs on the Google side the
      # tunnel carries when the tunnel is used in ROUTE-BASED (policy-less)
      # mode without BGP. With a BGP session (the normal HA VPN shape) Google
      # sets 0.0.0.0/0 and this stays empty; sent only when set. The ranges
      # must be disjoint. Immutable.
      local_traffic_selector = optional(list(string), [])

      # Remote traffic selectors -- the IPv4 CIDRs on the peer side, same
      # rules as local_traffic_selector. Immutable.
      remote_traffic_selector = optional(list(string), [])

      # Restrict the IKE ciphers this tunnel negotiates. Leave unset to accept
      # Google's defaults (which already exclude broken algorithms). Immutable.
      cipher_suite = optional(object({
        # Phase-1 (IKE SA) ciphers.
        phase1 = optional(object({
          # Encryption algorithms.
          encryption = optional(list(string), [])

          # Integrity algorithms.
          integrity = optional(list(string), [])

          # Pseudo-random functions.
          prf = optional(list(string), [])

          # Diffie-Hellman groups.
          dh = optional(list(string), [])
        }))

        # Phase-2 (child / IPsec SA) ciphers.
        phase2 = optional(object({
          # Encryption algorithms.
          encryption = optional(list(string), [])

          # Integrity algorithms.
          integrity = optional(list(string), [])

          # Perfect-forward-secrecy groups.
          pfs = optional(list(string), [])
        }))
      }))

      # User labels on the tunnel, merged with Planton's platform labels
      # (which win on key conflicts). The one mutable field on a tunnel.
      labels = optional(map(string), {})

      # The BGP session inside this tunnel: the router interface and peer.
      # Required -- an HA VPN tunnel without BGP carries no routes.
      bgp_session = object({
        # Name of the Cloud Router interface and BGP peer this session creates
        # (both take the same name). Defaults to the tunnel's name when empty.
        # RFC 1035, 1-63 characters; unique on the router. Immutable.
        name = optional(string, "")

        # The Cloud Router's address on the tunnel, as a link-local CIDR from
        # 169.254.0.0/16 with a /30 mask (e.g. "169.254.10.1/30"): Google takes
        # the address as its own end and the peer takes the other usable address
        # of the /30. Every session on the router needs a different /30 and it
        # must not overlap the router's identifier_range. Required for IPv4
        # sessions; Google assigns one when empty only for IPv6-only sessions.
        # Immutable.
        interface_ip_range = optional(string, "")

        # The IP version of the router interface:
        #   ""     -- Google infers it from interface_ip_range (IPv4) or from the
        #             gateway's stack type
        #   "IPV4"
        #   "IPV6" -- for IPv6-only sessions on an IPV6_ONLY or dual-stack
        #             gateway; Google assigns the link-local addresses
        # Sent only when set. Immutable.
        ip_version = optional(string)

        # The peer router's ASN -- the number the on-premises device (or the
        # other Google Cloud router) speaks as. Must differ from this router's
        # ASN. Any public or private ASN. Required. Mutable.
        peer_asn = number

        # The peer's address on the tunnel: the other usable address of
        # interface_ip_range (e.g. "169.254.10.2" for 169.254.10.1/30). Google
        # derives it from the interface range when empty, so it is normally left
        # unset; set it only when the device insists on a specific address.
        # Sent only when set. Immutable.
        peer_ip_address = optional(string)

        # The priority (MED) of routes advertised to this peer: where the peer
        # has more than one matching route of equal length, the LOWEST value
        # wins. Set different priorities on the tunnels of one connection to
        # make the peer prefer one tunnel (active/passive) instead of splitting
        # traffic (active/active, the default when equal). 0-65535; sent only
        # when set, so 0 is a real priority distinct from "let Google choose".
        # Mutable.
        advertised_route_priority = optional(number)

        # Per-session override of the router's advertisement mode. Empty
        # inherits the router's; DEFAULT advertises every subnet; CUSTOM
        # advertises only advertised_groups and advertised_ip_ranges below.
        # Mutable.
        advertise_mode = optional(string, "")

        # Prefix groups to advertise to this peer in CUSTOM mode; the API
        # accepts exactly ALL_SUBNETS. Overrides the router's list for this
        # session.
        advertised_groups = optional(list(string), [])

        # Individual ranges to advertise to this peer in CUSTOM mode, in
        # addition to any advertised_groups. Overrides the router's list for
        # this session.
        advertised_ip_ranges = optional(list(object({
          # The IP range to advertise, in CIDR form (e.g. "10.10.0.0/16").
          range = string

          # Human-readable description of this advertised range.
          description = optional(string, "")
        })), [])

        # Whether the session is up. Default true. Setting false tears the
        # session down and withdraws its routes without deleting anything -- the
        # switch for draining a tunnel before maintenance. Always sent. Mutable.
        enable = optional(bool)

        # Carry IPv4 routes over this session. Google enables it by default when
        # the peer address is IPv4, so it is sent only when set -- set false on
        # a dual-stack gateway to make a session IPv6-only. Mutable.
        enable_ipv4 = optional(bool)

        # Carry IPv6 routes over this session. Default false. Requires a gateway
        # with stack_type IPV4_IPV6 or IPV6_ONLY and a VPC with an internal IPv6
        # range. Mutable.
        enable_ipv6 = optional(bool, false)

        # Google Cloud's IPv6 next-hop address for the session, from
        # 2600:2d00:0:2::/64 or 2600:2d00:0:3::/64. Google assigns one when
        # empty; sent only when set. Mutable.
        ipv6_nexthop_address = optional(string)

        # The peer's IPv6 next-hop address, same range and rules as
        # ipv6_nexthop_address. Sent only when set. Mutable.
        peer_ipv6_nexthop_address = optional(string)

        # Google Cloud's IPv4 next-hop address for an IPv4-over-IPv6 session
        # (IPv4 routes exchanged over an IPv6 BGP session). Google assigns one
        # when empty; sent only when set. Mutable.
        ipv4_nexthop_address = optional(string)

        # The peer's IPv4 next-hop address for an IPv4-over-IPv6 session. Sent
        # only when set. Mutable.
        peer_ipv4_nexthop_address = optional(string)

        # Ranges to treat as learned from this peer even though it did not
        # advertise them -- for devices that cannot advertise everything behind
        # them. Mutable.
        custom_learned_ip_ranges = optional(list(object({
          # A CIDR prefix (an address without a mask is /32 or /128).
          range = string
        })), [])

        # Priority applied to every custom learned range of this session.
        # 0-65335; Google uses 100 when empty. Sent only when set, so 0 is a real
        # priority. Mutable.
        custom_learned_route_priority = optional(number)

        # BFD for fast failure detection on this session. Leave unset to rely
        # on BGP keepalives alone.
        bfd = optional(object({
          # Who starts the BFD session:
          #   "ACTIVE"   -- Cloud Router initiates (the usual choice)
          #   "PASSIVE"  -- Cloud Router waits for the peer to initiate
          #   "DISABLED" -- BFD off for this session
          session_initialization_mode = string

          # Minimum interval in milliseconds between BFD packets RECEIVED from the
          # peer; the negotiated value is the greater of this and the peer's
          # transmit interval. 1000-30000; 1000 when empty.
          min_receive_interval = optional(number, 0)

          # Minimum interval in milliseconds between BFD packets TRANSMITTED to
          # the peer; negotiated the same way. 1000-30000; 1000 when empty.
          min_transmit_interval = optional(number, 0)

          # How many consecutive BFD packets may be missed before the peer is
          # declared down. 5-16; 5 when empty. Detection time is roughly the
          # negotiated interval times this multiplier.
          multiplier = optional(number, 0)
        }))

        # MD5 authentication of the BGP session. Leave unset for an
        # unauthenticated session (the tunnel's IPsec already protects the
        # control traffic; MD5 defends against a misconfigured peer, not an
        # attacker).
        md5_authentication_key = optional(object({
          # Name of the key. RFC 1035, 1-63 characters. Defaults to the tunnel's
          # name with a `-md5` suffix when empty.
          name = optional(string, "")

          # The key material -- the same string configured on the peer router.
          # Sensitive: never logged or exported. No content rule, because sensitive
          # fields hold a managed-secret reference on consuming platforms and a
          # content-shape rule would reject every reference.
          key = string
        }))

        # Names of route policies (ROUTE_POLICY_TYPE_IMPORT) on the router,
        # applied to routes learned from this peer in order. Route policies are
        # created outside this kind (`gcloud compute routers update-route-policy`
        # or the provider's router route-policy resource) and referenced here by
        # name. Mutable.
        import_policies = optional(list(string), [])

        # Names of route policies (ROUTE_POLICY_TYPE_EXPORT) on the router,
        # applied to routes advertised to this peer in order. Same provenance as
        # import_policies. Mutable.
        export_policies = optional(list(string), [])
      })

      # Human-readable description of the tunnel. Immutable (Google keeps it
      # create-time on this resource).
      description = optional(string, "")
    }))

    # Resource Manager tags bound to every tunnel (and the external gateway,
    # when created) for org-policy and IAM conditions. Keys `tagKeys/{id}`,
    # values `tagValues/{id}`. Create-time only: a change replaces the
    # tunnels.
    resource_manager_tags = optional(map(string), {})

    # What destroying this resource does to the tunnels, router interfaces,
    # BGP peers, and external gateway in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- all are deleted; the site is disconnected
    #   "PREVENT" -- destroy FAILS; the guard for a production site link
    #   "ABANDON" -- the resources leave management but stay live in GCP
    deletion_policy = optional(string, "")
  })
}
