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
  description = "GcpGlobalForwardingRule specification"
  type = object({
    # The GCP project that owns the forwarding rule.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the rule.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the forwarding rule in GCP. Must be 1-63 characters: lowercase
    # letters, digits, and hyphens; must start with a letter and end with a
    # letter or digit. Private Service Connect rules that forward to Google
    # APIs are stricter: 1-20 characters, lowercase letters and digits only,
    # starting with a letter (the name doubles as the service-directory
    # entry). If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the rule — the VIP
    # itself survives only if it is a reserved static address.
    forwarding_rule_name = optional(string, "")

    # What this frontend serves and which proxy chain sits behind it — write
    # it for the operator tracing an incident from the VIP inward. Immutable.
    description = optional(string, "")

    # The scope selector. Empty builds a GLOBAL forwarding rule (the global
    # external ALB, the cross-region internal ALB, Traffic Director, PSC to
    # Google APIs); a region name such as us-central1 builds a REGIONAL one —
    # the front door of the regional external and internal Application Load
    # Balancers (target = a regional proxy), of the internal and external
    # passthrough Network Load Balancers (backend_service instead of target),
    # and of a Private Service Connect consumer endpoint (target = a service
    # attachment). Everything the rule points at must be in the same scope and
    # region: a regional proxy, a regional backend service, a regional
    # address (GcpAddress) rather than a GcpGlobalAddress. The regional-only
    # levers (backend_service, ports, all_ports, allow_global_access,
    # allow_psc_global_access, service_label, is_mirroring_collector,
    # ip_collection, recreate_closed_psc, source_ip_ranges, the L3_DEFAULT
    # protocol, the INTERNAL scheme, the STANDARD tier) are rejected when
    # region is empty; the global-only levers (metadata_filters, the
    # INTERNAL_SELF_MANAGED scheme, the backend-bucket migration canary, the
    # Service Directory region) are rejected when it is set. Immutable: a
    # rule cannot move between scopes or regions.
    region = optional(string, "")

    # The target that receives matched traffic — every proxy-based load
    # balancer's form. Reference a GcpTargetHttpsProxy (the default) or a
    # GcpTargetHttpProxy resource (a regional rule takes the regional arm of
    # the same kinds, in its own region), or provide a target URI directly —
    # other targets (target SSL/TCP proxies, target gRPC proxies, target
    # instances, target pools) attach by self-link until they exist as
    # Planton kinds. For Private Service Connect, pass the literal bundle name
    # "all-apis" or "vpc-sc" (Google APIs, global rule) or a service
    # attachment URI (a producer's service, regional rule). Exactly one of
    # target and backend_service is set: a passthrough Network Load Balancer
    # has no proxy and names its backend service instead. Mutable: GCP
    # repoints it in place (a dedicated setTarget call), enabling
    # zero-downtime frontend swaps.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    target = optional(string, "")

    # The regional backend service that receives matched traffic directly,
    # with no proxy in between — the form of the internal passthrough Network
    # Load Balancer (scheme INTERNAL) and of the backend-service-based
    # external passthrough Network Load Balancer (scheme EXTERNAL). Reference
    # a GcpBackendService declared with the same region, or provide its
    # self-link. Regional rules only, and exactly one of backend_service and
    # target: Google requires the backend service for the passthrough load
    # balancers and rejects it for every proxy-based one. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    backend_service = optional(string, "")

    # The IP address this rule accepts traffic on. Reference a
    # GcpGlobalAddress resource (the default kind, for a global rule) or a
    # regional GcpAddress resource in the rule's region (for a regional rule,
    # with valueFrom.kind: GcpAddress), provide a literal IP ("34.120.1.2"),
    # or an address resource URL. When omitted, Google Cloud assigns an
    # ephemeral IP — fine for testing, but production frontends should
    # reserve a static address so DNS never has to chase a new VIP. Required
    # for Private Service Connect rules to Google APIs. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ip_address = optional(string, "")

    # The IP protocol this rule matches (default TCP). All proxy-based load
    # balancers and Private Service Connect use TCP; UDP, ESP, AH, SCTP, and
    # ICMP exist for passthrough load balancing and protocol forwarding.
    # L3_DEFAULT — regional rules only — forwards every IP protocol at once
    # (the passthrough Network Load Balancer's multi-protocol form); it
    # requires all_ports and a backend service whose protocol is UNSPECIFIED.
    # Immutable.
    ip_protocol = optional(string)

    # IP version for the auto-assigned ephemeral address (IPV4 or IPV6; GCP
    # default IPV4). Only meaningful when ip_address is omitted — a referenced
    # static address already fixes the version. Immutable.
    ip_version = optional(string, "")

    # Which load balancer family this frontend belongs to (default EXTERNAL
    # on both scopes: the classic global external ALB, or the external
    # passthrough Network Load Balancer on a regional rule). EXTERNAL_MANAGED
    # is the envoy-based external ALB (global, or regional with region set);
    # INTERNAL_MANAGED is the internal ALB (cross-region on a global rule,
    # regional with region set); INTERNAL — regional rules only — is the
    # internal passthrough Network Load Balancer, which names a
    # backend_service instead of a target; INTERNAL_SELF_MANAGED — global
    # rules only — is Traffic Director / service mesh; NONE (sent to GCP as
    # an empty scheme) is Private Service Connect: to Google APIs on a global
    # rule, to a producer's service attachment on a regional one. The scheme
    # must match the family the target's backend services were created for.
    # Both engines send EXTERNAL explicitly when this is left empty, on both
    # scopes, so an unset scheme means the same thing wherever the rule
    # lives. Immutable — except the EXTERNAL → EXTERNAL_MANAGED canary
    # migration driven by external_managed_backend_bucket_migration_state.
    load_balancing_scheme = optional(string)

    # The port or contiguous port range ("443" or "8080-8090") this rule
    # matches. Requires a TCP/UDP/SCTP protocol. Proxy-based load balancers
    # accept only specific ports (80/8080/443 for HTTP(S)); two external
    # rules on the same IP+protocol cannot overlap ranges — which is exactly
    # how the port-80 redirect rule and the port-443 serving rule share one
    # VIP. Not used by Private Service Connect rules. On a regional rule, at
    # most one of port_range, ports, and all_ports is set. Immutable.
    port_range = optional(string, "")

    # Up to five individual ports or ranges ("80", "443", "8080-8090") this
    # rule matches — the passthrough Network Load Balancer's form (internal
    # passthrough, backend-service-based external passthrough, internal
    # protocol forwarding), where the ports need not be contiguous. Requires
    # a TCP, UDP, or SCTP protocol; regional rules only; mutually exclusive
    # with port_range and all_ports. Immutable.
    ports = optional(list(string), [])

    # Forward packets addressed to ANY port — and packets lacking a
    # destination port, such as UDP fragments after the first — to the
    # backends. The passthrough Network Load Balancer's form for services
    # that listen on many ports or for protocol forwarding; required by the
    # L3_DEFAULT protocol. Requires TCP, UDP, SCTP, or L3_DEFAULT; regional
    # rules only; mutually exclusive with port_range and ports. Immutable.
    all_ports = optional(bool, false)

    # The VPC network this frontend belongs to. Reference a GcpVpcNetwork
    # resource or provide a network self-link. Used by the internal-facing
    # schemes (INTERNAL, INTERNAL_MANAGED, INTERNAL_SELF_MANAGED), by Private
    # Service Connect (NONE — required, on both scopes), and by the REGIONAL
    # external ALB (EXTERNAL_MANAGED with region set, whose proxy-only subnet
    # lives in this network); the global external load balancers and the
    # external passthrough Network Load Balancer live on Google's edge and
    # reject it. If omitted where applicable, GCP uses the default network.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # The subnetwork the load-balanced IP belongs to, for internal load
    # balancing and for IPv6 external passthrough Network Load Balancers.
    # Optional when the network is auto-mode; required when it is custom-mode
    # (and for an IPv6 external passthrough rule). Reference a GcpSubnetwork
    # resource or provide a subnetwork self-link. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnetwork = optional(string, "")

    # Networking tier. PREMIUM (Google's global backbone; the default when
    # empty) on both scopes; STANDARD (regional ISP transit, cheaper egress)
    # only on a regional rule — a global rule is PREMIUM by definition. If
    # ip_address references a reserved address, the tiers must match.
    # Immutable.
    network_tier = optional(string, "")

    # Traffic Director metadata filters: restrict which xDS clients receive
    # this forwarding rule's configuration, by matching labels the clients
    # present in their node metadata. Only applies to INTERNAL_SELF_MANAGED
    # frontends, which are global rules. Filters set here can be overridden
    # by the URL map's own metadata filters. Immutable.
    metadata_filters = optional(list(object({
      # How the labels combine: MATCH_ALL (every label must match) or MATCH_ANY
      # (at least one).
      filter_match_criteria = string

      # The xDS node metadata labels to match against (1-64 entries).
      filter_labels = list(object({
        # Label name (1-1024 characters).
        name = string

        # The value the label must match (up to 1024 characters).
        value = string
      }))
    })), [])

    # Register this Private Service Connect frontend in Service Directory so
    # VPC workloads can discover the private endpoint by name. Used by PSC
    # rules (scheme NONE): a Google-APIs bundle rule names the registration
    # region, a regional consumer-endpoint rule names the service. Immutable.
    service_directory_registration = optional(object({
      # Service Directory namespace to register the forwarding rule under. If
      # omitted, GCP registers it under a Google-managed namespace.
      namespace = optional(string, "")

      # Service Directory region to register this global rule under (GCP
      # default us-central1). All PSC-for-Google-APIs rules on one network
      # should use the same region. Global rules only — a regional rule is
      # registered in its own region.
      service_directory_region = optional(string, "")

      # Service Directory service to register this regional consumer endpoint
      # under, inside the namespace. Regional rules only.
      service = optional(string, "")
    }))

    # Let clients in EVERY region reach this internal load balancer, instead
    # of only clients in the rule's own region (Google's default). For the
    # internal passthrough Network Load Balancer (scheme INTERNAL) and for
    # internal target-instance forwarding. Regional rules only. Mutable.
    allow_global_access = optional(bool, false)

    # Let clients in every region reach this Private Service Connect consumer
    # endpoint (a regional rule with scheme NONE whose target is a service
    # attachment), instead of only clients in the endpoint's region. Regional
    # rules only. Mutable.
    allow_psc_global_access = optional(bool, false)

    # A DNS label (RFC 1035, 1-63 characters) prepended to the internal
    # passthrough Network Load Balancer's service name, giving the VIP a
    # stable internal name like
    # <service_label>.<name>.il4.<region>.lb.<project>.internal (exposed as
    # the service_name output). INTERNAL scheme, regional rules only.
    # Immutable.
    service_label = optional(string, "")

    # Mark this internal passthrough Network Load Balancer as a Packet
    # Mirroring collector: mirrored traffic is delivered to its backends, and
    # to prevent mirroring loops those backends are never mirrored themselves
    # even when a PacketMirroring rule applies to them. INTERNAL scheme,
    # regional rules only. Immutable.
    is_mirroring_collector = optional(bool, false)

    # Bring your own IPv6 range: the PublicDelegatedPrefix (a sub-PDP in
    # EXTERNAL_IPV6_FORWARDING_RULE_CREATION mode) the external passthrough
    # Network Load Balancer's IPv6 address is drawn from, as a resource URL or
    # partial path (projects/{p}/regions/{r}/publicDelegatedPrefixes/{name}).
    # Regional rules only. Immutable.
    ip_collection = optional(string, "")

    # Recreate this Private Service Connect consumer endpoint when Google
    # reports its connection CLOSED (the producer removed or rejected it) —
    # the engines otherwise leave a closed endpoint in place until it is
    # changed by hand. Default false. Regional PSC rules only. Mutable.
    recreate_closed_psc = optional(bool, false)

    # Forward only traffic whose SOURCE address matches one of these IP
    # addresses ("1.2.3.4") or CIDR ranges ("1.2.3.0/24"), up to 64 — a
    # coarse allowlist at the VIP for the external passthrough Network Load
    # Balancer. Regional rules with scheme EXTERNAL only. Immutable.
    source_ip_ranges = optional(list(string), [])

    # Skip the DNS zone Google normally auto-creates for a Private Service
    # Connect Google-APIs frontend (the zone that maps googleapis.com names to
    # the private VIP). Set true when you manage that DNS yourself. Only
    # meaningful for PSC rules (scheme NONE). Immutable.
    no_automate_dns_zone = optional(bool, false)

    # Labels to organize and bill this forwarding rule (e.g. env, team,
    # cost-center). Keys and values follow GCP label rules. Mutable.
    labels = optional(map(string), {})

    # Canary state for migrating this frontend's backend BUCKETS from the
    # classic EXTERNAL scheme to EXTERNAL_MANAGED without recreating the VIP:
    # PREPARE stages the migration, TEST_BY_PERCENTAGE shifts the fraction set
    # in external_managed_backend_bucket_migration_testing_percentage, and
    # TEST_ALL_TRAFFIC must be reached before flipping load_balancing_scheme
    # to EXTERNAL_MANAGED. Roll back by walking the states in reverse.
    # Mutable.
    external_managed_backend_bucket_migration_state = optional(string, "")

    # Percentage (0-100) of backend-bucket requests served by the envoy-based
    # Global external ALB during a TEST_BY_PERCENTAGE canary migration. Only
    # meaningful with external_managed_backend_bucket_migration_state
    # TEST_BY_PERCENTAGE. Mutable.
    external_managed_backend_bucket_migration_testing_percentage = optional(number, 0)

    # Deletion policy for the global forwarding rule — what happens on
    # destroy:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the frontend is deleted; the reserved IP it served stays
    #                reserved (its own kind's policy governs it), but traffic
    #                to that IP stops routing the moment the rule is gone
    #   "PREVENT" -- destroy FAILS; protects the entry point of a live load
    #                balancer chain
    #   "ABANDON" -- the rule is removed from management but keeps serving
    #                traffic in GCP
    deletion_policy = optional(string, "")
  })
}
