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
  description = "AwsFsxLustreFileSystem specification"
  type = object({
    # The AWS region where the file system will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Deployment type controlling data durability and performance characteristics.
    # ForceNew — cannot be changed after creation.
    #
    # - "SCRATCH_1": temporary storage, no replication. Legacy; fixed 200 MB/s/TiB
    #   throughput. Growing storage on SCRATCH_1 replaces the file system.
    # - "SCRATCH_2": temporary storage, no replication, burst throughput up to
    #   1300 MB/s/TiB. Recommended for short-lived processing jobs.
    # - "PERSISTENT_1": within-AZ-replicated storage with automatic backups.
    #   The only deployment type supporting HDD storage and the legacy S3
    #   import/export arms alongside SCRATCH types.
    # - "PERSISTENT_2": the current persistent generation — higher throughput
    #   tiers (125-1000 MB/s/TiB), metadata IOPS configuration, EFA/GPUDirect
    #   support, and the INTELLIGENT_TIERING storage class. Recommended for new
    #   production workloads.
    #
    # Default: SCRATCH_2 (the provider's own default is the legacy SCRATCH_1;
    # this spec recommends SCRATCH_2 for its strictly better burst throughput at
    # the same price).
    deployment_type = optional(string)

    # Storage capacity in GiB. Minimum 1200. Valid step sizes depend on the
    # deployment and storage types (AWS enforces these at create time):
    #
    # - SCRATCH_2 / PERSISTENT_1 / PERSISTENT_2 (SSD): 1200, 2400, then
    #   increments of 2400.
    # - PERSISTENT_1 (HDD): increments of 6000 (12 MB/s/TiB) or 1800
    #   (40 MB/s/TiB).
    # - SCRATCH_1: 1200, 2400, 3600, then increments of 3600.
    #
    # Can be increased in place (never decreased; growth on SCRATCH_1 replaces
    # the file system). Leave unset in exactly two cases: restoring from a
    # backup (`backup_id` — capacity comes from the backup), or the
    # INTELLIGENT_TIERING storage class (capacity is elastic and never
    # provisioned).
    storage_capacity_gib = optional(number)

    # Storage class backing the file system. ForceNew.
    #
    # - "SSD": solid-state drives, sub-millisecond latency. Required for
    #   SCRATCH_1/SCRATCH_2 and the default for both PERSISTENT generations.
    # - "HDD": hard disk drives — lowest cost per TiB for sequential,
    #   throughput-oriented workloads. PERSISTENT_1 only; requires
    #   drive_cache_type and per_unit_storage_throughput 12 or 40.
    # - "INTELLIGENT_TIERING": elastic, pay-for-what-you-store capacity with a
    #   provisioned SSD read cache. PERSISTENT_2 only; requires
    #   throughput_capacity, data_read_cache_configuration, and
    #   metadata_configuration, and forbids provisioned storage_capacity_gib.
    #
    # Default: SSD
    storage_type = optional(string)

    # Throughput per unit of storage in MB/s/TiB, for provisioned-capacity
    # PERSISTENT deployments (SSD and HDD storage). Invalid for SCRATCH types
    # and for INTELLIGENT_TIERING (which provisions throughput absolutely via
    # throughput_capacity instead).
    #
    # Valid values by deployment and storage type:
    # - PERSISTENT_1 + SSD: 50, 100, 200
    # - PERSISTENT_1 + HDD: 12, 40
    # - PERSISTENT_2 + SSD: 125, 250, 500, 1000
    #
    # Can be changed in place on PERSISTENT_2 (throughput scales while the file
    # system stays online), except when efa_enabled pins it at creation.
    per_unit_storage_throughput = optional(number)

    # Absolute throughput in MB/s for the INTELLIGENT_TIERING storage class.
    # Must be 4000 or a multiple of 4000. Required when (and only meaningful
    # when) storage_type is INTELLIGENT_TIERING — provisioned-capacity file
    # systems size their throughput per-TiB via per_unit_storage_throughput
    # instead.
    throughput_capacity = optional(number)

    # Enable LZ4 data compression for all data on the file system. Reduces
    # storage consumption and can improve throughput for compressible data.
    # Can be changed after creation (new writes are compressed; existing data
    # is not rewritten).
    #
    # - "NONE": no compression (default).
    # - "LZ4": LZ4 compression.
    data_compression_type = optional(string)

    # Lustre file system version, in "x.y" format (e.g., "2.12", "2.15").
    # Leave empty to use the latest version supported by the deployment type.
    # Upgrades apply in place; a downgrade replaces the file system.
    file_system_type_version = optional(string, "")

    # Enable Elastic Fabric Adapter (EFA) and GPUDirect Storage (GDS) support,
    # giving GPU instances a direct, OS-bypass data path to the file system.
    # ForceNew — must be decided at creation, and while enabled it also pins
    # per_unit_storage_throughput. Requires PERSISTENT_2 with
    # metadata_configuration, and an EFA-enabled security group attached via
    # security_group_ids.
    efa_enabled = optional(bool, false)

    # Read cache for HDD-backed file systems. PERSISTENT_1 + HDD only, and
    # REQUIRED there (AWS's contract for HDD file systems). ForceNew.
    #
    # - "READ": provision an SSD read cache sized to 20% of storage capacity —
    #   gives HDD file systems SSD-like latency for frequently read data.
    # - "NONE": no read cache.
    drive_cache_type = optional(string, "")

    # Provisioned SSD read cache for the INTELLIGENT_TIERING storage class.
    # Required when storage_type is INTELLIGENT_TIERING; invalid otherwise.
    data_read_cache_configuration = optional(object({
      # How the read cache is sized.
      #
      # - "PROPORTIONAL_TO_THROUGHPUT_CAPACITY": AWS sizes the cache from the
      #   provisioned throughput (the recommended hands-off mode).
      # - "USER_PROVISIONED": you set the exact cache size via `size_gib`.
      # - "NO_CACHE": no SSD read cache (every read pays the tiered-storage
      #   latency; only for purely archival access patterns).
      sizing_mode = string

      # Read cache size in GiB when sizing_mode is "USER_PROVISIONED". The valid
      # range scales with throughput_capacity: for every 4000 MB/s provisioned,
      # AWS accepts 32 GiB to 131072 GiB of cache (e.g., 8000 MB/s allows
      # 64-262144 GiB).
      size_gib = optional(number)
    }))

    # Subnet for the file system's network interfaces. Required. ForceNew.
    #
    # Lustre file systems are single-AZ — exactly one subnet is supported. All
    # compute resources mounting this file system must have network connectivity
    # to this subnet.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_id = string

    # Security groups for the file system's network interfaces. ForceNew.
    # Up to 50. When empty, AWS attaches the VPC's default security group.
    #
    # Must allow Lustre traffic between the file system and its clients:
    # - TCP port 988 (Lustre protocol)
    # - TCP ports 1018-1023 (Lustre data channels)
    # EFA-enabled file systems additionally need an EFA-enabled security group.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Customer-managed KMS key ARN for encryption at rest. ForceNew. When
    # omitted, the file system uses the AWS-managed FSx key (all Lustre file
    # systems are encrypted at rest); this field upgrades to a customer-managed
    # key.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # ID of an FSx backup to restore this file system from ("backup-...").
    # ForceNew. When set, storage capacity and most file-system settings come
    # from the backup; leave storage_capacity_gib unset.
    backup_id = optional(string, "")

    # S3 URI to link as the file system's data repository (e.g., "s3://my-bucket"
    # or "s3://my-bucket/prefix"). ForceNew. Not supported on PERSISTENT_2 —
    # use AwsFsxDataRepositoryAssociation there (and prefer it for PERSISTENT_1
    # too; this arm is the legacy single-link generation).
    #
    # When set, the file system imports file metadata from S3 at creation; file
    # data is lazy-loaded on first access.
    import_path = optional(string, "")

    # S3 URI where changed files are exported back to S3 (e.g.,
    # "s3://my-bucket/output/"). ForceNew. Requires import_path and must use the
    # same bucket. Set equal to import_path to overwrite objects in place; when
    # omitted AWS exports to "s3://{import bucket}/FSxLustre{creation timestamp}".
    export_path = optional(string, "")

    # How the file system stays in sync as objects change in the linked S3
    # bucket. Requires import_path.
    #
    # - "NONE": import listings only at creation (default).
    # - "NEW": import metadata for objects added to the bucket.
    # - "NEW_CHANGED": also update metadata for changed objects.
    # - "NEW_CHANGED_DELETED": also delete file metadata when objects are
    #   deleted from the bucket.
    auto_import_policy = optional(string, "")

    # Stripe configuration for imported files: the maximum amount of data per
    # file (in MiB) stored on a single physical disk. Range: 1-512000. Requires
    # import_path; AWS defaults to 1024. ForceNew.
    imported_file_chunk_size = optional(number)

    # Root squash configuration — maps root (UID/GID 0) clients to an
    # unprivileged identity so no mounting host has automatic root access to
    # the file system's contents. A POSIX-security hardening measure for
    # multi-tenant compute fleets. Can be changed after creation.
    root_squash_configuration = optional(object({
      # The UID:GID pair that root users are squashed to (e.g., "65534:65534" for
      # nobody:nogroup). Both values range 0-4294967294. Setting this enables
      # root squash; omit the whole block to leave root access unrestricted.
      root_squash = optional(string, "")

      # Lustre NIDs (network identifiers) of clients EXEMPT from root squash —
      # administrative hosts that keep real root access. Format: an IPv4 address
      # (ranges allowed in brackets) followed by "@tcp", e.g. "10.0.1.6@tcp" or
      # "10.0.[2-10].[1-255]@tcp".
      no_squash_nids = optional(list(string), [])
    }))

    # CloudWatch logging for data repository events (imports/exports between
    # the file system and its linked S3 repositories). Useful for auditing
    # repository task failures and lifecycle debugging.
    log_configuration = optional(object({
      # CloudWatch Logs log group ARN to receive the events. When set, the log
      # group must exist and have a resource policy allowing FSx to write to it.
      # When left empty with a level set, FSx logs to its DEFAULT log group
      # (/aws/fsx/lustre) -- logging is NOT disabled by omitting the
      # destination. Logging is off only when the whole log_configuration
      # message is absent (or the level is DISABLED).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination = optional(string, "")

      # Log level controlling which events are logged.
      #
      # - "DISABLED": no logging (default when log_configuration is omitted).
      # - "WARN_ONLY": log warning-level events only.
      # - "ERROR_ONLY": log error-level events only.
      # - "WARN_ERROR": log both warning and error events.
      #
      # Default: WARN_ERROR
      level = optional(string)
    }))

    # Metadata performance configuration. PERSISTENT_2 only (required there
    # when storage_type is INTELLIGENT_TIERING or efa_enabled is set). Controls
    # the metadata IOPS available for file creation, listing, and similar
    # operations. Most workloads perform well with AUTOMATIC mode.
    metadata_configuration = optional(object({
      # Metadata IOPS mode.
      #
      # - "AUTOMATIC": FSx scales metadata IOPS with the file system's storage
      #   capacity. The right choice for most workloads.
      # - "USER_PROVISIONED": you specify exact metadata IOPS — higher
      #   metadata performance independent of storage size, at additional cost.
      #
      # Default: AUTOMATIC
      mode = optional(string)

      # Metadata IOPS when mode is "USER_PROVISIONED". IOPS can be increased in
      # place; a decrease replaces the file system.
      #
      # Valid values: 1500, 3000, 6000, then multiples of 12000 up to 192000.
      iops = optional(number)
    }))

    # Number of days to retain automatic backups. Range: 0-90; 0 disables
    # automatic backups. Backups are only supported on PERSISTENT deployments.
    #
    # Default: 0 (no automatic backups)
    automatic_backup_retention_days = optional(number)

    # Daily UTC time to start automatic backups, in "HH:MM" format (e.g.,
    # "05:00"). Only meaningful when automatic backups are enabled; when
    # omitted AWS chooses a window.
    daily_automatic_backup_start_time = optional(string, "")

    # Copy the file system's tags to its automatic backups. ForceNew.
    copy_tags_to_backups = optional(bool, false)

    # Skip creating a final backup when the file system is deleted. Applies to
    # PERSISTENT deployments (SCRATCH file systems have no backups).
    #
    # Default: true — deletion is clean by default; set to false to keep a
    # last-resort restore point (the final backup outlives the file system and
    # keeps billing until deleted).
    skip_final_backup = optional(bool)

    # Tags applied to the final backup taken on deletion. Only meaningful when
    # skip_final_backup is false.
    final_backup_tags = optional(map(string), {})

    # Weekly UTC maintenance window in "d:HH:MM" format where d is the day of
    # the week (1=Monday, 7=Sunday). Example: "1:05:00" for Monday 05:00 UTC.
    weekly_maintenance_start_time = optional(string, "")
  })
}
