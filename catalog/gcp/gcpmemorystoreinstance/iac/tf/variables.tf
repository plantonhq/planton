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
  description = "GcpMemorystoreInstance specification"
  type = object({
    # GCP project where the Memorystore instance will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Memorystore instance. This becomes the GCP resource name.
    # Must start with a lowercase letter, contain only lowercase letters,
    # numbers, and hyphens, and end with a lowercase letter or number.
    # 4-63 characters. Immutable after creation.
    instance_name = string

    # GCP region where the instance will be deployed (e.g., "us-central1").
    # Immutable after creation.
    location = string

    # Number of shards for the instance. Each shard handles a portion of
    # the keyspace. Minimum 1 shard.
    #
    # For CLUSTER mode: multiple shards distribute data across nodes.
    # For CLUSTER_DISABLED mode: typically 1 shard (single primary).
    shard_count = number

    # Instance mode controlling cluster topology.
    # CLUSTER: sharded mode with native cluster protocol support.
    #   Clients must use cluster-aware drivers.
    # CLUSTER_DISABLED: standalone mode with a single primary endpoint.
    #   Compatible with any Valkey/Redis client.
    # Immutable after creation.
    mode = optional(string, "")

    # Predefined node type determining CPU and memory per node.
    # Shared-core and custom (burstable, dev/test tiers, smallest first):
    #   SHARED_CORE_NANO, CUSTOM_PICO, CUSTOM_MICRO, CUSTOM_MINI.
    # Dedicated-core (production tiers):
    #   STANDARD_SMALL, STANDARD_LARGE — balanced CPU:memory;
    #   HIGHCPU_MEDIUM — compute-leaning;
    #   HIGHMEM_MEDIUM, HIGHMEM_XLARGE, HIGHMEM_2XLARGE — memory-leaning,
    #   for large keyspaces.
    # If not specified, GCP selects a default.
    node_type = optional(string, "")

    # Engine version (e.g., "VALKEY_8_0", "VALKEY_7_2").
    # If not specified, the latest supported version is used.
    engine_version = optional(string, "")

    # Engine configuration parameters as key-value pairs.
    # See Valkey/Redis configuration reference for supported parameters
    # (e.g., "maxmemory-policy", "notify-keyspace-events").
    engine_configs = optional(map(string), {})

    # Number of read replicas per shard (0-5). Default: 0 (no replicas).
    # Replicas provide read scaling and automatic failover.
    replica_count = optional(number, 0)

    # Private Service Connect (PSC) endpoints for VPC connectivity.
    # Each entry creates a PSC endpoint in the specified consumer VPC,
    # allowing applications in that VPC to reach the instance.
    #
    # A GcpServiceConnectionPolicy for the gcp-memorystore service class
    # must exist on each network in this region before the instance is
    # created — the connectivity automation refuses to place endpoints
    # without it.
    #
    # At least one PSC connection is recommended for the instance to be
    # reachable. Multiple connections enable cross-project or multi-VPC access.
    # Immutable after creation.
    psc_auto_connections = optional(list(object({
      # Consumer VPC network where the PSC endpoint will be created.
      # The API requires the relative resource path
      # (projects/{project_id}/global/networks/{network_id}) — full https://
      # self-link URLs are rejected, so the reference resolves to the
      # GcpVpcNetwork's network_id output, which is already in that form.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = string

      # Consumer project ID where the PSC endpoint will be created.
      # Usually the same project as the Memorystore instance, but can differ
      # for cross-project connectivity. If omitted, both engines resolve the
      # provider's effective project — the endpoint lands next to the
      # instance, which is the common case.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")
    })), [])

    # Authentication mode for client connections.
    # AUTH_DISABLED: no authentication required (default).
    # IAM_AUTH: clients authenticate using GCP IAM credentials.
    # Immutable after creation.
    authorization_mode = optional(string, "")

    # TLS encryption mode for client-to-server traffic.
    # TRANSIT_ENCRYPTION_DISABLED: no encryption (default).
    # SERVER_AUTHENTICATION: clients verify the server's identity via TLS.
    # Immutable after creation.
    transit_encryption_mode = optional(string, "")

    # Cloud KMS key for customer-managed encryption at rest (CMEK).
    # Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
    # If not specified, data is encrypted with Google-managed keys.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key = optional(string, "")

    # Persistence configuration for data durability.
    # Controls whether and how data is written to disk.
    persistence_config = optional(object({
      # Persistence mode.
      # DISABLED: no persistence, data is in-memory only.
      # RDB: periodic point-in-time snapshots.
      # AOF: append-only file logging every write.
      mode = string

      # RDB snapshot configuration. Required when mode is RDB.
      rdb_config = optional(object({
        # How often RDB snapshots are taken.
        rdb_snapshot_period = string

        # Optional RFC3339 timestamp for when to start the first snapshot.
        # If not specified, GCP picks an appropriate time.
        rdb_snapshot_start_time = optional(string, "")
      }))

      # AOF configuration. Required when mode is AOF.
      aof_config = optional(object({
        # How often the AOF buffer is flushed to disk.
        # NEVER: OS decides (best performance, risk of data loss on crash).
        # EVERY_SEC: flush once per second (good balance).
        # ALWAYS: flush on every write (strongest durability, lowest performance).
        append_fsync = string
      }))
    }))

    # Zone distribution configuration.
    # Controls how nodes are spread across availability zones.
    # Immutable after creation.
    zone_distribution_config = optional(object({
      # Zone distribution mode.
      # MULTI_ZONE: nodes spread across multiple zones for high availability (default).
      # SINGLE_ZONE: all nodes in a single zone for lowest latency.
      mode = string

      # Zone for SINGLE_ZONE mode (e.g., "us-central1-a").
      # Required when mode is SINGLE_ZONE. Ignored for MULTI_ZONE.
      zone = optional(string, "")
    }))

    # Maintenance policy for scheduled maintenance windows.
    maintenance_policy = optional(object({
      # Weekly maintenance window schedule.
      weekly_maintenance_window = object({
        # Day of the week for the maintenance window.
        day = string

        # Hour of day (0-23, UTC) when the maintenance window starts. The window
        # always starts on the hour — the API supports no finer granularity.
        hour = optional(number, 0)
      })
    }))

    # Automated backup configuration.
    # When configured, GCP takes daily backups at the specified hour
    # and retains them for the specified duration.
    automated_backup_config = optional(object({
      # Hour of day (0-23, UTC) when the daily backup starts.
      start_hour = optional(number, 0)

      # Backup retention duration in seconds.
      # Minimum: 86400s (1 day). Maximum: 31536000s (365 days).
      # Example: "3024000s" for 35 days.
      retention = string
    }))

    # Cross-region replication for disaster recovery: make this instance a
    # PRIMARY replicating to secondaries in other regions, or a SECONDARY
    # continuously replicating from a primary. Omit (or role NONE) for a
    # standalone instance.
    cross_instance_replication_config = optional(object({
      # This instance's role in the replication topology.
      # NONE: not participating in cross-instance replication.
      # PRIMARY: serves writes; replicates to the listed secondaries.
      # SECONDARY: read-only replica of primary_instance.
      instance_role = string

      # The primary this instance replicates from. Required when
      # instance_role is SECONDARY; must be unset otherwise.
      primary_instance = optional(object({
        # Full resource path of the primary instance
        # (projects/{project}/locations/{location}/instances/{instance}).
        # A reference resolves to another GcpMemorystoreInstance's name output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        instance = string
      }))

      # The secondaries replicating from this instance. Set when
      # instance_role is PRIMARY; must be empty otherwise.
      secondary_instances = optional(list(object({
        # Full resource path of the secondary instance
        # (projects/{project}/locations/{location}/instances/{instance}).
        # A reference resolves to another GcpMemorystoreInstance's name output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        instance = string
      })), [])
    }))

    # Seed the new instance's data from RDB files in Cloud Storage at
    # creation time. Mutually exclusive with managed_backup_source.
    # Immutable: seeding only happens at creation.
    gcs_source = optional(object({
      # Cloud Storage URIs of RDB files to import (gs://bucket/path.rdb).
      # The Memorystore service agent needs read access to the objects.
      uris = list(string)
    }))

    # Seed the new instance's data from an existing managed backup at
    # creation time. Mutually exclusive with gcs_source.
    # Immutable: seeding only happens at creation.
    managed_backup_source = optional(object({
      # Full resource path of the backup to restore from
      # (projects/{project}/locations/{location}/backupCollections/{collection}/backups/{backup}).
      backup = string
    }))

    # User-defined labels to organize and track the instance. Merged
    # beneath Planton's platform attribution labels (platform keys win on
    # conflict).
    labels = optional(map(string), {})

    # Whether deletion protection is enabled. When true (the default —
    # matching GCP's safety posture), destroying the instance fails until
    # this is explicitly set to false. Both IaC engines send the value
    # explicitly so destroy behavior is identical regardless of engine.
    deletion_protection_enabled = optional(bool)

    # Server certificate authority mode for the TLS-enabled instance —
    # which CA signs the server certificate clients verify:
    #   ""                             -- GCP default (GOOGLE_MANAGED_PER_INSTANCE_CA)
    #   "GOOGLE_MANAGED_PER_INSTANCE_CA" -- a Google-managed CA unique to
    #                                       this instance
    #   "GOOGLE_MANAGED_SHARED_CA"       -- a Google-managed CA shared
    #                                       across instances (clients trust
    #                                       one CA for a whole fleet)
    #   "CUSTOMER_MANAGED_CAS_CA"        -- your own CA pool in Certificate
    #                                       Authority Service (pair with
    #                                       server_ca_pool)
    # Meaningful with transit_encryption_mode SERVER_AUTHENTICATION.
    # Immutable after creation.
    server_ca_mode = optional(string, "")

    # The Certificate Authority Service CA pool that signs the server
    # certificate when server_ca_mode is CUSTOMER_MANAGED_CAS_CA.
    # Format: projects/{project}/locations/{region}/caPools/{caPoolId}.
    # Immutable after creation.
    server_ca_pool = optional(string, "")

    # Self-service maintenance version. Setting this to a newer available
    # version triggers the maintenance update on your schedule instead of
    # waiting for GCP's rollout — the lever for applying a security patch
    # immediately. Only settable as an UPDATE to an existing instance, and
    # only forward (downgrades are rejected). Leave unset to follow GCP's
    # automatic rollout.
    maintenance_version = optional(string, "")

    # Deletion policy for the instance — what happens when this resource
    # is destroyed (evaluated only after deletion_protection_enabled allows
    # the destroy at all):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the instance is deleted; all in-memory data is lost
    #   "PREVENT" -- destroy FAILS; a second, independent guard for a
    #                cache whose loss would stampede the backing store
    #   "ABANDON" -- the instance is removed from management but left
    #                running (and billing) in GCP with its data intact
    deletion_policy = optional(string, "")

    # The Memorystore ACL policy attached to the instance: a set of
    # Valkey ACL rules (users, key patterns, allowed commands) authored once
    # and shared across instances in the same region. Leave empty for the
    # instance's built-in default ACL (the "default" user with full access,
    # gated only by auth_enabled). Full resource name:
    # projects/{project}/locations/{region}/aclPolicies/{aclPolicyId}.
    # Mutable: attaching or swapping a policy is an in-place update; the
    # instance's is_acl_policy_in_sync status reports when the new rules
    # have propagated to every node.
    acl_policy = optional(string, "")
  })
}
