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
  description = "GcpAlloydbCluster specification"
  type = object({
    # GCP project where the AlloyDB cluster will be created.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the AlloyDB cluster. This becomes the GCP resource cluster_id.
    # Must start with a lowercase letter, can contain lowercase letters,
    # numbers, and hyphens, and must end with a lowercase letter or number.
    # Maximum 63 characters. Immutable after creation.
    cluster_name = string

    # GCP region where the cluster will be deployed (e.g., "us-central1").
    # Immutable after creation.
    location = string

    # VPC network for Private Service Access connectivity, as the relative
    # resource path "projects/{project}/global/networks/{network}" (the
    # AlloyDB API rejects full https:// self-link URLs). The VPC must have
    # Private Service Access configured (compose GcpGlobalAddress +
    # GcpServiceNetworkingConnection). Mutually exclusive with PSC.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = optional(string, "")

    # Private Service Connect configuration. When psc_enabled is true, network
    # must be unset — the cluster is reachable only through PSC endpoints.
    psc_config = optional(object({
      # When true, the cluster is reachable only through PSC endpoints.
      psc_enabled = optional(bool, false)
    }))

    # Cluster role. PRIMARY (default) is a writable primary; SECONDARY is a
    # cross-region DR replica that follows a primary cluster.
    cluster_type = optional(string)

    # Required when cluster_type is SECONDARY — names the primary cluster.
    secondary_config = optional(object({
      # Full resource name of the primary cluster this secondary follows.
      primary_cluster_name = string
    }))

    # Name of the allocated IP range for Private Service Access.
    # When set, the cluster uses this specific IP range for its private
    # connectivity instead of an auto-allocated range. This is common in
    # enterprise setups where IP ranges are pre-planned.
    allocated_ip_range = optional(string, "")

    # PostgreSQL major version for the cluster.
    # Supported values: "POSTGRES_14", "POSTGRES_15", "POSTGRES_16".
    # If not specified, GCP selects the latest stable version.
    database_version = optional(string, "")

    # Human-readable display name for the cluster.
    display_name = optional(string, "")

    # Initial database user created during cluster provisioning.
    # If not specified, no initial user is created. Access must then be
    # configured via AlloyDB Auth Proxy with IAM authentication.
    initial_user = optional(object({
      # Password for the initial user. Must be at least 8 characters.
      # This value is sensitive and should be handled accordingly.
      password = string

      # Username for the initial user. If not specified, defaults to "postgres"
      # per GCP AlloyDB conventions.
      user = optional(string, "")
    }))

    # Automated backup policy for periodic snapshot backups.
    # When not specified, GCP uses its default policy (enabled, daily, 14-day retention).
    automated_backup_policy = optional(object({
      # Whether automated backups are enabled. Set to false to explicitly
      # disable automated backups.
      enabled = optional(bool, false)

      # Length of the time window during which a backup can be taken.
      # Duration in seconds with 's' suffix, e.g., "3600s" (1 hour).
      # Default: "3600s".
      backup_window = optional(string, "")

      # GCP region where backups will be stored. If not specified,
      # backups are stored in the same region as the cluster.
      location = optional(string, "")

      # Number of backups to retain. Mutually exclusive with
      # time_based_retention_period.
      quantity_based_retention_count = optional(number, 0)

      # How long to retain backups. Duration in seconds with 's' suffix,
      # e.g., "1209600s" (14 days). Mutually exclusive with
      # quantity_based_retention_count.
      time_based_retention_period = optional(string, "")

      # Weekly schedule defining when backups are taken.
      weekly_schedule = optional(object({
        # Days of the week to take backups.
        # If not specified, GCP defaults to daily backups.
        days_of_week = optional(list(string), [])

        # Hour of day (0-23 UTC) to start backups. GCP's TimeOfDay structure
        # is simplified here since minutes/seconds/nanos are always zero for
        # AlloyDB backup schedules.
        start_hour = optional(number, 0)
      }))

      # Cloud KMS key for encrypting automated backups. If not specified,
      # backups use Google-managed encryption. When using a different key
      # from the cluster's encryption, this enables independent backup
      # encryption lifecycle management.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      encryption_kms_key_name = optional(string, "")

      # Labels applied to every backup created by this policy — how backup
      # storage costs are attributed and backup sets are filtered in the
      # console (distinct from the cluster's own labels).
      labels = optional(map(string), {})
    }))

    # Continuous backup configuration for point-in-time recovery (PITR).
    # Enabled by default with a 14-day recovery window.
    continuous_backup_config = optional(object({
      # Whether continuous backup is enabled. Defaults to true.
      # Set to false only if you do not need point-in-time recovery.
      enabled = optional(bool, false)

      # Number of days for which continuous backup data is retained,
      # enabling PITR within this window. Range: 1-35. Default: 14.
      recovery_window_days = optional(number, 0)

      # Cloud KMS key for encrypting continuous backup data. If not specified,
      # continuous backups use Google-managed encryption.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      encryption_kms_key_name = optional(string, "")
    }))

    # Cloud KMS key for encrypting the cluster's data at rest (CMEK).
    # Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
    # If not specified, data is encrypted with Google-managed keys.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Preferred maintenance window for system updates.
    maintenance_window = optional(object({
      # Day of the week for maintenance.
      day = string

      # Hour of day (0-23, UTC) when the maintenance window starts.
      start_hour = optional(number, 0)
    }))

    # Primary instance configuration. Required -- a cluster without a
    # primary instance cannot serve queries.
    primary_instance = object({
      # ID for the primary instance. This becomes the GCP resource name.
      # Must start with a lowercase letter, can contain lowercase letters,
      # numbers, and hyphens, and must end with a lowercase letter or number.
      # Maximum 63 characters. Immutable after creation.
      instance_id = string

      # Number of CPUs for the instance. Valid values: 2, 4, 8, 16, 32, 64, 96, 128.
      # GCP selects the appropriate machine family automatically.
      # Mutually exclusive with machine_type.
      cpu_count = optional(number, 0)

      # Explicit machine type (e.g., "n2-highmem-4", "c4a-highmem-4-lssd").
      # Use this for advanced scenarios where you need a specific machine family.
      # Mutually exclusive with cpu_count.
      machine_type = optional(string, "")

      # Availability type controlling the placement of the instance.
      # ZONAL: single-zone deployment (lower cost, single zone of failure).
      # REGIONAL: multi-zone deployment with automatic failover (recommended
      # for production). Default: REGIONAL when unset.
      availability_type = optional(string, "")

      # PostgreSQL database flags as key-value pairs.
      # These correspond to PostgreSQL server parameters (e.g.,
      # "max_connections", "work_mem", "shared_buffers").
      # See GCP AlloyDB documentation for supported flags.
      database_flags = optional(map(string), {})

      # Human-readable display name for the primary instance.
      display_name = optional(string, "")

      # Query insights configuration for performance monitoring.
      # If not specified, GCP uses default query insights settings
      # (enabled with sensible defaults).
      query_insights_config = optional(object({
        # Number of query execution plans captured per minute.
        # Range: 0-20. Default: 5. Set to 0 to disable plan capture.
        query_plans_per_minute = optional(number, 0)

        # Maximum length of the query string stored in insights.
        # Range: 256-4500. 0 means unset — GCP applies its default (1024).
        # Longer strings help debug complex queries but use more storage.
        query_string_length = optional(number, 0)

        # Whether to record application tags set via
        # pg_stat_statements.track_activity_query_size.
        # Useful for tagging queries by application or feature.
        record_application_tags = optional(bool, false)

        # Whether to record the client IP address for each query.
        # Useful for identifying which application instances generate load.
        record_client_address = optional(bool, false)
      }))

      # Whether to require the AlloyDB Auth Proxy or AlloyDB Language Connectors
      # for all connections. When true, direct IP connections are rejected.
      # This enforces IAM-based authentication for all database access.
      require_connectors = optional(bool, false)

      # SSL mode for client connections.
      # ENCRYPTED_ONLY: all connections must use TLS (recommended for production).
      # ALLOW_UNENCRYPTED_AND_ENCRYPTED: both TLS and plaintext allowed.
      ssl_mode = optional(string, "")

      # Instance activation: ALWAYS keeps the primary running (the default
      # posture); NEVER stops it. Flipping ALWAYS→NEVER→ALWAYS is the
      # stop/start lever — a stopped primary keeps its configuration and
      # storage but serves nothing and stops billing for compute. Stop read
      # pool instances before stopping the primary.
      activation_policy = optional(string, "")

      # Unstructured metadata stored on the primary instance (annotations, not
      # labels — not used for billing filtering). Mutable in place.
      annotations = optional(map(string), {})

      # Pin a ZONAL primary to a specific Compute Engine zone (e.g.
      # "us-central1-a"). Only valid when availability_type is ZONAL — GCP
      # rejects it on REGIONAL instances; leave empty to let GCP pick a zone
      # with available capacity. Mutable: changing it live-migrates the
      # primary to the new zone.
      gce_zone = optional(string, "")

      # AlloyDB managed connection pooling on the primary (built-in pooler).
      # Mutable in place.
      connection_pool_config = optional(object({
        # Turn managed connection pooling on or off. Mutable in place.
        enabled = optional(bool, false)

        # Pooler flags, keyed by flag name WITHOUT the "connection-pooling-"
        # prefix and with underscores instead of dashes (GCP's documented
        # convention for this provider surface): e.g. the flag
        # "connection-pooling-pool-mode" is set as key "pool_mode". Only
        # applied while enabled is true.
        flags = optional(map(string), {})
      }))

      # Enable a public IP on the primary instance. Pair with
      # authorized_external_networks to control who may reach it.
      enable_public_ip = optional(bool, false)

      # Enable outbound public IP for the primary instance.
      enable_outbound_public_ip = optional(bool, false)

      # CIDR ranges allowed to reach the primary's public IP. Requires
      # enable_public_ip.
      authorized_external_networks = optional(list(object({
        cidr_range = string
      })), [])

      # Draw the primary's private IPs from a specific Private Service Access
      # allocated range (RFC 1035 name) instead of the range the cluster uses.
      # Immutable: changing it destroys and recreates the primary instance.
      allocated_ip_range_override = optional(string, "")

      # Private Service Connect configuration for the primary instance —
      # meaningful only on PSC clusters (psc_config.psc_enabled).
      psc_instance_config = optional(object({
        # Consumer project numbers allowed to create PSC endpoints.
        allowed_consumer_projects = optional(list(string), [])

        # PSC service automation connections.
        psc_auto_connections = optional(list(object({
          # Consumer network, e.g. "projects/vpc-host/global/networks/default".
          consumer_network = optional(string, "")

          # Consumer project ID (not project number).
          consumer_project = optional(string, "")
        })), [])

        # PSC interfaces for outbound connectivity (0 or 1 supported by AlloyDB).
        psc_interface_configs = optional(list(object({
          # Network attachment resource in the consumer project.
          network_attachment_resource = optional(string, "")
        })), [])
      }))

      # What happens to the PRIMARY INSTANCE in GCP when this resource is
      # destroyed (the cluster has its own deletion_policy at the spec root).
      #   "DELETE"  -- (GCP's default when unset) the primary is deleted
      #   "PREVENT" -- destroy FAILS while the primary exists
      #   "ABANDON" -- the primary is removed from management but keeps
      #                running (and billing) in GCP
      deletion_policy = optional(string, "")
    })

    # Unstructured metadata stored on the cluster (annotations, not labels —
    # not used for billing filtering).
    annotations = optional(map(string), {})

    # Billing subscription tier: STANDARD (default) or TRIAL. A TRIAL cluster
    # converts to STANDARD when trial credits end.
    subscription_type = optional(string, "")

    # When true, a database_version change returns immediately instead of
    # waiting for the in-place major-version upgrade to finish. The upgrade
    # continues server-side; use for very large clusters where the wait
    # exceeds sane IaC timeouts.
    skip_await_major_version_upgrade = optional(bool, false)

    # User-defined labels on the cluster and its bundled primary instance
    # (cost attribution, team ownership, environment tagging). Merged with
    # the platform's attribution labels; on key conflicts the platform
    # labels win. Mutable in place.
    labels = optional(map(string), {})

    # Dataplex Universal Catalog integration (automatic metadata discovery).
    # GCP enables it by default when this block is absent; set
    # enabled: false to opt out explicitly.
    dataplex_config = optional(object({
      # Whether Dataplex integration is enabled for the cluster. Mutable.
      # Required inside dataplex_config: declaring the block takes the
      # integration under management, and the switch says which way.
      enabled = bool
    }))

    # Seed the new cluster from an AlloyDB backup. At most one restore
    # source may be set; all restore sources are create-time only (changing
    # one destroys and recreates the cluster).
    restore_backup_source = optional(object({
      # Full resource name of the source backup, e.g.
      # "projects/{project}/locations/{location}/backups/{backup}".
      backup_name = string
    }))

    # Seed the new cluster by point-in-time recovery from a source cluster's
    # continuous backup stream. At most one restore source may be set;
    # create-time only.
    restore_continuous_backup_source = optional(object({
      # The source cluster to restore from — its full resource name
      # "projects/{project}/locations/{location}/clusters/{cluster}", or a
      # reference to a GcpAlloydbCluster resource. Restore provenance, not
      # physical placement: the new cluster is seeded FROM the source, never
      # contained IN it (containment_exempt).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cluster = string

      # The point in time to restore to, in RFC 3339 format (e.g.
      # "2026-08-01T12:00:00Z"). Must fall inside the source cluster's
      # continuous-backup recovery window.
      point_in_time = string
    }))

    # Seed the new cluster from a Backup and DR Service backup. At most one
    # restore source may be set; create-time only.
    restore_backupdr_backup_source = optional(object({
      # Full resource name of the Backup and DR backup, in the format
      # "projects/{project}/locations/{location}/backupVaults/{vault}/dataSources/{dataSource}/backups/{backup}".
      backup = string
    }))

    # Seed the new cluster by point-in-time recovery through the Backup and
    # DR Service. At most one restore source may be set; create-time only.
    restore_backupdr_pitr_source = optional(object({
      # Full resource name of the Backup and DR data source, in the format
      # "projects/{project}/locations/{location}/backupVaults/{vault}/dataSources/{dataSource}".
      data_source = string

      # The point in time to restore to, in RFC 3339 format.
      point_in_time = string
    }))

    # Client-side destroy guard: while true (GCP's default), any destroy —
    # including the platform's own teardown flows — FAILS until this field
    # is flipped to false and applied. Both engines always send the value
    # explicitly, so the spec is the single source of truth. Note the
    # ordering quirk in the provider: deletion_policy ABANDON is evaluated
    # BEFORE this guard, so abandoning a protected cluster still works.
    deletion_protection = optional(bool)

    # What happens to the CLUSTER in GCP when this resource is destroyed
    # (the bundled primary has its own deletion_policy under
    # primary_instance). AlloyDB clusters use a different value set from
    # most GCP resources:
    #   "DEFAULT" -- (GCP's default when unset) the cluster is deleted;
    #                the API rejects the delete while any instance other
    #                than the bundled primary still exists
    #   "FORCE"   -- the cluster AND every instance still in it are
    #                deleted — required when destroying a SECONDARY
    #                cluster that has a secondary instance
    #   "PREVENT" -- destroy FAILS; the strongest guard, evaluated even
    #                before deletion_protection
    #   "ABANDON" -- the cluster is removed from management but keeps
    #                running (and billing) in GCP; bypasses
    #                deletion_protection
    deletion_policy = optional(string, "")
  })
}
