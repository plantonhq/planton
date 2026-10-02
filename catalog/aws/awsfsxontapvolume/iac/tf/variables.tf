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
  description = "AwsFsxOntapVolume specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The ID of the Storage Virtual Machine that this volume belongs to. Required.
    # ForceNew — the volume cannot be moved to a different SVM after creation.
    #
    # The SVM provides the network endpoints, protocol configuration, and Active
    # Directory integration. All volumes within an SVM share its protocol stack.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    storage_virtual_machine_id = string

    # The name of the volume within the ONTAP file system. Required. ForceNew.
    #
    # This is the ONTAP volume name (not the Planton metadata name). ONTAP volume
    # names must be alphanumeric plus underscores — hyphens are not allowed. This
    # name appears in junction paths, SnapMirror relationships, and ONTAP CLI
    # output.
    #
    # Constraints: 1-203 characters, alphanumeric and underscore only.
    name = string

    # The size of the volume in megabytes. Exactly one of `size_in_megabytes`
    # and `size_in_bytes` must be set.
    #
    # Minimum 20 MB. Maximum is constrained by the file system's total storage
    # capacity. ONTAP volumes support thin provisioning, so the logical size can
    # exceed the physical capacity available — ONTAP handles overcommit at the
    # aggregate level. Size can be increased or decreased in place.
    #
    # Covers volumes up to ~2 PiB (int32 megabytes). For larger volumes or
    # byte-precise sizing, use `size_in_bytes` instead.
    size_in_megabytes = optional(number)

    # The size of the volume in bytes — the byte-precise sizing arm, and the
    # only way to size volumes beyond 2 PiB (FlexGroup volumes scale to ~20 PiB).
    # Exactly one of `size_in_megabytes` and `size_in_bytes` must be set.
    #
    # Maximum: 22,517,998,000,000,000 bytes (~20 PiB). FLEXGROUP volumes
    # require at least 100 GiB per constituent.
    size_in_bytes = optional(number)

    # The location in the SVM namespace where this volume is mounted. Clients
    # access the volume at this path (e.g., mount nfs.svm.example.com:/vol1).
    #
    # If omitted, the volume is created but not mounted — it exists in ONTAP
    # but is not accessible via NFS/SMB until a junction path is set.
    #
    # Must start with "/" and be unique within the SVM. Examples: "/vol1",
    # "/data/prod", "/shares/finance". Can be changed after creation (the
    # volume remounts at the new path).
    #
    # Constraints: 1-255 characters.
    junction_path = optional(string, "")

    # The ONTAP volume type. ForceNew.
    #
    # - "RW": Read-write volume. The standard type for serving data to clients.
    # - "DP": Data protection volume. A read-only destination for SnapMirror
    #   replication. DP volumes cannot be mounted until the SnapMirror
    #   relationship is broken or the volume is converted.
    #
    # Default: RW
    ontap_volume_type = optional(string)

    # The volume style. ForceNew.
    #
    # - "FLEXVOL": Traditional ONTAP volume on a single aggregate. Suitable for
    #   most workloads. Simpler operations and faster metadata performance.
    # - "FLEXGROUP": A volume distributed across multiple aggregates for high
    #   throughput and large-scale workloads (hundreds of TBs to PBs). Requires
    #   aggregate_configuration. Ideal for data lakes, genomics, and media.
    #
    # Default: FLEXVOL
    volume_style = optional(string)

    # The security style for this volume's root directory. Controls how file
    # permissions are evaluated. Can be changed after creation.
    #
    # - "UNIX": UNIX permissions (mode bits, uid/gid). Best for Linux/NFS.
    # - "NTFS": Windows ACLs. Best for Windows/SMB with Active Directory.
    # - "MIXED": Both permission systems coexist. The effective security style
    #   depends on which protocol last set permissions on a file.
    #
    # If omitted, inherits from the parent SVM's root_volume_security_style.
    security_style = optional(string, "")

    # The name of the ONTAP snapshot policy to apply to this volume. Snapshot
    # policies control automatic snapshot creation and retention. Can be changed
    # after creation.
    #
    # Common policies: "default" (6 hourly + 2 daily + 2 weekly), "none"
    # (no automatic snapshots). Custom policies can be created via the ONTAP CLI.
    #
    # Constraints: 1-255 characters.
    snapshot_policy = optional(string, "")

    # Enable ONTAP storage efficiency features: deduplication, compression, and
    # compaction. These features reduce physical storage consumption by
    # identifying and eliminating redundant data blocks. Can be changed after
    # creation.
    #
    # Recommended for most workloads (set true). Disable only for workloads
    # that are already compressed or deduplicated (e.g., encrypted data,
    # pre-compressed media files) where the CPU overhead provides no benefit.
    # When omitted, ONTAP applies its own per-volume-type default.
    storage_efficiency_enabled = optional(bool)

    # Whether to copy resource tags to automatic volume backups.
    #
    # Default: false
    copy_tags_to_backups = optional(bool)

    # Whether to skip the automatic backup that AWS takes when the volume is
    # deleted. Set to true for development/test volumes where the backup is
    # unnecessary.
    #
    # Default: false (a final backup is taken)
    skip_final_backup = optional(bool)

    # Tags applied to the final backup taken on deletion. Only meaningful when
    # skip_final_backup is false.
    final_backup_tags = optional(map(string), {})

    # Whether to allow deletion of a SnapLock Enterprise volume that contains
    # WORM files with unexpired retention periods. Only relevant for SnapLock
    # Enterprise volumes — Compliance volumes can never bypass retention.
    #
    # Default: false
    bypass_snaplock_enterprise_retention = optional(bool)

    # Data tiering policy that controls when and how data moves from primary
    # SSD storage to lower-cost capacity pool storage. If omitted, the volume
    # uses the default tiering policy (SNAPSHOT_ONLY). Can be changed after
    # creation.
    tiering_policy = optional(object({
      # The tiering policy name. Required when a tiering policy is declared —
      # declaring the block is choosing a policy; to keep AWS's default
      # (SNAPSHOT_ONLY), omit tiering_policy entirely.
      #
      # - "NONE": All data remains on primary SSD storage. No tiering. Use for
      #   latency-sensitive workloads where all data must be instantly accessible.
      # - "SNAPSHOT_ONLY": Only snapshot data (point-in-time copies) is tiered.
      #   Active file system data stays on SSD. The safest tiering option.
      # - "AUTO": Data not accessed for the cooling period is automatically tiered.
      #   The most cost-effective option for mixed-access workloads.
      # - "ALL": All data (including active data) is stored on capacity pool. Only
      #   metadata stays on SSD. Lowest cost, highest latency for first access.
      name = optional(string, "")

      # The number of days before data is considered "cold" and eligible for
      # tiering to capacity pool storage. Only applicable when name is "AUTO" or
      # "SNAPSHOT_ONLY".
      #
      # Range: 2-183 days. Lower values tier data more aggressively (lower cost,
      # potentially higher latency for recently accessed data).
      cooling_period = optional(number, 0)
    }))

    # SnapLock configuration for WORM (Write Once Read Many) compliance storage.
    # When configured, files committed to this volume become immutable for their
    # retention period. ForceNew for snaplock_type.
    #
    # SnapLock has two modes:
    # - ENTERPRISE: Admins can delete WORM files before retention expiry (if
    #   privileged_delete is enabled). Suitable for internal governance.
    # - COMPLIANCE: No one — not even the root/admin user or AWS support — can
    #   delete WORM files before retention expiry. Required for SEC 17a-4, HIPAA,
    #   and similar regulations.
    #
    # Once set, the snaplock_type cannot be changed. Choosing the wrong type
    # requires deleting and recreating the volume.
    snaplock_configuration = optional(object({
      # The SnapLock retention mode. Required. ForceNew — cannot be changed after
      # volume creation.
      #
      # - "ENTERPRISE": Administrative deletion of WORM files is possible (if
      #   privileged_delete is enabled). Suitable for internal governance policies
      #   where an escape hatch is acceptable.
      # - "COMPLIANCE": Immutable. No one can delete WORM files before retention
      #   expiry — not the root user, not AWS Support, not even the account owner.
      #   Required for SEC 17a-4 and similar strict regulatory mandates.
      snaplock_type = string

      # Whether this volume is designated as the SnapLock audit log volume.
      # A single audit log volume per SVM records all SnapLock operations
      # (file commits, retention changes, privileged deletions).
      #
      # Default: false
      audit_log_volume = optional(bool)

      # Controls whether privileged deletion of WORM files is allowed before their
      # retention period expires. Only meaningful for ENTERPRISE SnapLock.
      #
      # - "DISABLED": Privileged delete is not allowed (default).
      # - "ENABLED": Administrators can delete WORM files early.
      # - "PERMANENTLY_DISABLED": Privileged delete is permanently disabled and
      #   cannot be re-enabled. Use this for Enterprise volumes that must never
      #   allow early deletion.
      #
      # Default: DISABLED
      privileged_delete = optional(string)

      # Whether volume-append mode is enabled. When enabled, files can be appended
      # to (new data added at the end) even after being committed to WORM state.
      # The existing content remains immutable. Useful for log files and audit
      # trails that need continuous appending.
      #
      # Default: false
      volume_append_mode_enabled = optional(bool)

      # Configures automatic commitment of files to WORM state after a period of
      # inactivity. When autocommit is configured, files that have not been modified
      # for the specified duration are automatically transitioned to WORM state.
      #
      # This eliminates the need for applications to explicitly commit files.
      autocommit_period = optional(object({
        # The unit of time for the autocommit period. Required when autocommit_period
        # is declared — the AWS API's AutocommitPeriod.Type member is required, and a
        # unit-less period cannot be expressed. To disable autocommit, set "NONE"
        # (or omit autocommit_period entirely; NONE is also AWS's default).
        #
        # - "NONE": Autocommit is disabled.
        # - "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": The time unit for the
        #   value field.
        type = optional(string, "")

        # The number of time units before an unmodified file is auto-committed to
        # WORM state. Required for every unit type except "NONE" (which takes no
        # value). Each unit carries its own AWS range:
        #
        # - MINUTES: 5-65,535
        # - HOURS: 1-65,535
        # - DAYS: 1-3,650
        # - MONTHS: 1-120
        # - YEARS: 1-10
        value = optional(number, 0)
      }))

      # Configures the default, minimum, and maximum retention periods for WORM
      # files on this volume. These bounds constrain how long files must be retained
      # and provide guardrails for retention policy enforcement.
      retention_period = optional(object({
        # The default retention period applied to files committed to WORM state
        # without an explicit retention period.
        default_retention = optional(object({
          # The unit of time for the retention duration. Required when the duration is
          # declared — the AWS API's RetentionPeriod.Type member is required, and a
          # unit-less duration cannot be expressed.
          #
          # - "SECONDS", "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": Standard
          #   time units. The value field specifies the count.
          # - "INFINITE": Files are retained forever. The value field takes no value.
          # - "UNSPECIFIED": No retention period is set. The value field takes no value.
          type = optional(string, "")

          # The number of time units for the retention duration. Each unit carries its
          # own AWS range (0 is legal and means a zero-length duration):
          #
          # - SECONDS: 0-65,535
          # - MINUTES: 0-65,535
          # - HOURS: 0-24
          # - DAYS: 0-365
          # - MONTHS: 0-12
          # - YEARS: 0-100
          #
          # INFINITE and UNSPECIFIED take no value.
          value = optional(number, 0)
        }))

        # The minimum retention period. Files cannot have retention periods shorter
        # than this value.
        minimum_retention = optional(object({
          # The unit of time for the retention duration. Required when the duration is
          # declared — the AWS API's RetentionPeriod.Type member is required, and a
          # unit-less duration cannot be expressed.
          #
          # - "SECONDS", "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": Standard
          #   time units. The value field specifies the count.
          # - "INFINITE": Files are retained forever. The value field takes no value.
          # - "UNSPECIFIED": No retention period is set. The value field takes no value.
          type = optional(string, "")

          # The number of time units for the retention duration. Each unit carries its
          # own AWS range (0 is legal and means a zero-length duration):
          #
          # - SECONDS: 0-65,535
          # - MINUTES: 0-65,535
          # - HOURS: 0-24
          # - DAYS: 0-365
          # - MONTHS: 0-12
          # - YEARS: 0-100
          #
          # INFINITE and UNSPECIFIED take no value.
          value = optional(number, 0)
        }))

        # The maximum retention period. Files cannot have retention periods longer
        # than this value.
        maximum_retention = optional(object({
          # The unit of time for the retention duration. Required when the duration is
          # declared — the AWS API's RetentionPeriod.Type member is required, and a
          # unit-less duration cannot be expressed.
          #
          # - "SECONDS", "MINUTES", "HOURS", "DAYS", "MONTHS", "YEARS": Standard
          #   time units. The value field specifies the count.
          # - "INFINITE": Files are retained forever. The value field takes no value.
          # - "UNSPECIFIED": No retention period is set. The value field takes no value.
          type = optional(string, "")

          # The number of time units for the retention duration. Each unit carries its
          # own AWS range (0 is legal and means a zero-length duration):
          #
          # - SECONDS: 0-65,535
          # - MINUTES: 0-65,535
          # - HOURS: 0-24
          # - DAYS: 0-365
          # - MONTHS: 0-12
          # - YEARS: 0-100
          #
          # INFINITE and UNSPECIFIED take no value.
          value = optional(number, 0)
        }))
      }))
    }))

    # Aggregate configuration for FLEXGROUP volumes. Controls how the volume is
    # distributed across the file system's aggregates. Ignored for FLEXVOL
    # volumes.
    #
    # All fields in this block are ForceNew — changing the aggregate layout
    # requires recreating the volume.
    aggregate_configuration = optional(object({
      # The list of aggregate names to use for the FlexGroup volume. Each name
      # must match the pattern "aggr" followed by 1-2 digits (e.g., "aggr1",
      # "aggr2") — aggregate numbering follows the file system's HA pairs.
      # Maximum 12 aggregates.
      #
      # ForceNew — changing this requires volume recreation.
      aggregates = optional(list(string), [])

      # The number of FlexGroup constituents (member volumes) to create per
      # aggregate. The total number of constituents equals
      # constituents_per_aggregate * len(aggregates).
      #
      # Higher values increase parallelism but also metadata overhead. Default
      # in AWS is typically 8.
      #
      # Range: 1-200. ForceNew.
      constituents_per_aggregate = optional(number, 0)
    }))
  })
}
