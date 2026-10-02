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
  description = "AwsFsxOntapFileSystem specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Deployment type controlling availability, performance, and scale-out
    # characteristics. ForceNew — cannot be changed after creation.
    #
    # - "SINGLE_AZ_1": first-generation single-AZ. Single HA pair only; adding
    #   HA pairs or increasing per-pair throughput replaces the file system.
    # - "SINGLE_AZ_2": current single-AZ generation with scale-out HA pairs
    #   (1-12) that can be added in place. Recommended for most workloads.
    # - "MULTI_AZ_1": first-generation multi-AZ with automatic failover across
    #   two AZs. Fixed at 1 HA pair. Requires two subnets and preferred_subnet_id.
    # - "MULTI_AZ_2": current multi-AZ generation with automatic failover.
    #   Fixed at 1 HA pair. Recommended for high-availability workloads.
    #
    # Default: SINGLE_AZ_2
    deployment_type = optional(string)

    # Storage capacity in GiB. Required.
    #
    # The valid range scales with HA pairs: minimum 1024 GiB per HA pair,
    # maximum 524288 GiB (512 TiB) per HA pair, up to the 1048576 GiB (1 PiB)
    # absolute ceiling on SINGLE_AZ_2 scale-out deployments. First-generation
    # deployments (SINGLE_AZ_1 / MULTI_AZ_1) cap at 196608 GiB (192 TiB).
    #
    # Storage can be increased after creation but never decreased. Choose based
    # on data size; ONTAP's built-in compression and deduplication typically
    # achieve 2-5x data reduction, and volume tiering to capacity-pool storage
    # stretches the SSD tier further.
    storage_capacity_gib = optional(number, 0)

    # Storage media type. ForceNew — cannot be changed after creation.
    #
    # ONTAP file systems support only "SSD" (sub-millisecond-latency solid-state
    # primary storage). The other FSx storage classes do not apply here: HDD is
    # a Windows/Lustre option and INTELLIGENT_TIERING an OpenZFS/Lustre option —
    # AWS rejects both for ONTAP at create time. Cost tiering on ONTAP is
    # instead achieved per volume via `tiering_policy` on AwsFsxOntapVolume,
    # which moves cold data to the built-in elastic capacity pool.
    #
    # Default: SSD
    storage_type = optional(string)

    # Total throughput capacity for the whole file system in MB/s — the
    # first-generation sizing arm. Exactly one of `throughput_capacity` and
    # `throughput_capacity_per_ha_pair` must be set.
    #
    # Valid values: 128, 256, 512, 1024, 2048, 4096.
    #
    # Use this arm for SINGLE_AZ_1 / MULTI_AZ_1 file systems (single HA pair,
    # so whole-system and per-pair sizing coincide). For the current generation
    # (SINGLE_AZ_2 / MULTI_AZ_2), prefer `throughput_capacity_per_ha_pair`,
    # which carries the second generation's throughput tiers.
    throughput_capacity = optional(number)

    # Throughput capacity per HA pair in MB/s — the per-pair sizing arm.
    # Exactly one of `throughput_capacity` and `throughput_capacity_per_ha_pair`
    # must be set. Total file system throughput = this value × ha_pairs.
    #
    # Valid values by deployment type:
    # - SINGLE_AZ_1 / MULTI_AZ_1: 128, 256, 512, 1024, 2048, 4096.
    # - SINGLE_AZ_2 / MULTI_AZ_2 with 1 HA pair: 384, 768, 1536, 3072, 6144.
    # - SINGLE_AZ_2 with multiple HA pairs: 1536, 3072, 6144 per pair.
    #
    # Scales in place on SINGLE_AZ_2 and MULTI_AZ_2. On SINGLE_AZ_1 and
    # MULTI_AZ_1, increasing this value replaces the file system.
    throughput_capacity_per_ha_pair = optional(number)

    # Number of high-availability pairs in the file system. Each HA pair adds
    # an independent pair of file servers contributing throughput, IOPS, and
    # up to 512 TiB of storage capacity.
    #
    # Only SINGLE_AZ_2 supports scale-out (1-12 HA pairs, added in place).
    # SINGLE_AZ_1, MULTI_AZ_1, and MULTI_AZ_2 are fixed at 1 HA pair.
    #
    # Default: 1
    ha_pairs = optional(number)

    # Subnet IDs for the file system's network interfaces. Required. ForceNew.
    #
    # - SINGLE_AZ_1 / SINGLE_AZ_2: exactly one subnet.
    # - MULTI_AZ_1 / MULTI_AZ_2: exactly two subnets in different availability
    #   zones (the active and standby file servers).
    #
    # All compute resources accessing this file system (via NFS, SMB, or iSCSI)
    # must have network connectivity to these subnets.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Preferred subnet for the active file server in a multi-AZ deployment.
    # ForceNew. Required when deployment_type is MULTI_AZ_1 or MULTI_AZ_2 (and
    # invalid otherwise — single-AZ file systems have only one subnet). Must be
    # one of the subnets specified in subnet_ids.
    #
    # In a failover event, the standby file server in the other subnet takes
    # over automatically. Place the active server in the same AZ as the bulk of
    # your clients to avoid cross-AZ data charges during normal operation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    preferred_subnet_id = optional(string, "")

    # Security groups for the file system's network interfaces. ForceNew.
    # Up to 50. When empty, AWS attaches the VPC's default security group.
    #
    # Must allow traffic between the file system and its clients:
    # - TCP port 111 (portmapper)
    # - TCP port 635 (mountd)
    # - TCP port 2049 (NFS)
    # - TCP ports 4045-4046 (NFS lock/status)
    # - TCP port 445 (SMB)
    # - TCP port 3260 (iSCSI)
    # - TCP port 443 (ONTAP REST API)
    # - TCP port 22 (SSH to the ONTAP CLI, if fsx_admin_password is used)
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # IP address range for the file system endpoints in a multi-AZ deployment,
    # in CIDR notation (e.g., "198.19.0.0/24"). ForceNew. AWS assigns floating
    # IPs from this range so the management, intercluster, and SVM data
    # endpoints survive failover with unchanged addresses.
    #
    # Must NOT overlap with any subnet in the VPC (AWS recommends a range
    # outside the VPC CIDR, such as the 198.19.0.0/16 block); clients reach it
    # through the route tables in route_table_ids. When omitted, AWS picks an
    # unused range automatically. Only valid for MULTI_AZ_1 / MULTI_AZ_2.
    endpoint_ip_address_range = optional(string, "")

    # Route tables in which AWS creates and manages routes to the floating
    # endpoint IP range of a multi-AZ deployment. Specify every route table
    # associated with the subnets your clients live in; AWS repoints the routes
    # automatically on failover. Up to 50 route tables. When omitted, AWS uses
    # the VPC's main route table. Only valid for MULTI_AZ_1 / MULTI_AZ_2.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    route_table_ids = optional(list(string), [])

    # Customer-managed KMS key ARN for encryption at rest. ForceNew — the KMS key
    # cannot be changed after creation. When omitted, the file system uses the
    # AWS-managed FSx key. All ONTAP file systems are encrypted at rest by
    # default; this field upgrades to a customer-managed key.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Password for the ONTAP administrative user ("fsxadmin"). Enables SSH and
    # REST API access to the file system for advanced administration such as LIF
    # management, SnapMirror configuration, and aggregate monitoring. Can be
    # changed after creation.
    #
    # Length: 8-50 characters. Optional — omit if ONTAP CLI access is not needed.
    # This value is sensitive and will not be returned in read operations.
    fsx_admin_password = optional(string, "")

    # SSD IOPS configuration for the file system. Controls the total provisioned
    # IOPS. When omitted, AWS uses AUTOMATIC mode which provisions 3 IOPS per GiB
    # of storage capacity. Use USER_PROVISIONED mode for workloads requiring IOPS
    # beyond what AUTOMATIC provides. Can be changed after creation.
    disk_iops_configuration = optional(object({
      # IOPS provisioning mode.
      #
      # - "AUTOMATIC": IOPS scale automatically based on storage capacity.
      #   Provides 3 IOPS per GiB of storage.
      # - "USER_PROVISIONED": you specify the exact IOPS via the `iops` field.
      #   Allows higher performance independent of storage size but at extra cost.
      #
      # Default: AUTOMATIC
      mode = optional(string)

      # Total SSD IOPS provisioned. Only valid when mode is "USER_PROVISIONED".
      #
      # Valid range: 0–2,400,000. The maximum achievable IOPS depends on the number
      # of HA pairs and their throughput capacity tier.
      iops = optional(number, 0)
    }))

    # Number of days to retain automatic backups. Range: 0-90. Set to 0 to
    # disable automatic backups. ONTAP's built-in snapshots provide point-in-time
    # recovery independently of FSx backups, and volume-level backup settings
    # live on AwsFsxOntapVolume.
    #
    # Default: 0 (no automatic backups)
    automatic_backup_retention_days = optional(number)

    # Daily UTC time to start automatic backups, in "HH:MM" format (e.g.,
    # "05:00"). Only meaningful when automatic backups are enabled; when
    # omitted AWS chooses a window.
    daily_automatic_backup_start_time = optional(string, "")

    # Weekly UTC maintenance window in "d:HH:MM" format where d is the day of
    # the week (1=Monday, 7=Sunday). Example: "7:02:00" for Sunday at 02:00 UTC.
    # If not specified, AWS chooses a default window.
    weekly_maintenance_start_time = optional(string, "")
  })
}
