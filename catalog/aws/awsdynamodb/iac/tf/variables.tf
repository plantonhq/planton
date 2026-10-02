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
  description = "AwsDynamodb specification"
  type = object({
    # The AWS region the table is created in. Must match the region of
    # any KMS key or Kinesis stream it references. Replicas live in
    # OTHER regions and are listed under replicas.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # How the table is billed and how capacity is managed:
    # "PAY_PER_REQUEST" (on-demand -- pay per read/write, no capacity
    # planning, the recommended default for new tables) or "PROVISIONED"
    # (reserved read/write units, required for reserved-capacity
    # pricing). Empty keeps the AWS default (PROVISIONED) -- which then
    # requires provisioned_throughput, so most manifests set this
    # explicitly. Switching modes on a live table is an in-place update
    # (AWS allows one switch per 24 hours).
    billing_mode = optional(string, "")

    # Attributes referenced by the table key schema and by every index
    # key schema. Only KEY attributes are declared here -- DynamoDB is
    # schemaless for everything else. Required unless the table is
    # created by restore (the source table carries the schema).
    attribute_definitions = optional(list(object({
      # The attribute name.
      name = string

      # The scalar type: "S" (string), "N" (number), or "B" (binary).
      type = string
    })), [])

    # The table's primary key: exactly one HASH (partition) element and
    # at most one RANGE (sort) element, each naming a declared
    # attribute. Create-time immutable -- changing the key schema
    # replaces the table. Required unless the table is created by
    # restore (the source table carries the key schema).
    key_schema = optional(list(object({
      # The declared attribute this key element uses.
      attribute_name = string

      # "HASH" (partition key) or "RANGE" (sort key).
      key_type = string
    })), [])

    # Reserved read/write capacity for the table. Required when the
    # effective billing mode is PROVISIONED; must stay unset for
    # PAY_PER_REQUEST. On provisioned tables the modules enforce this
    # capacity through an Application Auto Scaling target (pinned
    # min = max) so that capacity changes here always land -- and so that
    # adding the autoscaling block later never replaces the table. With
    # autoscaling configured, these values are the initial capacity only.
    provisioned_throughput = optional(object({
      # Reserved read capacity units (one strongly consistent 4 KB read
      # per second each).
      read_capacity_units = optional(number, 0)

      # Reserved write capacity units (one 1 KB write per second each).
      write_capacity_units = optional(number, 0)
    }))

    # Optional ceilings on on-demand consumption -- a spend guardrail
    # for PAY_PER_REQUEST tables. Requests beyond the ceiling are
    # throttled rather than billed. Only meaningful on
    # PAY_PER_REQUEST tables.
    on_demand_throughput = optional(object({
      # Maximum read request units per second. 0 = no ceiling configured;
      # -1 removes a previously-set ceiling.
      max_read_request_units = optional(number, 0)

      # Maximum write request units per second. 0 = no ceiling configured;
      # -1 removes a previously-set ceiling.
      max_write_request_units = optional(number, 0)
    }))

    # Pre-warmed minimum throughput the table can serve instantly,
    # decoupled from billing mode -- for launch events and traffic
    # cliffs where waiting for organic scale-up is not acceptable. AWS
    # only allows warm throughput to INCREASE; lowering it replaces the
    # table. Leave unset to keep AWS's defaults (12,000 reads/s, 4,000
    # writes/s warm).
    warm_throughput = optional(object({
      # Warm read units per second. 0 keeps the AWS default (12,000);
      # when set, AWS requires at least 12,000.
      read_units_per_second = optional(number, 0)

      # Warm write units per second. 0 keeps the AWS default (4,000);
      # when set, AWS requires at least 4,000.
      write_units_per_second = optional(number, 0)
    }))

    # Global secondary indexes: alternate query shapes with their own
    # key schema (including multi-attribute keys), projection, and --
    # on PROVISIONED tables -- their own capacity. Added, modified, and
    # removed in place on a live table (one GSI mutation at a time, an
    # AWS serialization rule).
    global_secondary_indexes = optional(list(object({
      # The index name, unique within the table.
      name = string

      # The index key: 1-4 HASH elements first (multi-attribute partition
      # keys), then 0-4 RANGE elements (multi-attribute sort keys). Every
      # element names a declared attribute. The common case is one HASH
      # and at most one RANGE.
      key_schema = list(object({
        # The declared attribute this key element uses.
        attribute_name = string

        # "HASH" (partition key) or "RANGE" (sort key).
        key_type = string
      }))

      # Which attributes the index carries.
      projection = object({
        # "ALL" (every attribute), "KEYS_ONLY" (table + index keys), or
        # "INCLUDE" (keys plus non_key_attributes). Projecting less makes
        # the index cheaper to store and write; projecting too little forces
        # costly fetch-backs to the table at query time.
        type = string

        # The non-key attributes projected when type is "INCLUDE". These do
        # not need to be declared in attribute_definitions.
        non_key_attributes = optional(list(string), [])
      })

      # Per-index reserved capacity. Required when the table's effective
      # billing mode is PROVISIONED; must stay unset for PAY_PER_REQUEST.
      provisioned_throughput = optional(object({
        # Reserved read capacity units (one strongly consistent 4 KB read
        # per second each).
        read_capacity_units = optional(number, 0)

        # Reserved write capacity units (one 1 KB write per second each).
        write_capacity_units = optional(number, 0)
      }))

      # Per-index on-demand ceilings (PAY_PER_REQUEST tables only).
      on_demand_throughput = optional(object({
        # Maximum read request units per second. 0 = no ceiling configured;
        # -1 removes a previously-set ceiling.
        max_read_request_units = optional(number, 0)

        # Maximum write request units per second. 0 = no ceiling configured;
        # -1 removes a previously-set ceiling.
        max_write_request_units = optional(number, 0)
      }))

      # Per-index pre-warmed throughput; increase-only, like the table's.
      warm_throughput = optional(object({
        # Warm read units per second. 0 keeps the AWS default (12,000);
        # when set, AWS requires at least 12,000.
        read_units_per_second = optional(number, 0)

        # Warm write units per second. 0 keeps the AWS default (4,000);
        # when set, AWS requires at least 4,000.
        write_units_per_second = optional(number, 0)
      }))
    })), [])

    # Local secondary indexes: alternate sort orders sharing the table's
    # partition key. CREATE-TIME ONLY -- LSIs can never be added or
    # removed after the table exists, and their presence permanently
    # caps each item collection at 10 GB. Prefer a GSI unless you need
    # strongly consistent reads on the alternate sort order.
    local_secondary_indexes = optional(list(object({
      # The index name, unique within the table.
      name = string

      # The alternate sort-key attribute (the partition key is always the
      # table's own HASH key). Must be declared in attribute_definitions.
      range_key = string

      # Which attributes the index carries.
      projection = object({
        # "ALL" (every attribute), "KEYS_ONLY" (table + index keys), or
        # "INCLUDE" (keys plus non_key_attributes). Projecting less makes
        # the index cheaper to store and write; projecting too little forces
        # costly fetch-backs to the table at query time.
        type = string

        # The non-key attributes projected when type is "INCLUDE". These do
        # not need to be declared in attribute_definitions.
        non_key_attributes = optional(list(string), [])
      })
    })), [])

    # Time-to-live: DynamoDB deletes items (within ~48h, free of write
    # cost) once the epoch-seconds value in the named attribute passes.
    ttl = optional(object({
      # Turn TTL on.
      enabled = optional(bool, false)

      # The attribute holding the expiry time as epoch seconds. Required
      # when enabled. MAY (and should) stay set when flipping enabled to
      # false: AWS's UpdateTimeToLive call requires the attribute name when
      # DISABLING TTL too, so keeping it is what makes the disable
      # expressible.
      attribute_name = optional(string, "")
    }))

    # Emit an ordered change stream of item modifications, consumable by
    # Lambda event sources and Kinesis. Required (with view type
    # NEW_AND_OLD_IMAGES) when the table has replicas.
    stream_enabled = optional(bool, false)

    # What each stream record carries: "KEYS_ONLY", "NEW_IMAGE",
    # "OLD_IMAGE", or "NEW_AND_OLD_IMAGES". Required when stream_enabled
    # is true; must stay empty when disabled. Global tables require
    # NEW_AND_OLD_IMAGES.
    stream_view_type = optional(string, "")

    # Continuous backups with per-second restore granularity over the
    # recovery window. The insurance policy for fat-finger deletes and
    # bad deploys -- production tables should enable it.
    point_in_time_recovery = optional(object({
      # Turn point-in-time recovery on.
      enabled = optional(bool, false)

      # How many days of history to retain, 1-35. 0 keeps the AWS default
      # (35 days).
      recovery_period_in_days = optional(number, 0)
    }))

    # Encryption at rest. DynamoDB always encrypts with an AWS-owned key
    # even when this is unset; enable this to switch to the AWS-managed
    # aws/dynamodb key or a customer-managed KMS key you control
    # (required for cross-account access patterns and key-rotation
    # policies of your own).
    server_side_encryption = optional(object({
      # Switch from the AWS-owned key to an AWS-managed or
      # customer-managed KMS key.
      enabled = optional(bool, false)

      # The customer-managed KMS key that encrypts the table. Empty (with
      # enabled true) uses the AWS-managed aws/dynamodb key. Reference an
      # AwsKmsKey key_arn output or pass a literal key ARN. Required to be
      # configured when the table is created by cross-region/cross-account
      # restore.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_arn = optional(string, "")
    }))

    # Storage class: "STANDARD" (the default) or
    # "STANDARD_INFREQUENT_ACCESS" (~60% cheaper storage, ~25% costlier
    # reads/writes -- for large, rarely-read tables like audit logs).
    # Empty keeps the AWS default. Switchable in place (twice per 30
    # days per AWS).
    table_class = optional(string, "")

    # Refuse deletion of the table while true. Flip to false (an
    # in-place update) before a genuine teardown.
    deletion_protection_enabled = optional(bool, false)

    # CloudWatch Contributor Insights: per-key access profiling that
    # answers "which partition keys are hot / throttled". Enables at the
    # table level and, optionally, per GSI.
    contributor_insights = optional(object({
      # Enable insights on the table itself.
      enabled = optional(bool, false)

      # What to profile: "ACCESSED_AND_THROTTLED_KEYS" (the AWS default)
      # or "THROTTLED_KEYS" (cheaper -- only throttle diagnostics). Empty
      # keeps the AWS default. Applies to the table and every listed
      # index.
      mode = optional(string, "")

      # Global secondary indexes (by name) that also get insights, in
      # addition to the table.
      gsi_index_names = optional(list(string), [])
    }))

    # A resource-based IAM policy attached to the table -- cross-account
    # access grants without assuming roles. Table-scoped only (stream
    # policies are a separate niche surface).
    resource_policy = optional(object({
      # The policy document, written as native YAML (serialized to JSON for
      # the AWS API). Statements address the table ARN and, for index
      # access, "{table_arn}/index/*".
      policy = any

      # Allow this policy to remove the applying caller's OWN access to the
      # table. AWS refuses such a policy unless this is set -- the guard
      # against locking yourself out. Set it deliberately for lockdown and
      # hand-off policies where the deploying principal is meant to lose
      # access.
      confirm_remove_self_resource_access = optional(bool, false)
    }))

    # A Kinesis Data Stream that receives the table's item-level change
    # data -- the fan-out path for analytics and search-indexing
    # pipelines (independent of DynamoDB Streams and usable alongside
    # it). AWS allows exactly one Kinesis destination per table.
    kinesis_streaming_destination = optional(object({
      # The destination Kinesis Data Stream. Reference an AwsKinesisStream
      # stream_arn output or pass a literal stream ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      stream_arn = string

      # Timestamp precision on emitted records: "MILLISECOND" or
      # "MICROSECOND". Empty keeps the AWS default (MICROSECOND).
      approximate_creation_date_time_precision = optional(string, "")
    }))

    # Global Tables v2: multi-region replicas of this table, each an
    # active read/write endpoint. Requires streams with
    # NEW_AND_OLD_IMAGES. Adding/removing an entry adds/removes that
    # region's replica in place. For Multi-Region Strong Consistency,
    # set consistency_mode STRONG on every replica (exactly two
    # replicas, or one replica plus global_table_witness).
    replicas = optional(list(object({
      # The region the replica lives in. Must differ from the table's own
      # region.
      region_name = optional(string, "")

      # The customer-managed KMS key encrypting THIS replica (each region
      # encrypts independently). Empty uses the AWS-managed aws/dynamodb
      # key in the replica region. Changing it recreates the replica, not
      # the table. Reference an AwsKmsKey key_arn output or pass a literal
      # key ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_arn = optional(string, "")

      # Enable point-in-time recovery on the replica (independent of the
      # source table's setting).
      point_in_time_recovery = optional(bool, false)

      # Refuse deletion of the replica while true.
      deletion_protection_enabled = optional(bool, false)

      # Propagate the table's tags to the replica. One-way: tag changes
      # flow from the table to the replica; replica-side drift is left
      # alone.
      propagate_tags = optional(bool, false)

      # "EVENTUAL" (the default -- classic global tables, last-writer-wins)
      # or "STRONG" (Multi-Region Strong Consistency -- synchronous quorum
      # writes; requires exactly two STRONG replicas, or one STRONG
      # replica plus the witness region).
      consistency_mode = optional(string, "")
    })), [])

    # The witness region of a Multi-Region Strong Consistency global
    # table -- it stores replicated writes to support quorum but serves
    # no reads or writes. Must accompany exactly one replica with
    # consistency_mode STRONG (the cheaper MRSC topology; the
    # alternative is two STRONG replicas and no witness).
    global_table_witness = optional(object({
      # The witness region. Must differ from the table's region and the
      # replica's region.
      region_name = optional(string, "")
    }))

    # Create this table by restoring another table's point-in-time
    # state: the name of the source table in this account and region.
    # Key schema and attributes are inherited from the source. Mutually
    # exclusive with the other restore/import sources.
    restore_source_name = optional(string, "")

    # Create this table by restoring from a source table ARN -- the
    # cross-region / cross-account form of point-in-time restore.
    # Server-side encryption must be configured on the restored table.
    # Mutually exclusive with the other restore/import sources.
    restore_source_table_arn = optional(string, "")

    # The point in time to restore, as a UTC RFC3339 timestamp (e.g.
    # "2026-07-04T06:00:00Z"). Exactly one of restore_date_time or
    # restore_to_latest_time accompanies a point-in-time restore source.
    restore_date_time = optional(string, "")

    # Restore the most recent recoverable state of the source table
    # instead of a specific timestamp.
    restore_to_latest_time = optional(bool, false)

    # Create this table by restoring an on-demand or AWS Backup backup,
    # by ARN. Key schema and attributes are inherited from the backup.
    # Mutually exclusive with the other restore/import sources.
    restore_backup_arn = optional(string, "")

    # Create this table pre-loaded with data imported from S3 (CSV,
    # DynamoDB JSON, or Amazon Ion) -- billed as a one-time import, far
    # cheaper than writing items individually. Mutually exclusive with
    # the restore sources; the key schema and attributes above are
    # required (imports define a brand-new table).
    import_table = optional(object({
      # The S3 bucket holding the source data. Reference an AwsS3Bucket
      # bucket_id output or pass a literal bucket name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      s3_bucket = string

      # The bucket owner's account ID -- set when importing from a bucket
      # in another account.
      s3_bucket_owner = optional(string, "")

      # Only objects under this key prefix are imported. Empty imports the
      # whole bucket.
      s3_key_prefix = optional(string, "")

      # The source data format: "CSV", "DYNAMODB_JSON", or "ION".
      input_format = string

      # How the source objects are compressed: "GZIP", "ZSTD", or "NONE".
      # Empty keeps the AWS default (NONE).
      input_compression_type = optional(string, "")

      # CSV parsing options; only meaningful when input_format is "CSV".
      csv = optional(object({
        # The field delimiter. Empty keeps the AWS default (",").
        delimiter = optional(string, "")

        # Column names, when the files carry no header row. Empty treats the
        # first row of each file as the header.
        header_list = optional(list(string), [])
      }))
    }))

    # Application Auto Scaling for PROVISIONED tables: target-tracking
    # policies that hold read/write capacity utilization near a target,
    # plus optional scheduled capacity adjustments. Application Auto
    # Scaling owns the table's live capacity on EVERY provisioned table
    # (without this block the modules register pinned min = max targets
    # from provisioned_throughput, so declared capacity still lands),
    # which is what lets this block be added or removed in place -- the
    # table resource itself never changes shape. Per-GSI autoscaling is
    # deliberately not modeled: neither engine can exempt only the inline
    # indexes' capacity from reconciliation, so an autoscaled GSI would
    # fight the scaler on every apply (the provider's standalone GSI
    # resource exists for that shape).
    autoscaling = optional(object({
      # Target tracking for read capacity. At least one of read/write must
      # be configured.
      read = optional(object({
        # The capacity floor the scaler never goes below.
        min_capacity = optional(number, 0)

        # The capacity ceiling the scaler never exceeds -- also the cost
        # guardrail.
        max_capacity = optional(number, 0)

        # The consumed-to-provisioned utilization percentage to hold. AWS
        # accepts 20-90 for DynamoDB. 70 is the usual production sweet spot:
        # headroom for spikes without paying for idle.
        target_utilization_percent = optional(number, 0)

        # Seconds to wait after a scale-in before another may follow. 0 keeps
        # the AWS default.
        scale_in_cooldown_seconds = optional(number, 0)

        # Seconds to wait after a scale-out before another may follow. 0 keeps
        # the AWS default.
        scale_out_cooldown_seconds = optional(number, 0)
      }))

      # Target tracking for write capacity.
      write = optional(object({
        # The capacity floor the scaler never goes below.
        min_capacity = optional(number, 0)

        # The capacity ceiling the scaler never exceeds -- also the cost
        # guardrail.
        max_capacity = optional(number, 0)

        # The consumed-to-provisioned utilization percentage to hold. AWS
        # accepts 20-90 for DynamoDB. 70 is the usual production sweet spot:
        # headroom for spikes without paying for idle.
        target_utilization_percent = optional(number, 0)

        # Seconds to wait after a scale-in before another may follow. 0 keeps
        # the AWS default.
        scale_in_cooldown_seconds = optional(number, 0)

        # Seconds to wait after a scale-out before another may follow. 0 keeps
        # the AWS default.
        scale_out_cooldown_seconds = optional(number, 0)
      }))

      # Scheduled capacity adjustments (e.g. raise the floor before a
      # nightly batch job, lower it after). Each entry is keyed by name and
      # targets one dimension's registered scalable target.
      scheduled_adjustments = optional(list(object({
        # The adjustment name -- keys the provider resource; renaming replaces
        # this adjustment without touching siblings.
        name = string

        # Which capacity dimension this adjustment changes: "READ" or "WRITE".
        # The dimension's autoscaling target (read/write above) must be
        # configured.
        dimension = string

        # The schedule: "cron(...)" (recurring, e.g. "cron(0 6 * * ? *)"),
        # "rate(...)" (fixed interval), or "at(yyyy-mm-ddThh:mm:ss)"
        # (one-shot).
        schedule = optional(string, "")

        # IANA timezone for the schedule (e.g. "America/Los_Angeles"). Empty
        # keeps the AWS default (UTC).
        timezone = optional(string, "")

        # The new capacity floor when the schedule fires. At least one of
        # min_capacity/max_capacity must be set. 0 leaves the floor unchanged.
        min_capacity = optional(number, 0)

        # The new capacity ceiling when the schedule fires. 0 leaves the
        # ceiling unchanged.
        max_capacity = optional(number, 0)

        # RFC3339 UTC timestamp the schedule takes effect (e.g.
        # "2026-09-01T00:00:00Z"). Empty starts immediately.
        start_time = optional(string, "")

        # RFC3339 UTC timestamp the schedule stops firing. Empty never
        # expires.
        end_time = optional(string, "")
      })), [])
    }))
  })
}
