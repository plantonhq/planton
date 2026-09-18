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
  description = "AwsRedisElasticache specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Cache engine to use. Redis is the dominant choice; Valkey is the
    # open-source Redis-compatible alternative. Values: "redis", "valkey".
    # Redis <-> Valkey switches apply in place (Valkey is protocol-compatible).
    # Required — unless the group joins a global datastore, where the engine
    # is inherited from the primary and must be left empty.
    engine = optional(string, "")

    # Engine version to deploy. Examples: "7.1", "7.0", "6.2" for Redis;
    # "7.2", "8.0" for Valkey. Redis 6+ and Valkey use major.minor ("7.1");
    # Redis 5 and earlier use full three-part versions ("5.0.6"). Leave empty
    # to use the provider default.
    # Must be left empty when joining a global datastore (inherited).
    engine_version = optional(string, "")

    # Human-readable description for the replication group. Required by AWS.
    description = string

    # ElastiCache node type. Determines CPU, memory, and network capacity.
    # Examples: "cache.t3.micro" (dev), "cache.r7g.large" (production),
    # "cache.r6gd.xlarge" (data tiering). Required — unless the group joins a
    # global datastore, where the node type is inherited from the primary and
    # must be left empty.
    node_type = optional(string, "")

    # Port on which the cluster accepts connections. Default: 6379.
    # This is a ForceNew attribute — changing it destroys and recreates the cluster.
    port = optional(number)

    # Total number of cache clusters (nodes) in the replication group. This includes
    # the primary and all read replicas. For example, 3 means 1 primary + 2 replicas.
    # Range: 1–6 — AWS's CreateReplicationGroup contract caps a non-clustered
    # group at 1 primary + 5 replicas (the Terraform provider stopped
    # validating this cap in 6.35.0; the spec deliberately mirrors AWS's
    # contract, not the provider's looseness).
    # Mutually exclusive with `num_node_groups`.
    num_cache_clusters = optional(number, 0)

    # Preferred Availability Zones for the cache clusters of a NON-clustered
    # group, in creation order (first entry hosts the primary). When provided,
    # the list length must match `num_cache_clusters`. For per-shard placement
    # in clustered mode use `node_group_configurations` instead — the two are
    # mutually exclusive.
    preferred_cache_cluster_azs = optional(list(string), [])

    # Number of node groups (shards) for Cluster Mode Enabled. Each shard holds a
    # partition of the keyspace. Mutually exclusive with `num_cache_clusters`,
    # and forbidden when joining a global datastore (the primary defines the
    # shard layout).
    num_node_groups = optional(number, 0)

    # Number of read replicas per shard. Range: 0–5 — AWS's
    # CreateReplicationGroup contract ("Valid values are 0 to 5"; the
    # Terraform provider stopped validating the ceiling in 6.35.0 — the spec
    # deliberately mirrors AWS's contract, not the provider's looseness).
    # Only valid when `num_node_groups` is set.
    replicas_per_node_group = optional(number, 0)

    # Per-shard placement for Cluster Mode Enabled — pin each shard's primary
    # and replicas to specific Availability Zones, control its replica count,
    # or assign its keyspace slots. Most clustered deployments leave this
    # empty and let AWS spread shards; reach for it when data locality or an
    # AZ-aligned client topology demands explicit placement. Requires
    # `num_node_groups`; mutually exclusive with `preferred_cache_cluster_azs`.
    # Changing explicit shard placement after creation replaces the group.
    node_group_configurations = optional(list(object({
      # Identifier for the shard. 1–4 digits (e.g. "0001"). Determines ordering;
      # when omitted AWS assigns sequential ids.
      node_group_id = optional(string, "")

      # Availability Zone hosting the shard's primary node.
      primary_availability_zone = optional(string, "")

      # Availability Zones for the shard's replicas, in order. The list length
      # should match replica_count when both are set.
      replica_availability_zones = optional(list(string), [])

      # Number of replicas in this shard. Overrides the group-level
      # replicas_per_node_group for this shard. Range: 0–5.
      replica_count = optional(number, 0)

      # Keyspace slots owned by this shard, as a range or list expression
      # (e.g. "0-5461"). Leave empty for AWS's even distribution — set only
      # when migrating an existing slot layout.
      slots = optional(string, "")
    })), [])

    # Enable automatic failover to a read replica if the primary fails.
    # Requires `num_cache_clusters >= 2` (non-clustered) or `num_node_groups > 0`
    # (clustered mode, where failover is always on).
    automatic_failover_enabled = optional(bool, false)

    # Deploy replicas across multiple Availability Zones for resilience against
    # AZ-level failures. Requires `automatic_failover_enabled` to be true.
    multi_az_enabled = optional(bool, false)

    # Durability mode for the replication group — controls how writes are
    # acknowledged relative to replica propagation. Values: "default",
    # "async", "sync", "disabled". Requires Cluster Mode Enabled
    # (`num_node_groups`) and engine Valkey 9.0 or later; leave empty
    # everywhere else. "sync" trades write latency for zero-data-loss
    # failovers. ForceNew — changing it replaces the group.
    durability = optional(string, "")

    # ID of an existing global replication group (Aurora-style cross-region
    # replication for ElastiCache) this group joins as a SECONDARY. The
    # secondary inherits engine, engine version, node type, encryption
    # settings, and shard layout from the global primary — leave `engine`,
    # `engine_version`, `node_type`, `num_node_groups`, the encryption fields,
    # the parameter-group fields, and the restore sources empty when set.
    # ForceNew — a group cannot change datastore membership in place.
    # Global replication groups themselves are created outside this component;
    # this field is the join path.
    global_replication_group_id = optional(string, "")

    # Subnet IDs for the ElastiCache subnet group. Provide subnets in at least two
    # AZs for multi-AZ deployments. A subnet group is created automatically from
    # these subnets. Mutually exclusive with `subnet_group_name`.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Name of an EXISTING ElastiCache subnet group to place the cluster in,
    # instead of building one from `subnet_ids`. Bring-your-own for
    # organizations that manage subnet groups centrally. ForceNew — changing
    # the subnet group replaces the cluster.
    subnet_group_name = optional(string, "")

    # VPC security groups to attach to the cluster nodes. Controls network-level
    # access to the Redis/Valkey endpoint.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # IP addressing for the cluster's network. Values: "ipv4" (default),
    # "ipv6", "dual_stack". ForceNew — changing the network type replaces the
    # cluster. Dual-stack requires subnets with both IPv4 and IPv6 CIDRs.
    network_type = optional(string, "")

    # Which address family DNS discovery returns to clients. Values: "ipv4",
    # "ipv6". Only meaningful alongside a dual-stack `network_type` (a
    # single-stack cluster has nothing to choose); updates in place, letting
    # clients migrate address families without replacing the cluster.
    ip_discovery = optional(string, "")

    # Enable encryption at rest for data stored on disk and in snapshots.
    # Presence matters: leave unset to let AWS apply its engine default,
    # set true/false to pin it explicitly. Must be left UNSET when joining a
    # global datastore — the setting is inherited from the primary, and the
    # provider rejects the argument's presence alongside
    # global_replication_group_id. ForceNew — changing it destroys and
    # recreates the cluster.
    at_rest_encryption_enabled = optional(bool)

    # Enable encryption in transit (TLS) for all client connections and
    # replication traffic. Strongly recommended for production; required for
    # AUTH tokens and IAM-authenticated RBAC users. Presence matters: leave
    # unset to let AWS apply its default (disabled), set true/false to pin it
    # explicitly. Must be left UNSET when joining a global datastore — the
    # setting is inherited from the primary, and the provider rejects the
    # argument's presence alongside global_replication_group_id.
    transit_encryption_enabled = optional(bool)

    # TLS enforcement mode. "preferred" allows both TLS and non-TLS connections
    # (useful during migration); "required" enforces TLS for all connections.
    # Only valid when `transit_encryption_enabled` is true.
    transit_encryption_mode = optional(string, "")

    # Customer-managed KMS key for at-rest encryption. When set, ElastiCache uses
    # this key instead of the AWS-managed key. ForceNew — changing this destroys
    # and recreates the cluster.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Redis AUTH token (password) for client authentication — the legacy
    # single-shared-credential model. Requires `transit_encryption_enabled` to
    # be true. 16–128 printable characters. Mutually exclusive with
    # `user_group_ids`; prefer RBAC user groups for new deployments.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    auth_token = optional(string, "")

    # How an auth-token CHANGE is applied to the running cluster. Values:
    # "ROTATE" (old and new tokens both work until the rotation completes —
    # zero-downtime), "SET" (the new token replaces the old immediately),
    # "DELETE" (remove the token entirely and turn AUTH off — the migration
    # step from AUTH to RBAC user groups). ROTATE and SET require
    # `auth_token`; DELETE requires it to be ABSENT (the provider rejects a
    # token alongside DELETE — you are removing it).
    auth_token_update_strategy = optional(string, "")

    # RBAC user groups controlling fine-grained access. Each group carries
    # users with specific command and key permissions
    # (AwsElasticacheUser/AwsElasticacheUserGroup) — AWS's recommended
    # production authentication model. Mutually exclusive with `auth_token`.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    user_group_ids = optional(list(string), [])

    # S3 ARNs of RDB snapshot files to seed the new replication group from —
    # the migration path from self-managed Redis (upload the RDB to S3, point
    # here). ForceNew and create-time-only; mutually exclusive with
    # `snapshot_name`.
    snapshot_arns = optional(list(string), [])

    # Name of an existing ElastiCache snapshot to restore into the new
    # replication group — the clone-from-backup path. ForceNew and
    # create-time-only; mutually exclusive with `snapshot_arns`.
    snapshot_name = optional(string, "")

    # Weekly maintenance window in UTC. Format: "ddd:hh24:mi-ddd:hh24:mi".
    # Example: "sun:05:00-sun:06:00". Leave empty for AWS-assigned default.
    maintenance_window = optional(string, "")

    # Number of days to retain automatic snapshots before deletion. 0 disables
    # automatic snapshots. Range: 0–35.
    snapshot_retention_limit = optional(number, 0)

    # Daily snapshot window in UTC. Format: "hh24:mi-hh24:mi".
    # Example: "03:00-04:00". Leave empty for AWS-assigned default.
    snapshot_window = optional(string, "")

    # Identifier for the final snapshot taken when the cluster is deleted. If not
    # provided, no final snapshot is created.
    final_snapshot_identifier = optional(string, "")

    # Apply changes immediately instead of waiting for the next maintenance window.
    # May cause brief downtime for some operations.
    apply_immediately = optional(bool, false)

    # Parameter group family for custom parameters. Required when `parameters` is
    # provided. Examples: "redis7", "redis6.x", "valkey7", "valkey8".
    parameter_group_family = optional(string, "")

    # Custom cache parameters to apply via a managed parameter group. Common
    # examples: maxmemory-policy, timeout, tcp-keepalive. Mutually exclusive
    # with `parameter_group_name`.
    parameters = optional(list(object({
      # Parameter name (e.g., "maxmemory-policy", "timeout").
      name = string

      # Parameter value (e.g., "volatile-lru", "300").
      value = string
    })), [])

    # Name of an EXISTING parameter group to use instead of managing
    # parameters here. Bring-your-own for organizations that share one tuned
    # group across many caches. Mutually exclusive with `parameters`.
    # Note: Cluster Mode Enabled requires a family ".cluster.on" group.
    parameter_group_name = optional(string, "")

    # Log delivery configurations for slow-log and/or engine-log. At most 2
    # entries — one per log type. Logs can be delivered to CloudWatch Logs or
    # Kinesis Data Firehose.
    log_delivery_configurations = optional(list(object({
      # Type of destination. Values: "cloudwatch-logs", "kinesis-firehose".
      destination_type = string

      # Destination identifier. For CloudWatch Logs: the log group name.
      # For Kinesis Firehose: the delivery stream name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination = string

      # Log serialization format. Values: "text", "json".
      log_format = string

      # Type of log to deliver. Values: "slow-log" (commands exceeding slowlog
      # threshold), "engine-log" (engine-level diagnostic output).
      log_type = string
    })), [])

    # SNS topic ARN for cluster event notifications (failover, maintenance,
    # configuration changes, etc.).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    notification_topic_arn = optional(string, "")

    # Automatically apply minor engine version upgrades during maintenance
    # windows. AWS enables this by default: leave unset to keep the default,
    # set false to pin the running minor version explicitly, set true to pin
    # the opt-in. Presence matters — unset is forwarded to AWS as "decide",
    # never as false.
    auto_minor_version_upgrade = optional(bool)

    # Enable data tiering — automatically moves less-frequently-accessed data to
    # SSD storage for up to 5x more data per node. Only available on r6gd node
    # types. ForceNew — cannot be changed after creation.
    data_tiering_enabled = optional(bool, false)

    # Cluster-mode migration setting. Values: "enabled", "compatible",
    # "disabled". "compatible" runs a non-clustered group in cluster-mode-
    # compatible form so clients can migrate to the cluster protocol before
    # the topology actually shards — the online migration path from
    # non-clustered to clustered. Leave empty to let the topology fields
    # decide (the common case).
    cluster_mode = optional(string, "")
  })
}
