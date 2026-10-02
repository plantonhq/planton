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
  description = "GcpServerlessVpcConnector specification"
  type = object({
    # The GCP project the connector is created in. Accepts a literal project
    # ID or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region the connector lives in, e.g. "us-central1". Immutable. Serverless
    # workloads can only use a connector in their own region.
    region = string

    # Name of the connector in GCP. Immutable. If not specified, defaults to
    # metadata.name. Maximum 25 characters (a GCP limit stricter than most
    # resource names): lowercase letters, digits, and hyphens, starting with
    # a letter and ending with a letter or digit.
    connector_name = optional(string, "")

    # VPC network to attach to, for network placement. Accepts a literal
    # network name or a reference to a GcpVpcNetwork resource. Requires
    # ip_cidr_range; mutually exclusive with subnet. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # Unused /28 range the connector's instances occupy, in CIDR notation
    # (e.g. "10.8.0.0/28") — network placement only. GCP requires exactly a
    # /28; the range must not overlap any existing subnet, peered range, or
    # route in the network. Immutable.
    ip_cidr_range = optional(string, "")

    # Existing /28 subnetwork the connector occupies, for subnet placement.
    # The required mode on Shared VPC (the subnet lives in the host project).
    # Mutually exclusive with network/ip_cidr_range. Immutable.
    subnet = optional(object({
      # Name of the /28 subnetwork the connector occupies (the short name, not
      # a path — GCP resolves it in the connector's region). The subnetwork
      # must be dedicated to this connector: exactly /28, no other workloads.
      # Accepts a literal name or a reference to a GcpSubnetwork resource.
      # Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      name = string

      # Project that owns the subnetwork. Only needed on Shared VPC, where the
      # subnet lives in the host project; defaults to the connector's project.
      project_id = optional(string, "")
    }))

    # Machine type of the connector's forwarding instances. Larger types
    # carry more throughput per instance (f1-micro ~100 Mbps, e2-micro
    # ~200 Mbps, e2-standard-4 ~1 Gbps class). Mutable in place — a
    # fleet-wide capacity lever that needs no replacement.
    machine_type = optional(string, "")

    # Minimum instances kept running (2–9; GCP defaults to 2). Must be
    # strictly lower than max_instances. The connector never scales below
    # this floor — and note the fleet NEVER scales in on its own: after a
    # burst it stays at the high-water mark until the instances are manually
    # reduced. Decreasing this value forces the connector to be REPLACED
    # (brief egress outage); increasing it applies in place.
    min_instances = optional(number)

    # Maximum instances the connector scales to (3–10; GCP defaults to 10).
    # Must be strictly higher than min_instances. Sizes the throughput
    # ceiling: instances × per-instance throughput for the machine type.
    # Decreasing this value forces the connector to be REPLACED (brief
    # egress outage); increasing it applies in place.
    max_instances = optional(number)

    # Deletion policy for the connector — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the connector is deleted (GCP tears down the forwarding
    #                fleet, ~3-5 minutes; serverless egress through it stops)
    #   "PREVENT" -- destroy FAILS; protects the private-egress path every
    #                function and service in the region may depend on
    #   "ABANDON" -- the connector is removed from management but keeps
    #                forwarding traffic in GCP
    deletion_policy = optional(string, "")
  })
}
