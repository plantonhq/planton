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
  description = "AwsServerlessElasticache specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Cache engine to use. Values: "redis", "valkey", "memcached".
    # Switching between redis and valkey is an in-place update. Switching
    # to/from memcached forces recreation.
    engine = string

    # Major engine version. Examples: "7", "8" for Redis/Valkey; "1.6" for
    # Memcached. Leave empty to use the provider default for the chosen engine.
    major_engine_version = optional(string, "")

    # Human-readable description of the serverless cache.
    description = optional(string, "")

    # Maximum data storage in GB. AWS auto-scales storage up to this limit.
    # Range: 1–5000. Leave as 0 to use the AWS default for the engine.
    data_storage_max_gb = optional(number, 0)

    # Minimum data storage in GB. AWS guarantees at least this capacity is
    # always provisioned. Range: 1–5000. Leave as 0 to use the AWS default.
    data_storage_min_gb = optional(number, 0)

    # Maximum ElastiCache Processing Units per second. AWS auto-scales compute
    # up to this limit. Range: 1000–15000000. Leave as 0 for AWS default.
    ecpu_max = optional(number, 0)

    # Minimum ElastiCache Processing Units per second. AWS guarantees at least
    # this compute capacity. Range: 1000–15000000. Leave as 0 for AWS default.
    ecpu_min = optional(number, 0)

    # Subnet IDs for the serverless cache's VPC endpoint. The cache creates
    # VPC endpoints in these subnets. ForceNew — changing this destroys and
    # recreates the cache.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = optional(list(string), [])

    # VPC security groups to attach to the serverless cache endpoint.
    # Controls network-level access.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # IP addressing for the cache's VPC endpoints. Values: "ipv4" (default),
    # "ipv6", "dual_stack". ForceNew — changing the network type destroys and
    # recreates the cache. Dual-stack requires subnets with both IPv4 and
    # IPv6 CIDRs.
    network_type = optional(string, "")

    # Customer-managed KMS key ARN for at-rest encryption. When set, ElastiCache
    # Serverless uses this key instead of the AWS-managed key. ForceNew —
    # changing this destroys and recreates the cache.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Daily automatic snapshot time in UTC. Format: "HH:mm" (e.g., "05:00").
    # Only valid for Redis/Valkey engines. Memcached has no persistence.
    daily_snapshot_time = optional(string, "")

    # Number of days to retain automatic snapshots. Range: 0–35. 0 disables
    # snapshots. Only valid for Redis/Valkey engines.
    snapshot_retention_limit = optional(number, 0)

    # ARNs of existing ElastiCache snapshots to seed the new cache from — the
    # migration path from a node-based Redis/Valkey cluster to serverless
    # (snapshot the cluster, restore here). Create-time-only; only valid for
    # Redis/Valkey engines.
    snapshot_arns_to_restore = optional(list(string), [])

    # RBAC user group controlling fine-grained access
    # (AwsElasticacheUser/AwsElasticacheUserGroup). Serverless caches accept
    # exactly one group. Only valid for Redis/Valkey engines — Memcached has
    # no authentication mechanism.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    user_group_id = optional(string, "")
  })
}
