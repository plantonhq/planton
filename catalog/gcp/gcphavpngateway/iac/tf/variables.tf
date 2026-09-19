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
  description = "GcpHaVpnGateway specification"
  type = object({
    # The GCP project the gateway and router are created in: a reference to
    # a GcpProject or the project ID as a literal. Empty means the provider's
    # default project.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the VPN gateway in GCP. Defaults to metadata.name when empty.
    # 1-63 characters, lowercase letters, digits, and hyphens, starting with
    # a letter. Immutable.
    gateway_name = optional(string, "")

    # The region the gateway and router live in (e.g. "us-central1"). Every
    # tunnel of every connection on this gateway is in this region. Immutable.
    region = string

    # The VPC network the gateway attaches to: a reference to a GcpVpcNetwork
    # (its network_self_link output) or the network's self link as a literal.
    # The tunnels carry traffic for THIS network's subnets (and, with custom
    # routes, whatever the router advertises). Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = string

    # Human-readable description of the gateway. Immutable (Google keeps it
    # create-time on this resource).
    description = optional(string, "")

    # The IP version of the gateway's two public interface addresses:
    #   ""     -- same as IPV4 (provider default)
    #   "IPV4" -- two public IPv4 addresses; what every on-premises device
    #             and every peer cloud supports
    #   "IPV6" -- two public IPv6 addresses; the peer must reach the gateway
    #             over IPv6 (Google-to-Google over IPv6, or an IPv6-capable
    #             device). The traffic INSIDE the tunnels is governed by
    #             stack_type, not by this.
    # Immutable.
    gateway_ip_version = optional(string, "")

    # Which IP versions the TUNNELS carry between the networks:
    #   ""          -- same as IPV4_ONLY (provider default)
    #   "IPV4_ONLY" -- IPv4 traffic only
    #   "IPV4_IPV6" -- dual-stack: the VPC must have an internal IPv6 range
    #                  and each BGP session enables IPv6 to exchange IPv6
    #                  routes
    #   "IPV6_ONLY" -- IPv6 traffic only
    # Immutable.
    stack_type = optional(string, "")

    # Pin the gateway's interfaces to Cloud Interconnect VLAN attachments for
    # HA VPN over Interconnect. Empty (the normal case) means an
    # internet-facing gateway whose interfaces get public IPs. At most two
    # entries, one per interface. Requires
    # router.encrypted_interconnect_router. Immutable.
    vpn_interfaces = optional(list(object({
      # Which of the gateway's two interfaces this entry configures: 0 or 1.
      # Each interface may appear once.
      id = optional(number, 0)

      # The Interconnect attachment this interface rides, as a self link or
      # `projects/{p}/regions/{r}/interconnectAttachments/{name}`. The
      # attachment must be an encrypted attachment (`encryption: IPSEC`) in the
      # same region, and the gateway's router must be an
      # encrypted_interconnect_router. Immutable.
      interconnect_attachment = string
    })), [])

    # User labels on the gateway, merged with Planton's platform labels
    # (which win on key conflicts). The one mutable surface on the gateway
    # resource itself.
    labels = optional(map(string), {})

    # Resource Manager tags bound to the gateway for org-policy and IAM
    # conditions. Keys `tagKeys/{id}`, values `tagValues/{id}`. Create-time
    # only: a change replaces the gateway (and its public IPs).
    resource_manager_tags = optional(map(string), {})

    # The Cloud Router the gateway's tunnels terminate BGP on -- its name,
    # ASN, and default advertisement. Required: an HA VPN without a BGP
    # router routes nothing.
    router = object({
      # Name of the Cloud Router. Defaults to the gateway's name when empty --
      # a router and a gateway are different resource types, so the same name
      # never collides. 1-63 characters, lowercase letters, digits, and
      # hyphens, starting with a letter. Immutable.
      name = optional(string, "")

      # Human-readable description of the router. Mutable.
      description = optional(string, "")

      # The router's BGP configuration: the ASN Google speaks as and the
      # default route advertisement. Required -- an HA VPN router exists to
      # run BGP, and every tunnel session needs the ASN.
      bgp = object({
        # Local BGP Autonomous System Number -- the number Google Cloud speaks as
        # toward every on-premises or peer-cloud router. Must be an RFC 6996
        # private ASN, 16-bit (64512-65534) or 32-bit (4200000000-4294967294),
        # and DIFFERENT from every peer's ASN. Fixed for the router's lifetime;
        # a change recreates the router and every tunnel session on it.
        asn = number

        # Route advertisement mode. DEFAULT (the value when empty): advertise
        # every subnet range of the VPC to every peer. CUSTOM: advertise only
        # what advertised_groups and advertised_ip_ranges specify -- the shape
        # for exposing a curated set of ranges to the other side.
        advertise_mode = optional(string, "")

        # Prefix groups to advertise in CUSTOM mode, in addition to
        # advertised_ip_ranges. The API accepts exactly one group value:
        # ALL_SUBNETS (re-adds the default subnet advertisement on top of the
        # custom ranges).
        advertised_groups = optional(list(string), [])

        # Individual IP ranges to advertise in CUSTOM mode, sent to all peers
        # in addition to any advertised_groups.
        advertised_ip_ranges = optional(list(object({
          # The IP range to advertise, in CIDR form (e.g. "10.10.0.0/16").
          range = string

          # Human-readable description of this advertised range.
          description = optional(string, "")
        })), [])

        # Interval in seconds between BGP keepalive messages (20-60; 20 when
        # empty). Hold time -- how long a peer waits before declaring the session
        # dead -- is three times this value. Lower values fail over faster at
        # the cost of more control traffic; BFD on the session is the precise
        # tool for sub-second detection.
        keepalive_interval = optional(number, 0)

        # Explicit range of valid BGP identifiers (what other vendors call the
        # router ID), as a link-local IPv4 CIDR from 169.254.0.0/16 of size at
        # least /30 (e.g. "169.254.8.0/30"). Must not overlap any session's
        # interface range on this router. Leave empty to let GCP choose.
        identifier_range = optional(string, "")
      })

      # Dedicate this router to encrypted Interconnect VLAN attachments (HA VPN
      # over Cloud Interconnect). Required when vpn_interfaces pins the gateway
      # to attachments; must be false otherwise. Immutable: an encrypted router
      # cannot be converted later.
      encrypted_interconnect_router = optional(bool, false)

      # Resource Manager tags bound to the router for org-policy and IAM
      # conditions. Keys `tagKeys/{id}`, values `tagValues/{id}`. Create-time
      # only: a change replaces the router.
      resource_manager_tags = optional(map(string), {})
    })

    # What destroying this resource does to the gateway and router in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- both are deleted; fails while any GcpHaVpnConnection's
    #                tunnels still reference the gateway (destroy the
    #                connections first -- a chart's dependency order does
    #                this when they reference the gateway)
    #   "PREVENT" -- destroy FAILS; the guard for the gateway every site's
    #                device is configured against
    #   "ABANDON" -- the resources leave management but stay live in GCP
    deletion_policy = optional(string, "")
  })
}
