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
  description = "AwsMemcachedElasticache specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Memcached engine version to deploy. Uses three-part versioning:
    # "1.6.22", "1.6.17", "1.5.16", etc. Leave empty to use the AWS default —
    # a versionless manifest never goes stale.
    # Transit encryption requires version 1.6.12 or later.
    engine_version = optional(string, "")

    # ElastiCache node type. Determines CPU, memory, and network capacity.
    # Examples: "cache.t3.micro" (dev), "cache.r7g.large" (production).
    # Changing node_type forces cluster recreation — Memcached does not support
    # vertical scaling in-place.
    node_type = string

    # Number of cache nodes in the cluster. Memcached distributes keys across
    # all nodes via consistent hashing. Range: 1–40. Default: 1.
    num_cache_nodes = optional(number, 0)

    # AZ distribution mode. "single-az" places all nodes in one AZ (default).
    # "cross-az" distributes nodes across multiple AZs for resilience.
    # cross-az requires num_cache_nodes > 1.
    az_mode = optional(string, "")

    # Port on which the cluster accepts connections. Default: 11211.
    # This is a ForceNew attribute — changing it destroys and recreates the
    # cluster.
    port = optional(number)

    # Enable encryption in transit (TLS) for all client connections.
    # Requires Memcached engine version 1.6.12 or later. Attempting to enable
    # this on earlier versions will result in an AWS API error.
    # Note: Memcached does NOT support encryption at rest.
    transit_encryption_enabled = optional(bool, false)

    # Subnet IDs for the ElastiCache subnet group. Provide subnets in at least
    # two AZs when using cross-az mode. A subnet group is created automatically
    # from these subnets. Mutually exclusive with `subnet_group_name`.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Name of an EXISTING ElastiCache subnet group to place the cluster in,
    # instead of building one from `subnet_ids`. Bring-your-own for
    # organizations that manage subnet groups centrally. ForceNew — changing
    # the subnet group replaces the cluster.
    subnet_group_name = optional(string, "")

    # VPC security groups to attach to the cluster nodes. Controls network-level
    # access to the Memcached endpoint. Since Memcached has no authentication,
    # security groups are the primary access control mechanism.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # IP addressing for the cluster's network. Values: "ipv4" (default),
    # "ipv6", "dual_stack". ForceNew — changing the network type replaces the
    # cluster. Dual-stack requires subnets with both IPv4 and IPv6 CIDRs.
    network_type = optional(string, "")

    # Which address family DNS discovery returns to clients. Values: "ipv4",
    # "ipv6". Only meaningful alongside a dual-stack `network_type`; updates
    # in place, letting clients migrate address families without replacing
    # the cluster.
    ip_discovery = optional(string, "")

    # Parameter group family for custom parameters. Required when `parameters`
    # is provided. Examples: "memcached1.6", "memcached1.5", "memcached1.4".
    parameter_group_family = optional(string, "")

    # Custom cache parameters to apply via a managed parameter group. Common
    # Memcached parameters include: chunk_size, chunk_size_growth_factor,
    # max_simultaneous_connections, binding_protocol. Mutually exclusive with
    # `parameter_group_name`.
    parameters = optional(list(object({
      # Parameter name (e.g., "chunk_size", "binding_protocol").
      name = string

      # Parameter value (e.g., "96", "auto").
      value = string
    })), [])

    # Name of an EXISTING parameter group to use instead of managing
    # parameters here. Bring-your-own for organizations that share one tuned
    # group across many caches. Mutually exclusive with `parameters`.
    parameter_group_name = optional(string, "")

    # Weekly maintenance window in UTC. Format: "ddd:hh24:mi-ddd:hh24:mi".
    # Example: "sun:05:00-sun:06:00". Leave empty for AWS-assigned default.
    maintenance_window = optional(string, "")

    # Apply changes immediately instead of waiting for the next maintenance
    # window. May cause brief downtime for some operations.
    apply_immediately = optional(bool, false)

    # Automatically apply minor engine version upgrades during maintenance
    # windows. AWS enables this by default: leave unset to keep the default,
    # set false to pin the running minor version explicitly, set true to pin
    # the opt-in. Presence matters — unset is forwarded to AWS as "decide",
    # never as false.
    auto_minor_version_upgrade = optional(bool)

    # SNS topic ARN for cluster event notifications (node additions, removals,
    # maintenance events, etc.).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    notification_topic_arn = optional(string, "")

    # Preferred Availability Zones for the cache nodes. When provided, the list
    # length must match num_cache_nodes. Nodes are placed in the specified AZs
    # in order. Leave empty for AWS-managed AZ distribution. Mutually exclusive
    # with `availability_zone` (the single-AZ pin).
    preferred_availability_zones = optional(list(string), [])

    # Pin ALL cache nodes to one Availability Zone (e.g. "us-west-2a") —
    # AZ-local latency for a client fleet that lives in a single zone.
    # ForceNew — changing the pin replaces the cluster. Leave empty for an
    # AWS-chosen AZ. Mutually exclusive with `preferred_availability_zones`
    # (per-node placement) and with az_mode "cross-az" — a single-AZ pin IS
    # single-az mode.
    availability_zone = optional(string, "")
  })
}
