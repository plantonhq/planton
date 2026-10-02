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
  description = "GcpRedisInstance specification"
  type = object({
    # GCP project where the Redis instance will be created.
    # If not specified, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Redis instance. This becomes the GCP resource name.
    # Must start with a lowercase letter, contain only lowercase letters, numbers,
    # and hyphens, and end with a lowercase letter or number. Maximum 40 characters.
    # Immutable after creation.
    instance_name = string

    # GCP region where the instance will be deployed (e.g., "us-central1").
    region = string

    # Service tier controlling availability and replication.
    # BASIC: standalone instance, no replication, no SLA.
    # STANDARD_HA: primary + replica with automatic failover, 99.9% SLA.
    # Immutable after creation.
    tier = string

    # Memory size in GiB for the Redis instance. This is the total memory
    # available for storing data. Minimum 1 GiB for BASIC; the GCP API requires
    # at least 5 GiB for STANDARD_HA and for enabling read replicas.
    memory_size_gb = number

    # Redis engine version (e.g., "REDIS_7_0", "REDIS_7_2", "REDIS_6_X").
    # If not specified, the latest supported version is used. Upgrades apply
    # in place; a version downgrade replaces the instance.
    redis_version = optional(string, "")

    # Human-readable display name for the instance.
    display_name = optional(string, "")

    # Zone within the region where the instance will be placed.
    # For STANDARD_HA, this is the primary zone. GCP automatically selects
    # a different zone for the replica unless alternative_location_id pins it.
    # If not specified, GCP picks a zone. Immutable after creation.
    location_id = optional(string, "")

    # Zone for the STANDARD_HA replica. Only applicable to STANDARD_HA tier;
    # must differ from location_id. Pinning both zones matters when co-locating
    # the cache with zonal workloads (e.g. keeping the replica in the same zone
    # as a standby application stack to bound cross-zone latency after
    # failover). If not specified, GCP picks a different zone automatically.
    # Immutable after creation.
    alternative_location_id = optional(string, "")

    # VPC network to which the instance is connected.
    # If not specified, the default network is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    authorized_network = optional(string, "")

    # How the instance connects to the VPC network.
    # DIRECT_PEERING: VPC peering (default). Simpler setup.
    # PRIVATE_SERVICE_ACCESS: uses the network's private services access
    # connection. Required for Shared VPC and lets the instance consume an
    # address range you allocated (compose GcpGlobalAddress +
    # GcpServiceNetworkingConnection on the network first).
    # Immutable after creation.
    connect_mode = optional(string, "")

    # CIDR range of internal addresses reserved for this instance.
    # For DIRECT_PEERING: a /29 block (e.g., "10.0.0.0/29"), unique and
    # non-overlapping with existing subnets; if not specified, GCP selects an
    # unused /29 automatically. For PRIVATE_SERVICE_ACCESS: the NAME of an
    # allocated address range on the private services access connection
    # (a GcpGlobalAddress with purpose VPC_PEERING).
    # Immutable after creation.
    reserved_ip_range = optional(string, "")

    # Additional IP range for node placement. Required when enabling read
    # replicas on an EXISTING instance (the original /29 has no room for the
    # extra nodes). For DIRECT_PEERING: a /28 CIDR or "auto". For
    # PRIVATE_SERVICE_ACCESS: the name of an allocated address range on the
    # private services access connection, or "auto". Mutable — this is the
    # field you set when scaling an in-place instance out to read replicas.
    secondary_ip_range = optional(string, "")

    # Whether Redis AUTH is enabled. When true, clients must provide
    # the AUTH string (exported in stack outputs) to connect.
    # AUTH provides an additional layer of security beyond network controls.
    auth_enabled = optional(bool, false)

    # TLS encryption mode for client-to-server traffic.
    # DISABLED: no encryption (default).
    # SERVER_AUTHENTICATION: clients verify the server's identity via TLS;
    # pair with the server_ca_certs stack output, which carries the CA
    # certificates clients must trust.
    # Immutable after creation.
    transit_encryption_mode = optional(string, "")

    # Redis configuration parameters as key-value pairs.
    # See https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances#Instance.FIELDS.redis_configs
    # for the list of supported parameters (e.g., "maxmemory-policy", "notify-keyspace-events").
    redis_configs = optional(map(string), {})

    # Weekly maintenance window. If not specified, GCP schedules maintenance
    # at its discretion.
    maintenance_window = optional(object({
      # Day of the week for the maintenance window.
      day = string

      # Hour of day (0-23, UTC) when the maintenance window starts.
      hour = optional(number, 0)

      # Minute of the hour (0-59, UTC) when the maintenance window starts.
      # Combined with hour, this pins the window start to the exact minute —
      # useful for coordinating with maintenance windows of dependent systems
      # (e.g. start Redis maintenance 30 minutes after the database's window).
      minute = optional(number, 0)

      # Human-readable description of what this maintenance policy is for
      # (e.g. "post-midnight window, after the nightly batch completes").
      # Maximum 512 characters — the API rejects longer descriptions.
      description = optional(string, "")
    }))

    # Self-service maintenance version. Setting this to a newer available
    # version triggers the maintenance update on your schedule instead of
    # waiting for GCP's rollout — the lever for applying a security patch
    # immediately. Leave unset to follow GCP's automatic rollout.
    maintenance_version = optional(string, "")

    # Read replica mode. Can only be set at creation time.
    # READ_REPLICAS_DISABLED (default): no read endpoint, no scaling.
    # READ_REPLICAS_ENABLED: read endpoint provided, instance can scale replicas.
    # Only available with STANDARD_HA tier.
    read_replicas_mode = optional(string, "")

    # Number of read replicas. Valid range is 1-5 when read_replicas_mode is
    # READ_REPLICAS_ENABLED and tier is STANDARD_HA.
    replica_count = optional(number, 0)

    # Persistence configuration for RDB snapshots.
    persistence_config = optional(object({
      # Persistence mode. DISABLED turns off persistence entirely.
      # RDB enables periodic RDB snapshots.
      persistence_mode = string

      # How often RDB snapshots are taken. Required when persistence_mode is RDB.
      rdb_snapshot_period = optional(string, "")

      # Date and time the first snapshot was/will be attempted, to which all
      # future snapshots align. RFC3339 UTC "Zulu" format (e.g.
      # "2014-10-02T15:01:23Z"). Anchoring the schedule lets you place snapshot
      # I/O in a low-traffic window instead of wherever instance creation time
      # happened to fall. If not provided, GCP uses the creation time.
      rdb_snapshot_start_time = optional(string, "")
    }))

    # Cloud KMS key for customer-managed encryption at rest (CMEK).
    # Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
    # If not specified, data is encrypted with Google-managed keys.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    customer_managed_key = optional(string, "")

    # User-defined labels to organize and track the instance, for cost
    # attribution and fleet queries. Merged beneath Planton's platform
    # attribution labels (platform keys win on conflict).
    labels = optional(map(string), {})

    # Whether deletion protection is enabled. When true (the default —
    # matching GCP's safety posture for stateful stores), destroying the
    # instance fails until this is explicitly set to false. Both IaC
    # engines send the value explicitly so destroy behavior is identical
    # regardless of engine.
    deletion_protection = optional(bool)

    # Deletion policy for the instance — what happens when this resource
    # is destroyed (evaluated only after deletion_protection allows the
    # destroy at all):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the instance is deleted; all in-memory data is lost
    #   "PREVENT" -- destroy FAILS; a second, independent guard for a
    #                cache whose loss would stampede the backing store
    #   "ABANDON" -- the instance is removed from management but left
    #                running (and billing) in GCP with its data intact
    deletion_policy = optional(string, "")
  })
}
