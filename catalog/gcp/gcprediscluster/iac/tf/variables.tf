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
  description = "GcpRedisCluster specification"
  type = object({
    # The GCP project the cluster is created in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The cluster's name in GCP. Defaults to metadata.name. 1-63
    # characters: lowercase letters, digits, and hyphens, starting with a
    # letter and ending alphanumeric. Immutable.
    cluster_name = optional(string, "")

    # The region the cluster lives in (e.g. "us-central1"). Immutable.
    region = string

    # Number of shards. Each shard owns a slice of the keyspace and is at
    # least one node; Google caps shards per cluster at 250. Resizes in
    # place without downtime -- Google rebalances slots across the new
    # shard count.
    shard_count = number

    # Replica nodes per shard (0-5). Replicas serve reads and take over on
    # primary failure; 0 (Google's default) means a shard outage loses that
    # shard's data until it recovers. Resizes in place. Always sent, so the
    # manifest value is authoritative on both engines.
    replica_count = optional(number, 0)

    # Node shape for every node in the cluster -- the per-node hourly rate
    # and the memory each shard holds:
    #   REDIS_SHARED_CORE_NANO -- shared-core, ~1.4 GB (dev and test)
    #   REDIS_STANDARD_SMALL   -- ~6.5 GB, balanced
    #   REDIS_HIGHCPU_MEDIUM   -- ~13 GB, compute-leaning
    #   REDIS_HIGHMEM_MEDIUM   -- ~13 GB, memory-leaning
    #   REDIS_STANDARD_LARGE   -- ~13 GB, balanced
    #   REDIS_HIGHMEM_XLARGE   -- ~58 GB
    #   REDIS_HIGHMEM_2XLARGE  -- ~116 GB
    # If unset, Google picks REDIS_HIGHMEM_MEDIUM. Updates in place.
    node_type = optional(string, "")

    # Native Redis configuration parameters Google lets a cluster tune, as
    # key-value pairs (e.g. "maxmemory-policy": "allkeys-lru",
    # "notify-keyspace-events": "Ex"). Only the parameters in Google's
    # supported subset are accepted. Updates in place.
    redis_configs = optional(map(string), {})

    # Consumer networks for Google-managed Private Service Connect
    # endpoints. Google supports one entry today. Leave empty to publish
    # service attachments only and register hand-built connections through
    # GcpRedisClusterEndpointSet. Updates in place.
    psc_configs = optional(list(object({
      # Consumer VPC network the automatic endpoints land in. The API takes
      # the relative resource path (projects/{project}/global/networks/{name});
      # a GcpVpcNetwork reference resolves to its network_id output, which is
      # already in that form.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      network = string
    })), [])

    # How clients authenticate:
    #   AUTH_MODE_DISABLED -- no authentication (Google's default; the
    #                         network is the boundary)
    #   AUTH_MODE_IAM_AUTH -- clients present an IAM access token as the
    #                         Redis password; access is governed by IAM
    # Immutable. Sent explicitly on both engines so the choice never
    # depends on a provider default.
    authorization_mode = optional(string, "")

    # TLS for client connections:
    #   TRANSIT_ENCRYPTION_MODE_DISABLED              -- plaintext (Google's
    #                                                    default)
    #   TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION -- TLS; clients verify
    #                                                    the server's
    #                                                    certificate
    # Immutable. Sent explicitly on both engines.
    transit_encryption_mode = optional(string, "")

    # Which certificate authority signs the server certificate when TLS is
    # on:
    #   SERVER_CA_MODE_GOOGLE_MANAGED_PER_INSTANCE_CA -- a Google CA unique
    #                                                    to this cluster
    #                                                    (Google's default)
    #   SERVER_CA_MODE_GOOGLE_MANAGED_SHARED_CA       -- a Google CA shared
    #                                                    across clusters, so
    #                                                    a fleet trusts one
    #                                                    CA
    #   SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA        -- your own CA pool in
    #                                                    Certificate
    #                                                    Authority Service
    #                                                    (server_ca_pool)
    # Meaningful only with TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION.
    server_ca_mode = optional(string, "")

    # The Certificate Authority Service pool that signs the server
    # certificate under SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA, as
    # projects/{project}/locations/{region}/caPools/{pool}.
    server_ca_pool = optional(string, "")

    # Customer-managed encryption key (CMEK) for data at rest: a full
    # crypto key ID (projects/*/locations/*/keyRings/*/cryptoKeys/*) or a
    # GcpKmsKey reference. The key must be in the cluster's region and the
    # Memorystore service agent needs encrypter/decrypter on it. If unset,
    # Google-managed keys encrypt the data.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key = optional(string, "")

    # Whether and how data survives a restart. Unset means in-memory only.
    persistence_config = optional(object({
      # Persistence mode:
      #   DISABLED -- in-memory only (Google's default); a restart loses data
      #   RDB      -- periodic point-in-time snapshots (rdb_config)
      #   AOF      -- append-only write log (aof_config)
      mode = optional(string, "")

      # Snapshot schedule. Only meaningful when mode is RDB.
      rdb_config = optional(object({
        # How often RDB snapshots are taken. If unset, Google defaults to
        # TWENTY_FOUR_HOURS.
        rdb_snapshot_period = optional(string, "")

        # RFC 3339 timestamp the snapshot schedule is anchored to (the first
        # snapshot happens at this time, then every rdb_snapshot_period). If
        # unset, Google anchors the schedule at creation.
        rdb_snapshot_start_time = optional(string, "")
      }))

      # Flush policy. Only meaningful when mode is AOF.
      aof_config = optional(object({
        # How often the AOF buffer is flushed to disk (Redis appendfsync):
        #   NO       -- the OS decides (fastest; risks the last seconds of writes)
        #   EVERYSEC -- once per second (Google's default; the usual balance)
        #   ALWAYS   -- on every write (strongest durability, slowest)
        append_fsync = optional(string, "")
      }))
    }))

    # How nodes spread across the region's zones. Unset means MULTI_ZONE.
    # Immutable.
    zone_distribution_config = optional(object({
      # MULTI_ZONE (Google's default) spreads primaries and replicas across
      # zones so a zonal outage keeps the cluster serving; SINGLE_ZONE puts
      # every node in `zone` for the lowest latency at the cost of zonal
      # failure.
      mode = optional(string, "")

      # The zone every node lives in (e.g. "us-central1-a"). Required for
      # SINGLE_ZONE; must be empty for MULTI_ZONE.
      zone = optional(string, "")
    }))

    # The weekly window Google may apply maintenance in. Unset lets Google
    # choose.
    maintenance_policy = optional(object({
      # The weekly window.
      weekly_maintenance_window = object({
        # Day of the week (UTC).
        day = string

        # Hour of day (0-23, UTC) the one-hour window starts.
        hour = optional(number, 0)
      })
    }))

    # Daily backups into the cluster's managed backup collection. Unset
    # means no automated backups (on-demand backups stay available through
    # the console and gcloud).
    automated_backup_config = optional(object({
      # Hour of day (0-23, UTC) the daily backup starts.
      start_hour = optional(number, 0)

      # How long backups are kept, as a seconds duration between one day
      # ("86400s") and 365 days ("31536000s"), e.g. "3024000s" for 35 days.
      retention = string
    }))

    # Cross-region disaster recovery: make this cluster a PRIMARY that
    # replicates to secondaries in other regions, or a SECONDARY that
    # replicates from a primary. Unset (or role NONE) is a standalone
    # cluster.
    cross_cluster_replication_config = optional(object({
      # This cluster's role:
      #   NONE      -- not replicating across regions (Google's default)
      #   PRIMARY   -- serves writes and replicates to the secondaries
      #   SECONDARY -- read-only replica of primary_cluster
      cluster_role = optional(string, "")

      # The primary this cluster replicates from. Required for SECONDARY;
      # must be unset otherwise.
      primary_cluster = optional(object({
        # Full resource path of the primary
        # (projects/{project}/locations/{region}/clusters/{cluster}). A
        # reference resolves to another GcpRedisCluster's name output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        cluster = string
      }))

      # The secondaries replicating from this cluster. Only for PRIMARY.
      secondary_clusters = optional(list(object({
        # Full resource path of the secondary
        # (projects/{project}/locations/{region}/clusters/{cluster}). A
        # reference resolves to another GcpRedisCluster's name output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        cluster = string
      })), [])
    }))

    # Seed the new cluster from RDB files in Cloud Storage at creation.
    # Mutually exclusive with managed_backup_source. Immutable.
    gcs_source = optional(object({
      # Cloud Storage URIs of the RDB files (gs://bucket/path.rdb). The
      # Memorystore service agent needs read access to the objects.
      uris = list(string)
    }))

    # Seed the new cluster from a managed backup at creation. Mutually
    # exclusive with gcs_source. Immutable.
    managed_backup_source = optional(object({
      # Full resource path of the backup
      # (projects/{project}/locations/{region}/backupCollections/{collection}/backups/{backup});
      # backup collections are listed in another cluster's backup_collection
      # output.
      backup = string
    }))

    # User labels on the cluster, merged beneath Planton's platform
    # attribution labels (platform keys win on conflict).
    labels = optional(map(string), {})

    # Whether deletion protection is on. Defaults to true (Google's own
    # posture): destroying the cluster fails until this is set to false.
    # Both engines send the value explicitly so a manifest that never
    # mentions it behaves the same everywhere. Updates in place.
    deletion_protection_enabled = optional(bool)

    # Self-service maintenance: setting this to a newer version from the
    # cluster's available maintenance versions applies the update on your
    # schedule instead of waiting for Google's rollout. Update-only and
    # forward-only; leave unset to follow Google's rollout.
    maintenance_version = optional(string, "")

    # A Memorystore for Redis Cluster ACL policy attached to the cluster:
    # Redis ACL rules (users, key patterns, allowed commands) authored once
    # and shared by clusters in the same region, as
    # projects/{project}/locations/{region}/aclPolicies/{policy}. Leave
    # empty for the built-in default ACL. Updates in place; the cluster's
    # is_acl_policy_in_sync status reports when the rules have reached
    # every node.
    acl_policy = optional(string, "")

    # What happens to the cluster when this resource is destroyed
    # (evaluated after deletion_protection_enabled allows the destroy):
    #   "" / "DELETE" -- the cluster is deleted and its data lost
    #   "PREVENT"     -- destroy fails; a second guard for a cache whose
    #                    loss would stampede the backing store
    #   "ABANDON"     -- the cluster leaves management but keeps running
    #                    (and billing) in GCP
    deletion_policy = optional(string, "")
  })
}
