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
  description = "AwsFsxOpenzfsFileSystem specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Deployment type controlling availability and performance characteristics.
    # ForceNew — cannot be changed after creation.
    #
    # - "SINGLE_AZ_1": first-generation single-AZ. Throughput 64-4096 MB/s.
    # - "SINGLE_AZ_2": current-generation single-AZ. Throughput 160-10240 MB/s.
    #   Recommended for most workloads.
    # - "SINGLE_AZ_HA_1" / "SINGLE_AZ_HA_2": HA (active/standby) variants of the
    #   two generations within one AZ — automatic failover without cross-AZ
    #   data transfer charges.
    # - "MULTI_AZ_1": active/standby across two AZs. Requires two subnets in
    #   different AZs plus preferred_subnet_id; the only type supporting the
    #   INTELLIGENT_TIERING storage class.
    #
    # Default: SINGLE_AZ_2
    deployment_type = optional(string)

    # Storage capacity in GiB for provisioned (SSD) storage. Range: 64-524288.
    # Can be increased after creation but never decreased.
    #
    # Leave unset in exactly two cases: the INTELLIGENT_TIERING storage class
    # (capacity is elastic and never provisioned) or a backup restore
    # (`backup_id` — capacity comes from the backup).
    storage_capacity_gib = optional(number)

    # Storage class backing the file system. ForceNew.
    #
    # - "SSD": provisioned solid-state storage (default; all deployment types).
    # - "INTELLIGENT_TIERING": elastic, pay-for-what-you-store capacity with a
    #   provisioned SSD read cache. MULTI_AZ_1 only; forbids
    #   storage_capacity_gib and requires read_cache_configuration.
    #
    # Default: SSD
    storage_type = optional(string)

    # Throughput capacity in MB/s. Required.
    #
    # Valid values by deployment generation (the values AWS accepts):
    # - SINGLE_AZ_1: 64, 128, 256, 512, 1024, 2048, 3072, 4096
    # - SINGLE_AZ_2 / MULTI_AZ_1: 160, 320, 640, 1280, 2560, 3840, 5120,
    #   7680, 10240
    # - The HA variants follow their generation's value set (validated by AWS
    #   at create time).
    #
    # Can be changed after creation to scale performance up or down.
    throughput_capacity = optional(number, 0)

    # Provisioned SSD read cache for the INTELLIGENT_TIERING storage class.
    # Required when storage_type is INTELLIGENT_TIERING; invalid otherwise.
    read_cache_configuration = optional(object({
      # How the read cache is sized.
      #
      # - "PROPORTIONAL_TO_THROUGHPUT_CAPACITY": AWS sizes the cache from the
      #   provisioned throughput (the recommended hands-off mode).
      # - "USER_PROVISIONED": you set the exact cache size via `size_gib`.
      # - "NO_CACHE": no SSD read cache (every read pays the tiered-storage
      #   latency; only for purely archival access patterns).
      sizing_mode = string

      # Read cache size in GiB when sizing_mode is "USER_PROVISIONED".
      size_gib = optional(number)
    }))

    # Subnet IDs for the file system's network interfaces. Required. ForceNew.
    #
    # - Single-AZ types (incl. the HA variants): exactly one subnet.
    # - MULTI_AZ_1: exactly two subnets in different availability zones.
    #
    # All compute resources mounting this file system must have network
    # connectivity to these subnets.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Security groups for the file system's network interfaces. ForceNew.
    # Up to 50. When empty, AWS attaches the VPC's default security group.
    #
    # Must allow NFS traffic between the file system and its clients:
    # - TCP port 111 (portmapper)
    # - TCP port 2049 (NFS)
    # - TCP ports 20001-20003 (NFS mount)
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # Preferred subnet for the active file server in a MULTI_AZ_1 deployment.
    # ForceNew. REQUIRED for MULTI_AZ_1 (AWS's contract) and invalid for the
    # single-AZ types. Must be one of the subnets in subnet_ids.
    #
    # In a failover event, the standby file server in the other subnet takes
    # over.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    preferred_subnet_id = optional(string, "")

    # IP address range for the file system endpoints in a MULTI_AZ_1
    # deployment. ForceNew. Must be a CIDR block within the VPC's CIDR range
    # that does not overlap with any existing subnets; AWS assigns floating IPs
    # from this range for seamless failover. When omitted, AWS picks a range.
    #
    # The provider does not validate this field's format for OpenZFS (unlike its
    # ONTAP sibling) — the CEL below mirrors the ONTAP kind so a malformed range
    # fails at validate instead of at the AWS API.
    endpoint_ip_address_range = optional(string, "")

    # Route tables in which AWS manages routes to the floating file-system
    # endpoints of a MULTI_AZ_1 deployment. Specify every VPC route table
    # associated with the subnets your NFS clients live in; when omitted, AWS
    # uses the VPC's default route table. Up to 50.
    #
    # Reference an AwsSubnet's route_table_id output when the subnet owns its
    # table, or the AwsVpc's main/default route-table outputs when subnets ride
    # the VPC main table; literals also work.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    route_table_ids = optional(list(string), [])

    # Customer-managed KMS key ARN for encryption at rest. ForceNew — the KMS
    # key cannot be changed after creation. When omitted, the file system uses
    # the AWS-managed FSx key. All OpenZFS file systems are encrypted at rest
    # by default; this field upgrades to a customer-managed key.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # ID of an FSx backup to restore this file system from ("backup-...").
    # ForceNew. When set, capacity and most settings come from the backup;
    # leave storage_capacity_gib unset.
    backup_id = optional(string, "")

    # SSD IOPS configuration for the file system. Controls the total
    # provisioned IOPS. When omitted, AWS uses AUTOMATIC mode which scales IOPS
    # with storage (3 IOPS per GiB).
    disk_iops_configuration = optional(object({
      # IOPS provisioning mode.
      #
      # - "AUTOMATIC": IOPS scale automatically based on storage capacity.
      #   Provides 3 IOPS per GiB of storage, up to the deployment type limit.
      # - "USER_PROVISIONED": you specify the exact IOPS via the `iops` field.
      #   Allows higher performance independent of storage size but at extra cost.
      #
      # Default: AUTOMATIC
      mode = optional(string)

      # Total SSD IOPS provisioned. Only valid when mode is "USER_PROVISIONED".
      #
      # Ceilings by deployment generation: 160,000 (SINGLE_AZ_1) and 400,000
      # (SINGLE_AZ_2); MULTI_AZ_1 and the HA variants are validated by AWS at
      # create time.
      iops = optional(number)
    }))

    # Configuration for the file system's root volume. The root volume is
    # automatically created with the file system and serves as the default NFS
    # mount target. Settings here control compression, NFS access, quotas, and
    # record size.
    #
    # When omitted, the root volume uses defaults: no compression, no NFS
    # access restrictions, 128 KiB record size.
    root_volume_configuration = optional(object({
      # Data compression type applied to all data on the root volume. Reduces
      # storage consumption and can improve throughput for compressible data.
      #
      # - "NONE": no compression (default).
      # - "ZSTD": Zstandard compression. Best compression ratio.
      # - "LZ4": LZ4 compression. Faster with lower CPU overhead.
      #
      # Default: NONE
      data_compression_type = optional(string)

      # NFS export configuration for the root volume. Controls which clients can
      # mount the volume and with what permissions. When omitted, the volume uses
      # default NFS settings (accessible from within the VPC).
      nfs_exports = optional(object({
        # NFS client configurations. Each entry defines which clients can access the
        # volume and with what mount options. Up to 25 client configurations.
        client_configurations = list(object({
          # Client specification: an IP address, CIDR block, or wildcard (*).
          # 1-128 characters.
          #
          # Examples: "*" (all clients), "10.0.0.0/16", "192.168.1.100"
          clients = string

          # NFS mount options for the specified clients. At least one option is
          # required; each option is 1-128 characters.
          #
          # Common options:
          # - "rw" (read-write) or "ro" (read-only)
          # - "crossmnt" (allow traversal into child volumes)
          # - "root_squash" (map root to anonymous) or "no_root_squash"
          # - "sync" or "async"
          #
          # Up to 20 options.
          options = list(string)
        }))
      }))

      # Whether the root volume is read-only. When true, clients can mount the
      # volume but cannot write to it. Useful for shared reference data.
      read_only = optional(bool, false)

      # ZFS record size in KiB. Affects read/write performance characteristics.
      # Larger record sizes are better for sequential I/O (analytics, streaming).
      # Smaller record sizes are better for random I/O (databases, OLTP).
      #
      # Valid values: 4, 8, 16, 32, 64, 128, 256, 512, 1024.
      #
      # Default: 128
      record_size_kib = optional(number)

      # Per-user and per-group storage quotas for the root volume. Limits how
      # much storage individual users or groups can consume.
      user_and_group_quotas = optional(list(object({
        # The numeric user ID (UID) or group ID (GID). Range: 0-2147483647.
        #
        # Common values: 0 (root), 1000+ (regular users/groups).
        id = optional(number, 0)

        # Storage capacity quota in GiB. The maximum amount of storage this user or
        # group can consume on the volume. Range: 0-2147483647.
        #
        # Set to 0 to remove quota restrictions for this user/group.
        storage_capacity_quota_gib = optional(number, 0)

        # Quota type: "USER" for per-user quota, "GROUP" for per-group quota.
        type = string
      })), [])

      # Copy tags from the root volume to snapshots created from it. ForceNew —
      # changing this replaces the whole FILE SYSTEM (a subtle provider trap:
      # this is the one root-volume setting that cannot change in place).
      copy_tags_to_snapshots = optional(bool, false)
    }))

    # Number of days to retain automatic backups. Range: 0-90; 0 disables
    # automatic backups.
    #
    # Default: 0 (no automatic backups)
    automatic_backup_retention_days = optional(number)

    # Daily UTC time to start automatic backups, in "HH:MM" format (e.g.,
    # "05:00"). If not specified and backups are enabled, AWS chooses a window.
    daily_automatic_backup_start_time = optional(string, "")

    # Copy tags from the file system to backups.
    copy_tags_to_backups = optional(bool, false)

    # Copy tags from the file system to volumes. When true, tags are propagated
    # to the root volume and any child volumes created on this file system.
    copy_tags_to_volumes = optional(bool, false)

    # Skip creating a final backup when the file system is deleted.
    #
    # Default: true — deletion is clean by default; set to false to keep a
    # last-resort restore point (the final backup outlives the file system and
    # keeps billing until deleted).
    skip_final_backup = optional(bool)

    # Tags applied to the final backup taken on deletion. Only meaningful when
    # skip_final_backup is false.
    final_backup_tags = optional(map(string), {})

    # Options applied when the file system is deleted.
    # The single supported value, "DELETE_CHILD_VOLUMES_AND_SNAPSHOTS", deletes
    # all child volumes and snapshots along with the file system — without it,
    # deletion fails while children exist. Use deliberately: it turns a
    # guard-railed delete into a cascading one.
    delete_options = optional(list(string), [])

    # Weekly UTC maintenance window in the format "d:HH:MM" where d is the day
    # of the week (1=Monday, 7=Sunday). Example: "1:05:00" for Monday at 05:00
    # UTC. If not specified, AWS chooses a default window.
    weekly_maintenance_start_time = optional(string, "")
  })
}
