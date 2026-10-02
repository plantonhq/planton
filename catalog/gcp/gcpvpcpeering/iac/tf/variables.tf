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
  description = "GcpVpcPeering specification"
  type = object({
    # The name of the peering entry on `network`: what `gcloud compute
    # networks peerings list` shows, and the name the other side does NOT
    # need to match (each side names its own entry). Defaults to
    # metadata.name when empty. 1-63 characters, lowercase letters, digits,
    # and hyphens, starting with a letter. Immutable in the CREATE form.
    #
    # In the ROUTES-CONFIG form this is the name of the EXISTING peering
    # whose routes are managed -- for Google-managed private services access
    # it is `servicenetworking-googleapis-com`.
    peering_name = optional(string, "")

    # The VPC network this side of the peering lives in: a reference to a
    # GcpVpcNetwork (its network_self_link output) or the network's self link
    # as a literal. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = string

    # The OTHER network -- the one this side peers with: a reference to a
    # GcpVpcNetwork or its self link as a literal
    # (`projects/{project}/global/networks/{name}` or the full
    # `https://www.googleapis.com/compute/v1/...` form). May live in another
    # project or another organization. Immutable.
    #
    # Set it and this resource CREATES the peering. Leave it EMPTY and this
    # resource manages the route exchange of a peering that already exists
    # on `network` under `peering_name` (the routes-config form above).
    #
    # The peer network is where the traffic goes, not where this resource
    # lives, so it is an access edge on diagrams, not a containment edge.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    peer_network = optional(string, "")

    # Export this network's CUSTOM routes (static routes and dynamic routes
    # learned over Cloud VPN / Interconnect / Router Appliance) to the peer.
    # Default false: only subnet routes cross a peering unless both sides opt
    # in -- this side exports, the other side imports. The one flag that
    # makes an on-premises network reachable from the peer (or, on a
    # service-networking peering, makes Cloud SQL reachable from
    # on-premises). Mutable. Always sent; in the routes-config form the
    # provider requires it.
    export_custom_routes = optional(bool, false)

    # Import the peer network's custom routes into this network. Default
    # false. Takes effect only when the peer's side exports them. Mutable.
    # Always sent; in the routes-config form the provider requires it.
    import_custom_routes = optional(bool, false)

    # Export subnet routes whose ranges are PUBLIC IPs (privately used public
    # IP ranges) to the peer. Default TRUE -- Google exports them unless told
    # not to. Always sent on both forms (the default when unset, so the
    # manifest states the exchange in full). Immutable in the CREATE form
    # (changing it recreates the peering); changes in place in the
    # routes-config form.
    export_subnet_routes_with_public_ip = optional(bool)

    # Import the peer's public-IP subnet routes into this network. Default
    # FALSE -- a network does not learn a peer's privately-used public ranges
    # unless it asks to. Always sent on both forms (the default when unset).
    # Immutable in the CREATE form; changes in place in the routes-config
    # form.
    import_subnet_routes_with_public_ip = optional(bool)

    # Which IP versions the peering carries:
    #   ""          -- same as IPV4_ONLY (provider default)
    #   "IPV4_ONLY" -- only IPv4 subnet ranges are exchanged
    #   "IPV4_IPV6" -- IPv4 and IPv6 ranges; both networks must be dual-stack
    #                  (an internal IPv6 range on each GcpVpcNetwork)
    # CREATE form only. Mutable.
    stack_type = optional(string, "")

    # How a change to this peering's settings is applied:
    #   ""            -- same as INDEPENDENT (provider default)
    #   "INDEPENDENT" -- this side's route-exchange flags take effect on
    #                    their own, as soon as applied
    #   "CONSENSUS"   -- a changed setting takes effect only once BOTH sides
    #                    agree on it; until then the peering keeps the old
    #                    behavior. Use it when the two sides are owned by
    #                    different teams and a one-sided flip must never
    #                    change what traffic flows.
    # CREATE form only. Mutable.
    update_strategy = optional(string, "")

    # What destroying this resource does to the peering in GCP (CREATE form
    # only -- destroying the routes-config form never touches the peering):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- this side's peering entry is removed; the pair goes
    #                INACTIVE and traffic stops. The other side's entry stays
    #                until it is removed too.
    #   "PREVENT" -- destroy FAILS; the guard for the peering production
    #                traffic depends on
    #   "ABANDON" -- the resource leaves management but the peering entry
    #                stays live
    deletion_policy = optional(string, "")
  })
}
