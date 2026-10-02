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
  description = "AwsMemorydbCluster specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Database engine to run. Redis OSS is the long-standing choice; Valkey is
    # the open-source, Linux-Foundation-governed fork with lower per-node
    # pricing on AWS. Values: "redis", "valkey". Updates in place — AWS
    # supports switching a Redis cluster to Valkey (never the reverse:
    # downgrades are not supported).
    engine = string

    # Engine version to deploy. Examples: "7.1", "7.0", "6.2" for Redis;
    # "7.2", "7.3" for Valkey. Leave empty to let AWS pick the default for
    # the engine. Upgrades apply in place; downgrades are not supported.
    # One-way once applied: removing the field keeps the running version
    # (the provider adopts the live value rather than reverting) — change
    # it by naming the new version explicitly.
    engine_version = optional(string, "")

    # Human-readable description shown in the AWS console. Leave empty for
    # none — the modules always send an explicit value so the two IaC engines
    # never inject their own differing "Managed by ..." defaults.
    description = optional(string, "")

    # MemoryDB node type. Determines CPU, memory, and network capacity of every
    # node in the cluster. Examples: "db.t4g.small" (dev),
    # "db.r7g.large" (production), "db.r6gd.xlarge" (required for data
    # tiering). Updates in place — AWS performs a rolling vertical scale.
    node_type = string

    # Port on which the cluster accepts connections. Default: 6379.
    # ForceNew — changing it destroys and recreates the cluster.
    port = optional(number)

    # Number of shards (data partitions) in the cluster. Each shard holds a
    # portion of the keyspace and has its own primary. Default: 1. Scales in
    # place (resharding redistributes slots online).
    num_shards = optional(number)

    # Number of read replicas per shard. Range: 0–5. Default: 1 (each shard
    # has 1 primary + 1 replica = 2 nodes). Multi-AZ durability comes from the
    # transaction log even at 0 replicas, but failover is fastest with at
    # least one replica. Scales in place.
    num_replicas_per_shard = optional(number)

    # The Access Control List the cluster authenticates against — MemoryDB's
    # only authentication model. Reference an AwsMemorydbAcl for
    # per-application users, or set the literal value "open-access" (the
    # built-in allow-everything ACL, no authentication) for development.
    # Required: stating open access explicitly is deliberate — an invisible
    # default that grants unauthenticated access has no place in a manifest.
    # Updates in place. Note AWS's own coupling: a cluster with `tls_enabled:
    # false` only accepts the "open-access" ACL (rejected at create otherwise).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    acl_name = string

    # Subnet IDs for a module-managed MemoryDB subnet group. Provide subnets in
    # at least two AZs for multi-AZ resilience; a subnet group named after the
    # cluster is created from them. Mutually exclusive with
    # `subnet_group_name`. When BOTH are omitted, AWS falls back to the
    # account's "default" subnet group — which only exists in accounts with a
    # default VPC, so production manifests set one of the two arms.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # Name of an existing MemoryDB subnet group to place the cluster in —
    # the bring-your-own arm, mutually exclusive with `subnet_ids`. ForceNew:
    # the subnet-group choice is fixed at create time.
    subnet_group_name = optional(string, "")

    # VPC security groups to attach to the cluster nodes. Controls
    # network-level access to the MemoryDB endpoint. Updates in place, with
    # one AWS quirk: once a cluster has at least one security group, the set
    # can never be emptied again — only swapped.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # IP address type for the cluster's network. Values: "ipv4" (default),
    # "ipv6", "dual_stack". ForceNew — the network type is fixed at create
    # time. The subnets in the subnet group must support the chosen type.
    network_type = optional(string, "")

    # How cluster discovery commands (CLUSTER SLOTS / CLUSTER SHARDS) report
    # node addresses to clients. Values: "ipv4" (default), "ipv6". Setting
    # "ipv6" requires `network_type` "ipv6" or "dual_stack" — on a dual-stack
    # cluster this is the dial that moves client traffic onto IPv6. Updates
    # in place.
    ip_discovery = optional(string, "")

    # Enable TLS for in-transit encryption on all client connections. Default
    # true — and the provider enforces that default itself, so omitting the
    # field is identical to sending true; explicit false is the only way to
    # disable. When false, AWS only accepts the "open-access" ACL (no
    # authentication without encryption). ForceNew — changing it destroys and
    # recreates the cluster. IAM-authenticated users also require TLS.
    tls_enabled = optional(bool)

    # Customer-managed KMS key for at-rest encryption. MemoryDB always
    # encrypts data at rest; this optionally substitutes your own key for the
    # AWS-managed one. ForceNew — choose the key at create time.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_arn = optional(string, "")

    # Weekly maintenance window in UTC, minimum 60 minutes.
    # Format: "ddd:hh24:mi-ddd:hh24:mi". Example: "sun:05:00-sun:06:00".
    # Leave empty for an AWS-assigned window. One-way once applied: removing
    # the field keeps the current window rather than reverting to an
    # AWS-assigned one — change it by naming a new window explicitly.
    maintenance_window = optional(string, "")

    # Number of days to retain automatic snapshots. 0 disables automatic
    # snapshots. Range: 0–35. Updates in place.
    snapshot_retention_limit = optional(number, 0)

    # Daily snapshot window in UTC. Format: "hh24:mi-hh24:mi".
    # Example: "05:00-09:00". Leave empty for an AWS-assigned window.
    # One-way once applied: removing the field keeps the current window
    # rather than reverting to an AWS-assigned one — change it by naming a
    # new window explicitly.
    snapshot_window = optional(string, "")

    # Name of the final snapshot to create when the cluster is deleted. If not
    # provided, the cluster's data is gone when the cluster is. Consumed only
    # at delete time — it never affects the running cluster. Format: 1–255
    # lowercase alphanumeric or hyphen characters, no consecutive hyphens, no
    # trailing hyphen (the provider rejects violations at plan time; the rule
    # below fails them at manifest time, before any engine runs).
    final_snapshot_name = optional(string, "")

    # ARN(s) of RDB snapshot files stored in S3 to seed the new cluster from
    # (the offline-migration path from self-managed Redis). ForceNew — only
    # read at cluster creation. Mutually exclusive with snapshot_name. Each
    # entry must be an S3 object ARN (AWS's CreateCluster contract; the
    # provider itself accepts any ARN shape — a recorded looseness) and must
    # not contain commas (an AWS API constraint).
    snapshot_arns = optional(list(string), [])

    # Name of a MemoryDB snapshot to restore from. ForceNew — only read at
    # cluster creation. Mutually exclusive with snapshot_arns.
    snapshot_name = optional(string, "")

    # Parameter group family for the module-managed parameter group. Required
    # when `parameters` is provided. Examples: "memorydb_redis7",
    # "memorydb_valkey7", "memorydb_redis6".
    parameter_group_family = optional(string, "")

    # Custom engine parameters to apply via a module-managed parameter group
    # (named after the cluster). Common examples: activedefrag,
    # maxmemory-policy. Mutually exclusive with `parameter_group_name`.
    # Removing an entry resets that parameter to its family default.
    parameters = optional(list(object({
      # Parameter name (e.g., "activedefrag", "maxmemory-policy").
      name = string

      # Parameter value (e.g., "yes", "volatile-lru").
      value = string
    })), [])

    # Name of an existing MemoryDB parameter group to attach — the
    # bring-your-own arm, mutually exclusive with the folded `parameters`
    # list. Leave both empty for the family default parameter group.
    # Updates in place (AWS waits for the group to be in-sync).
    parameter_group_name = optional(string, "")

    # Name of an existing MemoryDB multi-region cluster to join, making this
    # regional cluster one of its members (active-active writes across
    # regions). The multi-region cluster is created outside this resource
    # (its name is server-generated with a chosen suffix); joining is
    # ForceNew. Leave empty for a single-region cluster.
    multi_region_cluster_name = optional(string, "")

    # SNS topic for cluster event notifications (failover, maintenance,
    # scaling, configuration changes). Updates in place; clearing it disables
    # notifications.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    sns_topic_arn = optional(string, "")

    # Automatically apply minor engine version upgrades during maintenance
    # windows. Default: true — and the provider enforces that default itself,
    # so omitting the field is identical to sending true; explicit false is
    # the only way to opt out. ForceNew — AWS fixes this posture at create
    # time.
    auto_minor_version_upgrade = optional(bool)

    # Enable data tiering — automatically moves less-frequently-accessed data
    # to local SSD for cost efficiency with large datasets. Only available on
    # db.r6gd.* node types. ForceNew — cannot be changed after creation.
    data_tiering = optional(bool, false)
  })
}
