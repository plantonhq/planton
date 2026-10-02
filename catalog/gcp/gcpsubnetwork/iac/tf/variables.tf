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
  description = "GcpSubnetwork specification"
  type = object({
    # The GCP project that owns this subnetwork.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the subnetwork.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The parent VPC network, by self-link.
    # Reference a GcpVpcNetwork resource or provide the self-link directly.
    # Immutable: a subnet cannot move between networks.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_self_link = string

    # Name of the subnetwork in GCP. Must be 1-63 characters: lowercase
    # letters, digits, and hyphens; must start with a letter and end with a
    # letter or digit.
    # Immutable: changing it destroys and recreates the subnetwork.
    subnetwork_name = string

    # Region the subnetwork lives in (e.g. "us-central1"). Subnets are
    # regional: workloads in any zone of this region can use it.
    # Immutable: a subnet cannot move between regions.
    region = string

    # Primary IPv4 CIDR range (e.g. "10.10.0.0/20"). Must not overlap any
    # other range in the VPC. Mutable in ONE direction: the range can be
    # expanded in place (e.g. /20 → /18), but shrinking it forces destroy and
    # recreate — plan for growth up front. Leave empty for IPV6_ONLY subnets
    # (which carry no IPv4 range) or when reserved_internal_range supplies
    # the CIDR from a Network Connectivity internal range.
    ip_cidr_range = optional(string, "")

    # What this subnet carries and which workloads should use it — write it
    # for the operator doing IP planning later.
    # Immutable: description updates force recreation on this resource.
    description = optional(string, "")

    # What the subnet is FOR. PRIVATE (the default) is a regular workload
    # subnet. REGIONAL_MANAGED_PROXY reserves address space for the region's
    # Envoy-based load balancers (required before creating a regional internal
    # or regional external Application Load Balancer in the VPC);
    # GLOBAL_MANAGED_PROXY is the cross-region equivalent.
    # PRIVATE_SERVICE_CONNECT backs published PSC services. PEER_MIGRATION
    # stages subnet migration between peered VPCs. PRIVATE_NAT provides source
    # ranges for Private NAT gateways.
    # Immutable in practice: choose the purpose when the subnet is created.
    purpose = optional(string)

    # For REGIONAL_MANAGED_PROXY subnets only: ACTIVE means the region's Envoy
    # proxies allocate addresses from this subnet now; BACKUP holds the subnet
    # ready for a drain-and-swap (promote a BACKUP to ACTIVE to migrate the
    # proxy fleet onto fresh address space). Mutable.
    role = optional(string, "")

    # Secondary IPv4 ranges for alias IPs — the mechanism GKE uses for pod
    # and service IPs (VPC-native clusters), and any workload that needs
    # per-container addresses. Up to 170 per subnet; names are how consumers
    # (e.g. a GKE cluster's ip_allocation_policy) select a range.
    secondary_ip_ranges = optional(list(object({
      # Name of this secondary range, unique within the subnet — the handle
      # consumers use to select it (e.g. a GKE cluster's pods range). 1-63
      # characters, RFC1035. Renaming a range in place is not supported by GCP.
      range_name = string

      # IPv4 CIDR of this secondary range. Must not overlap any other range in
      # the VPC. Size for the consumer: GKE pod ranges are commonly /14-/18,
      # service ranges /20. Alternative to reserved_internal_range.
      ip_cidr_range = optional(string, "")

      # Source this secondary range's CIDR from a pre-planned Network
      # Connectivity internal range instead of a literal ip_cidr_range, e.g.
      # "networkconnectivity.googleapis.com/projects/{project}/locations/global/internalRanges/{name}".
      # Alternative to ip_cidr_range; unlike the subnet's primary range this
      # is mutable in place.
      reserved_internal_range = optional(string, "")
    })), [])

    # Let VMs without external IPs reach Google APIs and services over the
    # subnet's internal addresses. Effectively mandatory for private-only
    # subnets whose workloads pull from GCR/Artifact Registry or call any
    # Google API. Mutable.
    private_ip_google_access = optional(bool, false)

    # IPv6 counterpart of private_ip_google_access, controlling VM-to-Google
    # traffic over IPv6: DISABLE_GOOGLE_ACCESS,
    # ENABLE_OUTBOUND_VM_ACCESS_TO_GOOGLE, or
    # ENABLE_BIDIRECTIONAL_ACCESS_TO_GOOGLE. Leave unset to keep GCP's
    # default (disabled). Mutable.
    private_ipv6_google_access = optional(string, "")

    # The subnet's IP stack: IPV4_ONLY (the default), IPV4_IPV6 (dual-stack —
    # VMs can carry both), or IPV6_ONLY. Dual-stack and IPv6-only subnets
    # also need ipv6_access_type. Mutable between IPV4_ONLY and IPV4_IPV6;
    # moving to or from IPV6_ONLY recreates.
    stack_type = optional(string)

    # Where the subnet's IPv6 addresses are reachable from: EXTERNAL
    # (internet-routable GUAs assigned by Google) or INTERNAL (ULAs, only
    # routable inside the VPC — requires the VPC to have an internal IPv6
    # range enabled). Required for IPV4_IPV6 and IPV6_ONLY subnets.
    # Immutable: the access type cannot change after creation.
    ipv6_access_type = optional(string, "")

    # For EXTERNAL IPv6 subnets: pin a specific /64 external IPv6 prefix
    # (must belong to a range GCP can assign). Leave empty to let Google
    # allocate one — the normal path. Immutable.
    external_ipv6_prefix = optional(string, "")

    # Permit this subnet's CIDR to overlap with routes to destinations
    # OUTSIDE the VPC (e.g. re-using RFC1918 space that a peer or on-prem
    # route also claims). Subnet routes still take precedence; use only when
    # deliberately reclaiming address space. Mutable.
    allow_subnet_cidr_routes_overlap = optional(bool, false)

    # Controls what an EMPTY secondary_ip_ranges list means on update: true
    # sends the empty list (removing all secondary ranges); false (the
    # default) omits it, leaving existing secondary ranges untouched. A
    # safety latch against wiping GKE pod ranges with a partial manifest.
    send_secondary_ip_range_if_empty = optional(bool, false)

    # VPC Flow Logs for this subnet: samples of TCP/UDP flows for network
    # monitoring, forensics, and cost analysis, delivered to Cloud Logging.
    # Unset means flow logs are OFF. Logging every flow at full sampling on a
    # busy subnet is expensive — tune flow_sampling and filter_expr.
    log_config = optional(object({
      # How long flows are aggregated before a log entry is emitted. Longer
      # intervals mean fewer, larger entries (lower cost, coarser timing).
      aggregation_interval = optional(string)

      # Fraction of flows to sample, 0.0-1.0 (GCP default 0.5). 1.0 captures
      # everything — the right choice for security forensics, at real Logging
      # cost on busy subnets; lower it for trend monitoring.
      flow_sampling = optional(number)

      # Which metadata joins each log entry: INCLUDE_ALL_METADATA (the
      # default), EXCLUDE_ALL_METADATA (smallest entries), or CUSTOM_METADATA
      # (only the fields listed in metadata_fields).
      metadata = optional(string)

      # Metadata fields to include when metadata is CUSTOM_METADATA (e.g.
      # "src_instance", "dest_vpc").
      metadata_fields = optional(list(string), [])

      # CEL expression selecting which flows are logged (GCP default "true" =
      # all sampled flows), e.g. restricting to one port:
      # connection.dest_port == 443.
      filter_expr = optional(string, "")

      # Whether flow logs are on. Unset means on: declaring the block has always
      # meant enabling them, and this switch lets a manifest say the opposite
      # out loud.
      enabled = optional(bool)
    }))

    # Source the primary CIDR from a pre-planned Network Connectivity
    # internal range instead of typing a literal ip_cidr_range: the value is
    # the range's resource path prefixed with the API host, e.g.
    # "networkconnectivity.googleapis.com/projects/{project}/locations/global/internalRanges/{name}".
    # The range's CIDR becomes the subnet's primary range — the enterprise
    # path for centrally allocated IP plans. Immutable: changing it destroys
    # and recreates the subnetwork.
    reserved_internal_range = optional(string, "")

    # For INTERNAL IPv6 subnets: pin a specific internal IPv6 prefix (ULA)
    # instead of letting Google allocate one from the VPC's internal IPv6
    # range. Immutable: the prefix cannot change after creation.
    internal_ipv6_prefix = optional(string, "")

    # BYOIP for IPv6: resource path of a PublicDelegatedPrefix (a sub-PDP in
    # EXTERNAL_IPV6_SUBNETWORK_CREATION or INTERNAL_IPV6_SUBNETWORK_CREATION
    # mode) the subnet's IPv6 space is drawn from, e.g.
    # "projects/{project}/regions/{region}/publicDelegatedPrefixes/{name}".
    # Only meaningful on dual-stack or IPv6-only subnets. Mutable.
    ip_collection = optional(string, "")

    # Resource Manager tags bound to the subnetwork at create time, for
    # org-policy conditions and IAM scoping. Keys are "tagKeys/{tag_key_id}"
    # and values "tagValues/{tag_value_id}" (IDs, not short names). Changing
    # tags REPLACES the subnetwork (the provider sends them through a
    # create-only params block) -- plan tag changes deliberately.
    resource_manager_tags = optional(map(string), {})

    # How VMs in this subnet resolve the subnet mask in ARP responses —
    # relevant for appliance/NFV subnets whose workloads inspect ARP.
    # ARP_ALL_RANGES answers for every range in the subnet;
    # ARP_PRIMARY_RANGE answers only for the primary range;
    # the ARP_BROADCAST_* modes answer with the broadcast-domain mask
    # (WITH_LEARNING additionally learns from observed traffic). Leave unset
    # for GCP's default behavior. Immutable: changing it recreates the
    # subnetwork.
    resolve_subnet_mask = optional(string, "")

    # What happens to the subnetwork in GCP when this resource is destroyed.
    #   "DELETE"  -- (GCP's default when unset) the subnet is deleted; GCP
    #                rejects the delete while any VM, connector, or proxy
    #                still holds an address in it
    #   "PREVENT" -- destroy FAILS; protects the address space workloads
    #                and peers may still route to
    #   "ABANDON" -- the subnet is removed from management but stays in GCP
    #                (free at rest; its CIDR stays claimed in the VPC)
    deletion_policy = optional(string, "")
  })
}
