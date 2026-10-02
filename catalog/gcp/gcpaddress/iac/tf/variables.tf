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
  description = "GcpAddress specification"
  type = object({
    # The GCP project in which to create this regional address reservation.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Example: "my-prod-project-123"
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the regional address resource in GCP.
    # Must be 1-63 characters, lowercase letters, numbers, or hyphens.
    # Must start with a lowercase letter and end with a letter or number.
    # Example: "nat-external-ip", "ilb-vip"
    address_name = string

    # The GCP region in which to reserve this address (e.g. "us-central1").
    # Immutable: a regional address cannot move between regions.
    region = string

    # The static IP address to reserve. If omitted, GCP assigns an address
    # automatically. For INTERNAL VPC_PEERING addresses this is the start of
    # the reserved CIDR range.
    address = optional(string, "")

    # The type of address to reserve.
    # EXTERNAL reserves a public IP address (default). INTERNAL reserves a
    # private IP within a VPC network or subnetwork.
    address_type = optional(string)

    # Human-readable description of this address reservation.
    # Example: "Static IP for Cloud NAT in us-central1"
    description = optional(string, "")

    # The IP version for this address. Defaults to IPV4.
    ip_version = optional(string)

    # The VPC network for internal address reservations with VPC_PEERING or
    # IPSEC_INTERCONNECT purpose. Accepts a network name or full self-link URL.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # The subnetwork for internal address reservations with GCE_ENDPOINT or
    # DNS_RESOLVER purpose. Accepts a subnetwork name or full self-link URL.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnetwork = optional(string, "")

    # The network tier for EXTERNAL addresses only. PREMIUM (default) or
    # STANDARD. Must not be set for INTERNAL addresses — internal traffic
    # always uses Premium tier.
    network_tier = optional(string, "")

    # The prefix length of the IP range to reserve (8-29).
    # Used for INTERNAL VPC_PEERING or IPSEC_INTERCONNECT ranges.
    prefix_length = optional(number)

    # The purpose of this INTERNAL address reservation.
    # GCE_ENDPOINT — VM instances, alias IP ranges, or similar endpoints.
    # SHARED_LOADBALANCER_VIP — internal load balancer VIP shared across
    # backends.
    # VPC_PEERING — reserves a CIDR range for VPC network peering.
    # IPSEC_INTERCONNECT — reserves a range for HA VPN over Cloud Interconnect.
    # DNS_RESOLVER — DNS resolver endpoint address.
    # Leave empty for standard EXTERNAL address reservations.
    # PRIVATE_SERVICE_CONNECT is global-only — use GcpGlobalAddress instead.
    purpose = optional(string, "")

    # The endpoint type for external IPv6 addresses: VM or NETLB.
    # Determines whether the reserved IPv6 address can be used by a VM
    # instance or a network load balancer after reservation.
    ipv6_endpoint_type = optional(string, "")

    # User labels attached to the reserved address, merged with Planton's
    # platform labels (which win on key conflicts). The one mutable surface on
    # this resource — every other change destroys and re-reserves the address.
    labels = optional(map(string), {})

    # Source of externally provisioned (BYOIP) addresses: a
    # PublicDelegatedPrefix in full or partial URL form, e.g.
    # "projects/{project}/regions/{region}/publicDelegatedPrefixes/{pdp-name}".
    # An IPv4 PDP must support enhanced IPv4 allocations; an IPv6 PDP must be
    # in EXTERNAL_IPV6_FORWARDING_RULE_CREATION mode. Only meaningful for
    # EXTERNAL addresses. Create-time only: changing it re-reserves.
    ip_collection = optional(string, "")

    # Deletion policy for the reserved address — what happens on destroy:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the reservation is released; the IP returns to Google's
    #                pool (an external static IP is gone for good)
    #   "PREVENT" -- destroy FAILS; protects an IP that DNS records or
    #                allow-lists outside GCP may still point at
    #   "ABANDON" -- the reservation is removed from management but stays
    #                reserved (and billed while unattached) in GCP
    deletion_policy = optional(string, "")
  })
}
