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

    # The target that receives matched traffic. Reference a
    # GcpTargetHttpsProxy (the default) or a GcpTargetHttpProxy resource, or
    # provide a target URI directly — other global targets (target SSL/TCP
    # proxies, target gRPC proxies) attach by self-link until they exist as
    # Planton kinds. For Private Service Connect, pass the literal bundle name
    # "all-apis" or "vpc-sc" (Google APIs) or a service attachment URI
    # (producer services). Required. Mutable: GCP repoints it in place (a
    # dedicated setTarget call), enabling zero-downtime frontend swaps.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    target = string

    # The IP address this rule accepts traffic on. Reference a
    # GcpGlobalAddress resource (its reserved IP), provide a literal IP
    # ("34.120.1.2"), or an address resource URL. When omitted, Google Cloud
    # assigns an ephemeral IP — fine for testing, but production frontends
    # should reserve a static address so DNS never has to chase a new VIP.
    # Required for Private Service Connect rules. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    ip_address = optional(string, "")

    # The IP protocol this rule matches (default TCP). All proxy-based global
    # load balancers and Private Service Connect use TCP; the other protocols
    # exist for protocol forwarding. Immutable.
    ip_protocol = optional(string)

    # IP version for the auto-assigned ephemeral address (IPV4 or IPV6; GCP
    # default IPV4). Only meaningful when ip_address is omitted — a referenced
    # static address already fixes the version. Immutable.
    ip_version = optional(string, "")

    # Which load balancer family this frontend belongs to (default EXTERNAL,
    # the classic global external Application LB). EXTERNAL_MANAGED is the
    # newer envoy-based global external ALB; INTERNAL_MANAGED is the
    # cross-region internal ALB; INTERNAL_SELF_MANAGED is Traffic Director /
    # service mesh; NONE (sent to GCP as an empty scheme) is Private Service
    # Connect. The scheme must match the family the target proxy's backend
    # services were created for. Immutable — except the EXTERNAL →
    # EXTERNAL_MANAGED canary migration driven by
    # external_managed_backend_bucket_migration_state.
    load_balancing_scheme = optional(string)

    # The port or contiguous port range ("443" or "8080-8090") this rule
    # matches. Requires a TCP/UDP/SCTP protocol. Proxy-based global load
    # balancers accept only specific ports (80/8080/443 for HTTP(S)); two
    # external rules on the same IP+protocol cannot overlap ranges — which is
    # exactly how the port-80 redirect rule and the port-443 serving rule
    # share one VIP. Not used by Private Service Connect rules. Immutable.
    port_range = optional(string, "")

    # The VPC network this frontend belongs to. Only used by internal-facing
    # schemes and Private Service Connect (INTERNAL_MANAGED,
    # INTERNAL_SELF_MANAGED, NONE); external load balancers live on Google's
    # edge, not in a VPC. Reference a GcpVpcNetwork resource or provide a network
    # self-link. For PSC rules a network is required. If omitted where
    # applicable, GCP uses the default network. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # The subnetwork the load-balanced IP belongs to, for internal load
    # balancing. Optional when the network is auto-mode; required when it is
    # custom-mode. Reference a GcpSubnetwork resource or provide a subnetwork
    # self-link. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnetwork = optional(string, "")

    # Networking tier. Global forwarding rules only support PREMIUM (Google's
    # global backbone); STANDARD tier exists only on regional forwarding
    # rules. Empty means PREMIUM. If ip_address references a reserved
    # address, the tiers must match. Immutable.
    network_tier = optional(string, "")

    # Traffic Director metadata filters: restrict which xDS clients receive
    # this forwarding rule's configuration, by matching labels the clients
    # present in their node metadata. Only applies to INTERNAL_SELF_MANAGED
    # frontends. Filters set here can be overridden by the URL map's own
    # metadata filters. Immutable.
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
    # VPC workloads can discover the private Google-APIs endpoint by name.
    # Only used by PSC-for-Google-APIs rules (scheme NONE with an "all-apis" /
    # "vpc-sc" target). Immutable.
    service_directory_registration = optional(object({
      # Service Directory namespace to register the forwarding rule under. If
      # omitted, GCP registers it under a Google-managed namespace.
      namespace = optional(string, "")

      # Service Directory region to register this global rule under (GCP
      # default us-central1). All PSC-for-Google-APIs rules on one network
      # should use the same region.
      service_directory_region = optional(string, "")
    }))

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
