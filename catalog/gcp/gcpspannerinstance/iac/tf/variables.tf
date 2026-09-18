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
  description = "GcpSpannerInstance specification"
  type = object({
    # The GCP project the Spanner instance is created in. Accepts a literal
    # project ID or a reference to a GcpProject resource. If omitted, the
    # provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Unique identifier of the Spanner instance in GCP. Immutable. If not
    # specified, defaults to metadata.name. Must be 6-30 characters: start
    # with a lowercase letter, contain only lowercase letters, digits, and
    # hyphens, and end with a letter or digit. This is the value downstream
    # databases and backup schedules reference.
    instance_name = optional(string, "")

    # Instance configuration defining geographic placement and replication
    # topology. Immutable. Regional configs (e.g. "regional-us-central1")
    # give the lowest latency in one region; multi-region configs (e.g.
    # "nam6", "nam-eur-asia1") replicate across regions for higher
    # availability and enable the 99.999% SLA on ENTERPRISE_PLUS. This is
    # the single most consequential choice on the instance — it cannot be
    # changed without recreating the instance and moving the data.
    config = string

    # Human-readable display name shown in the GCP console.
    # Must be 4-30 characters.
    display_name = string

    # Labels applied to the instance for cost attribution and organization.
    # Merged with Planton's platform labels (which win on key conflicts).
    labels = optional(map(string), {})

    # Number of nodes allocated to the instance. Each node provides roughly
    # 10,000 QPS of reads or 2,000 QPS of writes and 10 TB of storage.
    # Mutable — capacity changes apply online. Mutually exclusive with
    # processing_units and autoscaling_config; must not be set for
    # FREE_INSTANCE. If no capacity field is set for a PROVISIONED instance,
    # GCP defaults to 1 node.
    num_nodes = optional(number, 0)

    # Number of processing units allocated to the instance — the
    # fine-grained alternative to nodes (1 node = 1000 processing units).
    # Values below 1000 must be multiples of 100 (the smallest billable
    # Spanner footprint is 100). Mutable. Mutually exclusive with num_nodes
    # and autoscaling_config; must not be set for FREE_INSTANCE.
    processing_units = optional(number, 0)

    # Managed autoscaling: Spanner adjusts capacity within bounds based on
    # CPU and storage utilization, including per-replica overrides for
    # multi-region instances. Mutually exclusive with num_nodes and
    # processing_units; must not be set for FREE_INSTANCE.
    autoscaling_config = optional(object({
      # Required. The floor and ceiling the autoscaler operates within.
      autoscaling_limits = object({
        # Minimum number of nodes the autoscaler may scale down to. Use together
        # with max_nodes. Each node is 1000 processing units and provides roughly
        # 10,000 QPS of reads or 2,000 QPS of writes plus 10 TB of storage.
        min_nodes = optional(number, 0)

        # Maximum number of nodes the autoscaler may scale up to.
        # Must be >= min_nodes.
        max_nodes = optional(number, 0)

        # Minimum number of processing units the autoscaler may scale down to.
        # Use together with max_processing_units. Values below 1000 must be
        # multiples of 100; values 1000 and above must be multiples of 1000.
        min_processing_units = optional(number, 0)

        # Maximum number of processing units the autoscaler may scale up to.
        # Must be >= min_processing_units.
        max_processing_units = optional(number, 0)
      })

      # Utilization targets that trigger scaling decisions.
      # If not set, GCP uses default targets.
      autoscaling_targets = optional(object({
        # Target percentage of high-priority CPU utilization (user reads/writes,
        # as opposed to background maintenance work). When the instance runs
        # hotter than this, the autoscaler adds capacity. Google recommends 65
        # for regional configurations and lower (around 45) for multi-region
        # configurations, where replication headroom matters during failover.
        high_priority_cpu_utilization_percent = optional(number, 0)

        # Target percentage of storage utilization. When storage crosses this
        # threshold, the autoscaler adds capacity regardless of CPU (Spanner
        # couples storage capacity to compute capacity). Google recommends 80
        # to leave headroom for growth spikes.
        storage_utilization_percent = optional(number, 0)

        # Target percentage of TOTAL CPU utilization — user traffic plus
        # Spanner's background maintenance work — as a second CPU signal
        # alongside high_priority_cpu_utilization_percent. Useful when
        # background work (change streams, backups) is a meaningful share of
        # the load and scaling on user traffic alone would run the instance
        # hot. Scale 0 (no utilization) to 100 (full utilization).
        total_cpu_utilization_percent = optional(number, 0)
      }))

      # Per-replica-location overrides for multi-region instances: scale a
      # read-heavy region independently instead of sizing every region for the
      # hottest one. Each entry selects one replica location and overrides its
      # capacity bounds, CPU targets, or both.
      asymmetric_autoscaling_options = optional(list(object({
        # The replica location (a region of the multi-region configuration,
        # e.g. "europe-west1") whose autoscaling this option overrides.
        replica_location = string

        # The per-replica tuning applied at this location: capacity bounds
        # (nodes or processing units) replacing the instance-wide
        # autoscaling_limits, replica-local CPU targets, or switches disabling
        # a CPU signal for this replica.
        overrides = object({
          # Minimum number of nodes for the selected replica location.
          # Use together with max_nodes; mutually exclusive with the
          # processing-unit bounds.
          min_nodes = optional(number, 0)

          # Maximum number of nodes for the selected replica location.
          # Must be >= min_nodes.
          max_nodes = optional(number, 0)

          # Minimum number of processing units for the selected replica
          # location — the fine-grained alternative to node bounds. Per-replica
          # PU bounds must be multiples of 1000 (unlike instance-wide limits,
          # which allow multiples of 100 below 1000). Use together with
          # max_processing_units; mutually exclusive with the node bounds.
          min_processing_units = optional(number, 0)

          # Maximum number of processing units for the selected replica
          # location. Must be >= min_processing_units and a multiple of 1000.
          max_processing_units = optional(number, 0)

          # Target high-priority CPU utilization percentage for THIS replica,
          # overriding the instance-wide autoscaling_targets value there.
          # Scale 0-100.
          autoscaling_target_high_priority_cpu_utilization_percent = optional(number, 0)

          # Target total CPU utilization percentage for THIS replica, overriding
          # the instance-wide autoscaling_targets value there. Scale 0-100.
          autoscaling_target_total_cpu_utilization_percent = optional(number, 0)

          # Disable high-priority CPU autoscaling for this replica: the
          # instance-wide high_priority_cpu_utilization_percent target is
          # ignored there and only the remaining signals drive scaling.
          disable_high_priority_cpu_autoscaling = optional(bool, false)

          # Disable total CPU autoscaling for this replica: the instance-wide
          # total_cpu_utilization_percent target is ignored there.
          disable_total_cpu_autoscaling = optional(bool, false)
        })
      })), [])
    }))

    # Instance type. PROVISIONED (default) requires explicit capacity via
    # one of the capacity fields. FREE_INSTANCE provisions the billing
    # account's one zero-cost development instance and must not set any
    # capacity field. Upgrading FREE_INSTANCE → PROVISIONED works in place;
    # the reverse does not exist.
    instance_type = optional(string, "")

    # Edition controlling available features and SLA level. Mutable —
    # upgrades apply in place. STANDARD: cost-optimized, single-region
    # feature set. ENTERPRISE: adds granular sizing, asymmetric autoscaling,
    # and incremental backups. ENTERPRISE_PLUS: 99.999% multi-region SLA and
    # advanced compliance features. Cannot be set for FREE_INSTANCE.
    edition = optional(string, "")

    # Default backup schedule type for NEW databases created on this
    # instance. NONE (default): new databases get no automatic backup
    # schedule. AUTOMATIC: GCP attaches a default backup schedule to each
    # new database (explicit GcpSpannerBackupSchedule resources give full
    # control instead). Cannot be AUTOMATIC for FREE_INSTANCE.
    default_backup_schedule_type = optional(string, "")

    # Whether destroying the instance also deletes all backups held on it.
    # When false (default), destroy fails if any database on the instance
    # has backups — a safety net against losing the last restore point. Set
    # true only when the backups are intentionally disposable.
    force_destroy = optional(bool, false)

    # Deletion policy for the instance — what happens when this resource
    # is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the instance is deleted, taking every database and
    #                backup on it with it (backups additionally gated by
    #                force_destroy above)
    #   "PREVENT" -- destroy FAILS; protects the instance every database
    #                in the topology depends on
    #   "ABANDON" -- the instance is removed from management but left
    #                running (and billing) in GCP with its data intact
    deletion_policy = optional(string, "")
  })
}
