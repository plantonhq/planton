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
  description = "GcpVpcNetwork specification"
  type = object({
    # The GCP project that owns this VPC network.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the network.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Whether to use auto subnet mode (true) or custom subnet mode (false).
    # **Default:** false (custom mode). Auto mode is not recommended for production.
    auto_create_subnetworks = optional(bool, false)

    # Dynamic routing mode for the VPC's Cloud Routers: REGIONAL or GLOBAL.
    # **Default:** REGIONAL (Cloud Router advertises routes only in one region).
    # Use GLOBAL only for multi-region routing needs.
    routing_mode = optional(string)

    # Name of the VPC network to create in GCP.
    # Must be 1-63 characters, lowercase letters, numbers, or hyphens.
    # Must start with a lowercase letter and end with a lowercase letter or number.
    # Example: "my-vpc-network", "prod-network"
    network_name = string

    # Human-readable description of the network. Immutable: changing it
    # destroys and recreates the network.
    description = optional(string, "")

    # Maximum Transmission Unit in bytes (1300–8896). Default 1460.
    # Jumbo frames (up to 8896) apply within the VPC; traffic to the internet
    # or other VPCs may still be subject to lower effective MTUs.
    mtu = optional(number)

    # Enable ULA internal IPv6 on this network. When true, GCP assigns a /48
    # from the fd20::/20 ULA prefix (or uses internal_ipv6_range when set).
    enable_ula_internal_ipv6 = optional(bool, false)

    # When enabling ULA internal IPv6, optionally specify the /48 range from
    # fd20::/20. Immutable. If omitted, GCP allocates one automatically.
    internal_ipv6_range = optional(string, "")

    # Order in which firewall policies and classic firewall rules are evaluated.
    # Default AFTER_CLASSIC_FIREWALL.
    network_firewall_policy_enforcement_order = optional(string, "")

    # Full or partial URL of a network profile to apply at creation time.
    # Immutable.
    network_profile = optional(string, "")

    # BGP best-path selection settings for the network.
    bgp_best_path_selection = optional(object({
      # The BGP best selection algorithm: LEGACY (default) or STANDARD.
      mode = optional(string, "")

      # When mode is STANDARD, enables comparison of MED across routes with
      # different neighbor ASNs.
      always_compare_med = optional(bool, false)

      # When mode is STANDARD, controls inter-regional cost behavior in the BPS
      # algorithm: DEFAULT or ADD_COST_TO_MED.
      inter_region_cost = optional(string, "")
    }))

    # When true, default routes (0.0.0.0/0) are not created automatically.
    # Immutable.
    delete_default_routes_on_create = optional(bool, false)

    # Resource Manager tags bound to the network for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
    # Create-time only: changing them later replaces the network.
    resource_manager_tags = optional(map(string), {})

    # Deletion policy for the VPC network — what happens when this resource
    # is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the network is deleted (GCP refuses while subnets,
    #                peerings, or attached resources remain in it)
    #   "PREVENT" -- destroy FAILS; protects the network every subnet,
    #                route, and peering in it depends on
    #   "ABANDON" -- the network is removed from management but left
    #                serving in GCP
    deletion_policy = optional(string, "")
  })
}
