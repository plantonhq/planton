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
  description = "AwsElasticFileSystem specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Enable encryption at rest for all data and metadata stored in the file system.
    # ForceNew — cannot be added after creation. Production environments should
    # always enable encryption.
    encrypted = optional(bool, false)

    # Customer-managed KMS key for encryption at rest. When omitted, EFS uses the
    # AWS-managed key `aws/elasticfilesystem`. ForceNew — the KMS key cannot be
    # changed after creation. Requires `encrypted` to be true.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # File system performance mode. ForceNew — cannot be changed after creation.
    # Empty keeps the AWS default ("generalPurpose").
    #
    # - "generalPurpose" (default): lowest latency, suitable for most workloads.
    #   Recommended for all new file systems, especially with "elastic" throughput.
    # - "maxIO": higher aggregate throughput for highly parallelized workloads
    #   (thousands of EC2 instances). Slightly higher per-operation latency.
    #   AWS recommends generalPurpose + elastic throughput as a replacement, and
    #   maxIO cannot be combined with elastic throughput or One Zone storage.
    performance_mode = optional(string, "")

    # Throughput mode controlling how EFS delivers read/write bandwidth.
    # Empty keeps the AWS default ("bursting").
    #
    # - "bursting" (default): throughput scales with file system size. 50 MiB/s
    #   per TiB of Standard storage, with bursts up to 100 MiB/s.
    # - "provisioned": fixed throughput independent of storage size. Set
    #   `provisioned_throughput_in_mibps` to specify the exact value.
    # - "elastic": automatically scales throughput up/down based on workload.
    #   Recommended for unpredictable or spiky access patterns. Requires the
    #   generalPurpose performance mode (AWS rejects elastic + maxIO).
    #
    # Throughput mode is mutable, but AWS enforces a 24-hour cooldown between
    # throughput-mode changes (and between decreases of provisioned throughput).
    throughput_mode = optional(string, "")

    # Provisioned throughput in MiB/s. Only applicable when `throughput_mode` is
    # "provisioned". AWS accepts 1.0–3414.0 for generalPurpose and 1.0–1024.0 for
    # maxIO (server-side limits; higher values require a quota increase).
    provisioned_throughput_in_mibps = optional(number, 0)

    # AWS Availability Zone name for One Zone storage classes (e.g., "us-east-1a").
    # ForceNew — cannot be changed after creation. One Zone storage is ~47% cheaper
    # than Standard (multi-AZ) but data is stored in a single AZ with no cross-AZ
    # redundancy. Suitable for dev/test or workloads that tolerate AZ-level failure.
    #
    # When set, only a single mount target can be created (in a subnet belonging
    # to this AZ).
    availability_zone_name = optional(string, "")

    # Transition files to Infrequent Access (IA) storage after the specified period
    # of not being accessed. IA storage costs ~92% less than Standard but charges
    # per-access fees.
    #
    # Valid values: AFTER_1_DAY, AFTER_7_DAYS, AFTER_14_DAYS, AFTER_30_DAYS,
    # AFTER_60_DAYS, AFTER_90_DAYS, AFTER_180_DAYS, AFTER_270_DAYS, AFTER_365_DAYS.
    transition_to_ia = optional(string, "")

    # Transition IA files to Archive storage after the specified period. Archive
    # storage costs ~96% less than Standard. Requires `transition_to_ia` to be set
    # (files must pass through IA before reaching Archive).
    #
    # Same valid values as transition_to_ia.
    transition_to_archive = optional(string, "")

    # Transition files back to Standard storage when accessed from IA or Archive.
    # This enables automatic "warming" of frequently accessed files.
    #
    # Only valid value: "AFTER_1_ACCESS". Leave empty to keep files in IA/Archive
    # even after access.
    transition_to_primary_storage_class = optional(string, "")

    # Enable automatic daily backups via AWS Backup. AWS recommends enabling
    # backups for all production file systems. Mutable — can be toggled at any time.
    backup_enabled = optional(bool, false)

    # Controls whether this file system can be used as the DESTINATION of an EFS
    # replication configuration. Empty keeps the AWS default ("ENABLED").
    #
    # - "ENABLED" (default): the file system is protected — it cannot be
    #   overwritten by a replication from another file system.
    # - "DISABLED": the file system may be targeted as a replication destination.
    #   AWS also requires protection to be DISABLED before a replication
    #   destination file system can be modified or deleted after replication
    #   stops — set this when adopting an existing file system as a replica.
    replication_overwrite_protection = optional(string, "")

    # Mount targets expose the file system as an NFS endpoint inside a VPC.
    # Required (min 1) — an EFS without mount targets cannot be mounted. Declare
    # one mount target per Availability Zone (AWS allows at most one per AZ;
    # two subnets in the same AZ fail at deploy time).
    #
    # For regional (multi-AZ) file systems, declare one mount target per AZ for
    # maximum availability and to avoid cross-AZ data charges. For One Zone file
    # systems, declare exactly one mount target in a subnet belonging to
    # `availability_zone_name`.
    mount_targets = list(object({
      # The subnet to place this mount target in. The subnet's Availability Zone
      # determines which AZ's clients this target serves.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_id = string

      # Static IPv4 address for the mount target, from the subnet's IPv4 CIDR.
      # ForceNew. When omitted, AWS assigns an address automatically. Useful for
      # static NFS mount configurations that cannot resolve the EFS DNS names.
      ip_address = optional(string, "")

      # Address family for the mount target. Empty keeps the AWS default
      # ("IPV4_ONLY"). ForceNew.
      #
      # - "IPV4_ONLY" (default): IPv4 address only.
      # - "IPV6_ONLY": IPv6 address only — requires an IPv6-enabled subnet.
      # - "DUAL_STACK": both IPv4 and IPv6 addresses.
      ip_address_type = optional(string, "")

      # Static IPv6 address for the mount target, from the subnet's IPv6 CIDR.
      # ForceNew. Only valid when `ip_address_type` is "IPV6_ONLY" or "DUAL_STACK".
      ipv6_address = optional(string, "")
    }))

    # Security groups applied to ALL mount targets. These must allow inbound NFS
    # traffic (TCP port 2049) from the clients that will mount the file system.
    # When omitted, AWS attaches the VPC's default security group.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])

    # IAM resource policy for the file system. Common uses:
    # - Enforce encryption in transit (deny unencrypted NFS connections)
    # - Restrict access to specific IAM principals or VPCs
    # - Prevent root access from NFS clients
    #
    # Provide as a JSON object structure. Serialized to JSON by IaC modules.
    # Consistent with SQS policy, SNS policy, and EventBridge event_pattern.
    policy = optional(any)

    # Skip AWS's policy-lockout safety check when putting the resource policy.
    # By default AWS rejects a policy that would lock the requesting principal
    # out of future PutFileSystemPolicy calls. Only set this when intentionally
    # deploying such a policy — a locked-out policy can only be fixed by the
    # account root. Requires `policy` to be set.
    bypass_policy_lockout_safety_check = optional(bool, false)

    # Replicate this file system to another region or Availability Zone for
    # disaster recovery. EFS replication is one-per-file-system and create-time
    # immutable: changing the destination replaces the replication configuration
    # (the destination file system itself is never deleted by that replacement).
    replication = optional(object({
      # Destination region for the replica (e.g., "us-east-2"). At least one of
      # `destination_region` or `destination_availability_zone_name` is required.
      # Same-region replication (with a different AZ) is valid.
      destination_region = optional(string, "")

      # Destination Availability Zone name (e.g., "us-east-2a"). Creates the
      # replica as a One Zone file system in that AZ — the cheaper DR shape. At
      # least one of `destination_region` or `destination_availability_zone_name`
      # is required.
      destination_availability_zone_name = optional(string, "")

      # KMS key for the replica's encryption at rest. When omitted, AWS uses the
      # AWS-managed key `aws/elasticfilesystem` in the destination region.
      # Replicas are always encrypted regardless of the source's encryption state.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination_kms_key_id = optional(string, "")

      # Replicate into an EXISTING file system instead of having AWS create the
      # replica. The referenced file system must have
      # `replication_overwrite_protection` set to "DISABLED". When omitted, AWS
      # creates a fresh destination file system.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      destination_file_system_id = optional(string, "")
    }))
  })
}
