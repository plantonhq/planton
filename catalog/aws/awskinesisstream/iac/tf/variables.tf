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
  description = "AwsKinesisStream specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Capacity mode for the stream. This is a fundamental design choice that
    # determines pricing model, scaling behavior, and operational overhead.
    #
    # Valid values:
    # - "PROVISIONED": You manage shard count explicitly. Each shard provides
    #   1 MB/s ingestion and 2 MB/s consumption. Cost-effective for steady,
    #   predictable workloads.
    # - "ON_DEMAND": AWS auto-scales shards to match throughput. Supports up to
    #   200 MB/s write and 400 MB/s read. Best for variable or unpredictable
    #   workloads. No capacity planning required.
    #
    # This field is required. There is no default -- you must make an explicit
    # choice because the two modes have fundamentally different cost and
    # operational characteristics.
    stream_mode = optional(string, "")

    # Number of open shards in the stream. Each shard provides 1 MB/s write
    # (1,000 records/s) and 2 MB/s read capacity.
    #
    # Required when stream_mode is "PROVISIONED" (must be >= 1).
    # Must be 0 (or omitted) when stream_mode is "ON_DEMAND" because AWS
    # manages shard count automatically.
    #
    # Can be updated after creation to scale provisioned streams. AWS uses
    # uniform scaling (UpdateShardCount with UNIFORM_SCALING strategy).
    shard_count = optional(number, 0)

    # Duration in hours that data records remain accessible after being added to
    # the stream. After the retention period expires, records are no longer
    # accessible via GetRecords.
    #
    # Range: 24–8760 (1 day to 365 days). AWS default: 24 hours.
    # Leave at 0 to use the AWS default.
    #
    # Increasing retention is useful for reprocessing scenarios and late-arriving
    # consumers. Note: extended retention (beyond 24h) incurs additional cost.
    retention_period_hours = optional(number, 0)

    # Customer-managed KMS key for server-side encryption of data at rest. When
    # set, the IaC modules automatically configure KMS encryption (encryption_type
    # = "KMS"). When absent, encryption is disabled (encryption_type = "NONE").
    #
    # Accepts a KMS key ID, key ARN, alias name (e.g., "alias/aws/kinesis"), or
    # alias ARN. Also accepts a reference to an AwsKmsKey resource.
    #
    # Encryption can be enabled or disabled after stream creation (not ForceNew).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Maximum size of a single data record in KiB (kibibytes). Records exceeding
    # this limit are rejected by the PutRecord/PutRecords API.
    #
    # Range: 1024–10240 (1 MiB to 10 MiB). AWS default: 1024 (1 MiB).
    # Leave at 0 to use the AWS default.
    #
    # Larger record sizes are useful for aggregated events, rich JSON payloads,
    # or binary data. Note: larger records consume more shard capacity.
    max_record_size_in_kib = optional(number, 0)

    # Shard-level CloudWatch metrics to enable. By default, Kinesis only provides
    # stream-level metrics. Enabling shard-level metrics allows monitoring of
    # individual shard performance, which is critical for identifying hot shards
    # and capacity bottlenecks in production.
    #
    # Valid values (one or more):
    # - "IncomingBytes" -- bytes written per shard
    # - "IncomingRecords" -- records written per shard
    # - "OutgoingBytes" -- bytes read per shard
    # - "OutgoingRecords" -- records read per shard
    # - "WriteProvisionedThroughputExceeded" -- throttled write requests
    # - "ReadProvisionedThroughputExceeded" -- throttled read requests
    # - "IteratorAgeMilliseconds" -- consumer lag per shard
    # - "ALL" -- shorthand that enables every shard-level metric above
    #
    # Enhanced metrics incur additional CloudWatch cost per metric per shard.
    # Leave empty to use stream-level metrics only (no additional cost).
    shard_level_metrics = optional(list(string), [])

    # When true, all registered enhanced fan-out consumers are automatically
    # deregistered before the stream is deleted, preventing deletion errors.
    # When false (default), deleting a stream with active consumers will fail.
    #
    # This is an operational setting that only affects stream deletion. It has
    # no impact on the running stream.
    enforce_consumer_deletion = optional(bool, false)

    # Pre-provisioned warm write throughput in MiB/s for ON_DEMAND streams.
    # On-demand streams normally scale up in response to observed traffic;
    # warm throughput keeps capacity for a known burst level ready in advance,
    # so a sudden spike (a product launch, a scheduled batch) is absorbed
    # without ProvisionedThroughputExceeded throttling while scaling catches
    # up. Billed per MiB/s-hour on top of on-demand data charges.
    #
    # Mutually exclusive with shard_count (warm throughput is meaningless on
    # PROVISIONED streams, where capacity IS the shard count). Leave at 0 to
    # let on-demand scaling manage capacity reactively.
    warm_throughput_mib_ps = optional(number, 0)

    # Resource-based access policy for the stream, as a standard IAM policy
    # document. The primary use is cross-account access: granting another
    # account's principals PutRecord/GetRecords on this stream without role
    # assumption. AWS models this as a separate resource-policy API keyed by
    # the stream ARN; it is folded here because the policy has no identity of
    # its own and follows the stream's lifecycle.
    resource_policy = optional(any)
  })
}
