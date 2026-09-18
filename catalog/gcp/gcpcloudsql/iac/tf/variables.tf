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
  description = "GcpCloudSql specification"
  type = object({
    # The GCP project in which to create this Cloud SQL instance.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Example: "my-prod-project-123"
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Cloud SQL instance in GCP. Immutable.
    # After deletion, the name stays reserved for about one week and cannot be
    # reused — plan instance naming with that soft-delete window in mind.
    # Example: "orders-db-prod"
    instance_name = string

    # The GCP region hosting the instance (e.g. "us-central1"). Immutable:
    # an instance cannot move between regions in place.
    region = string

    # The database engine family. Drives engine-specific validation: the
    # database_version prefix, PITR vs binary-log semantics, and the SQL
    # Server-only surface (time_zone, collation, audit, Active Directory).
    database_engine = string

    # The exact engine version, e.g. "MYSQL_8_0", "POSTGRES_16",
    # "SQLSERVER_2022_STANDARD". In-place major version upgrades are supported
    # by the API (no destroy/recreate), so this field is mutable — but always
    # take a backup before upgrading.
    database_version = string

    # Machine type for the instance, e.g. "db-f1-micro" (shared-core,
    # dev/test), "db-custom-4-15360" (4 vCPU / 15 GB), or an Enterprise Plus
    # performance tier like "db-perf-optimized-N-4". Mutable: changing tier
    # resizes the instance in place (with a restart).
    tier = string

    # Cloud SQL edition. ENTERPRISE is the standard tier (99.95% SLA on HA).
    # ENTERPRISE_PLUS adds a 99.99% HA SLA, the data cache, near-zero-downtime
    # maintenance, and 35-day transaction log retention. Mutable: edition
    # upgrades happen in place.
    edition = optional(string)

    # ZONAL runs a single instance in one zone. REGIONAL enables high
    # availability: a standby in a second zone with automatic failover.
    # REGIONAL requires automated backups (and binary logs on MySQL).
    # Mutable: HA can be enabled or disabled in place.
    availability_type = optional(string)

    # When the instance is activated. ALWAYS keeps it running (default);
    # NEVER stops the instance (storage is retained and billed — the
    # stop/start lever without destroying data); ON_DEMAND is legacy
    # first-generation behavior.
    activation_policy = optional(string)

    # Data disk configuration. If omitted: 10 GB PD_SSD with auto-resize.
    disk = optional(object({
      # Disk type: PD_SSD (default, general purpose), PD_HDD (cheaper, slower —
      # dev/archive only), or HYPERDISK_BALANCED (min 20 GB). Immutable.
      type = optional(string)

      # Disk size in GB (10–65536). Can grow in place; can NEVER shrink —
      # shrinking requires replacing the instance. With auto_resize enabled,
      # GCP grows the disk past this value as data accumulates.
      size_gb = optional(number)

      # Automatically grow the disk as it approaches capacity. Enabled by
      # default — running a database out of disk is an outage.
      auto_resize = optional(bool)

      # Upper bound in GB for automatic growth (0 = no limit). The brake that
      # stops a runaway workload from growing a disk — and a bill — without
      # bound.
      auto_resize_limit = optional(number)

      # HYPERDISK_BALANCED only: provisioned I/O operations per second for the
      # data disk — the IOPS dial decoupled from disk size.
      provisioned_iops = optional(number)

      # HYPERDISK_BALANCED only: provisioned throughput in MiB/s for the data
      # disk — the bandwidth dial decoupled from disk size.
      provisioned_throughput = optional(number)
    }))

    # Connectivity configuration: public IPv4, private VPC IP, and/or Private
    # Service Connect. If omitted, the instance gets a public IPv4 address
    # with NO authorized networks — reachable only through the Cloud SQL Auth
    # Proxy or connectors (IAM-authenticated), which is a safe default.
    network = optional(object({
      # The VPC network for private IP connectivity, in
      # projects/{project}/global/networks/{network} form. Accepts a literal or
      # a reference to a GcpVpcNetwork resource. Setting this ENABLES private IP.
      # The network must already have a service networking connection (compose
      # GcpGlobalAddress with purpose VPC_PEERING + GcpServiceNetworkingConnection)
      # or instance creation fails. Can be set or changed in place, but never
      # removed — removing it forces instance replacement.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      private_network = optional(string, "")

      # Whether the instance gets a public IPv4 address. With no
      # authorized_networks, a public IP is reachable only through the Cloud
      # SQL Auth Proxy or connectors (IAM-authenticated) — a safe pattern.
      ipv4_enabled = optional(bool, false)

      # CIDR ranges allowed to connect DIRECTLY to the public IP. Prefer the
      # Auth Proxy over widening this list; never add 0.0.0.0/0 to a production
      # instance.
      authorized_networks = optional(list(object({
        # The CIDR range, e.g. "203.0.113.0/24".
        value = string

        # Display label for this entry in the console.
        name = optional(string, "")

        # RFC 3339 timestamp after which this entry stops being honored —
        # built-in expiry for temporary access grants.
        expiration_time = optional(string, "")
      })), [])

      # Name of a specific allocated IP range (a GcpGlobalAddress with purpose
      # VPC_PEERING) from which the private IP is assigned. If empty, GCP picks
      # any range on the service networking connection.
      allocated_ip_range = optional(string, "")

      # Allows Google Cloud services (e.g. BigQuery federated queries) to reach
      # this instance over its private IP path.
      enable_private_path_for_google_cloud_services = optional(bool, false)

      # TLS posture for DIRECT connections. ALLOW_UNENCRYPTED_AND_ENCRYPTED
      # (default), ENCRYPTED_ONLY (reject plaintext), or
      # TRUSTED_CLIENT_CERTIFICATE_REQUIRED (mutual TLS with client certs).
      # Connector/Auth Proxy traffic is always encrypted regardless.
      ssl_mode = optional(string, "")

      # Which certificate authority signs the server certificate.
      # GOOGLE_MANAGED_INTERNAL_CA (default, per-instance CA),
      # GOOGLE_MANAGED_CAS_CA (Google-managed CA hierarchy in CA Service), or
      # CUSTOMER_MANAGED_CAS_CA (your own CA pool — set server_ca_pool).
      # Immutable after creation.
      server_ca_mode = optional(string, "")

      # The CA Service CA pool (full resource path) that signs the server
      # certificate when server_ca_mode is CUSTOMER_MANAGED_CAS_CA.
      server_ca_pool = optional(string, "")

      # Additional DNS names embedded in the server certificate (customer-
      # managed CA only) — lets clients validate the cert against your own
      # hostnames instead of the instance IP.
      custom_subject_alternative_names = optional(list(string), [])

      # Private Service Connect: expose the instance as a PSC service
      # attachment that consumer VPCs connect to via PSC endpoints — private
      # connectivity WITHOUT VPC peering. The PSC alternative to
      # private_network.
      psc = optional(object({
        # Whether PSC connectivity is enabled for this instance. Immutable.
        enabled = optional(bool, false)

        # Consumer projects allowed to create PSC endpoints to this instance.
        allowed_consumer_projects = optional(list(string), [])

        # Network attachment (full resource path) for outbound connectivity from
        # the instance into a consumer VPC (used by outbound features such as
        # external replication over PSC).
        network_attachment_uri = optional(string, "")

        # PSC endpoints GCP creates automatically in the listed consumer
        # networks — endpoint provisioning without consumer-side IaC.
        auto_connections = optional(list(object({
          # The consumer VPC network (full resource path) in which GCP creates the
          # PSC endpoint.
          consumer_network = string

          # The project owning the consumer network (defaults to the network's
          # project).
          consumer_service_project_id = optional(string, "")
        })), [])

        # Automatically create DNS records for the PSC endpoints in the
        # consumer networks — clients resolve the instance by name instead of
        # tracking endpoint IPs.
        auto_dns_enabled = optional(bool, false)

        # Enterprise Plus only: also create a DNS record for the PSA write
        # endpoint, so clients follow the primary across switchovers by name.
        write_endpoint_dns_enabled = optional(bool, false)

        # Whether Cloud SQL also creates a Service Connection Policy for the
        # auto_connections above, so the consumer networks need no separately
        # authored policy before the automatic endpoints can be provisioned.
        # Leave unset to keep the API's own default; sent only when set because
        # the API fills the value itself.
        auto_connection_policy_enabled = optional(bool)
      }))

      # Controls automatic rotation of the server certificate.
      # NO_AUTOMATIC_ROTATION (default) or
      # AUTOMATIC_ROTATION_DURING_MAINTENANCE (requires a CA Service
      # server_ca_mode: GOOGLE_MANAGED_CAS_CA or CUSTOMER_MANAGED_CAS_CA).
      server_certificate_rotation_mode = optional(string, "")
    }))

    # Preferred zone placement for the primary (and the standby on REGIONAL
    # instances). If omitted, GCP picks zones automatically.
    location_preference = optional(object({
      # Preferred zone for the primary, e.g. "us-central1-a". Must be in the
      # instance's region.
      zone = optional(string, "")

      # Preferred zone for the standby of a REGIONAL instance. Must differ from
      # zone.
      secondary_zone = optional(string, "")
    }))

    # Automated backup configuration. Strongly recommended for anything that
    # holds real data; required for REGIONAL availability and for creating
    # read replicas.
    backup = optional(object({
      # Whether daily automated backups run. The foundation for PITR, read
      # replicas, and REGIONAL high availability.
      enabled = optional(bool, false)

      # Start of the daily backup window in "HH:MM" (UTC). If empty, GCP
      # assigns a window.
      start_time = optional(string, "")

      # Multi-region or region where backups are stored, e.g. "us" or
      # "us-central1". If empty, GCP picks the closest multi-region.
      location = optional(string, "")

      # MySQL only: write-ahead binary logging — MySQL's mechanism for
      # point-in-time recovery and the prerequisite for MySQL replicas and HA.
      binary_log_enabled = optional(bool, false)

      # PostgreSQL / SQL Server only: point-in-time recovery via write-ahead /
      # transaction logs. Lets you restore to any second inside the log
      # retention window — the defense against bad migrations and accidental
      # deletes.
      point_in_time_recovery_enabled = optional(bool, false)

      # Days of transaction logs retained for PITR: 1–7 (ENTERPRISE) or 1–35
      # (ENTERPRISE_PLUS). Default 7.
      transaction_log_retention_days = optional(number)

      # Number of daily backups retained (default 7). Older backups are pruned
      # automatically.
      retained_backups = optional(number)

      # The unit retained_backups counts in. COUNT (the default and the only
      # unit the API defines) retains that many most-recent backups.
      retention_unit = optional(string, "")
    }))

    # One-hour weekly window in which GCP may restart the instance to apply
    # updates. Without it, maintenance can happen at any time.
    maintenance_window = optional(object({
      # Day of week: 1 (Monday) through 7 (Sunday).
      day = number

      # Hour of day 0–23 (UTC) at which the window opens. 0 is midnight.
      hour = optional(number)

      # Release cadence: "canary" (about one week after notification), "stable"
      # (about two weeks), or "week5" (about five weeks — maximum notice).
      update_track = optional(string, "")
    }))

    # A date range during which maintenance is denied (e.g. an end-of-year
    # freeze). At most 90 days per period.
    deny_maintenance_period = optional(object({
      # First denied date, "yyyy-mm-dd" (specific year) or "mm-dd" (recurs
      # annually).
      start_date = string

      # Last denied date, same format as start_date.
      end_date = string

      # Time of day (UTC) at which the deny period starts and ends, "HH:mm:SS",
      # e.g. "00:00:00".
      time = string
    }))

    # Query Insights: per-query performance telemetry in the console.
    # Negligible overhead for most workloads — enable it in production.
    insights_config = optional(object({
      # Whether Query Insights collects per-query performance telemetry.
      query_insights_enabled = optional(bool, false)

      # Maximum captured query text length in bytes (default 1024). Raise it
      # when long analytical queries get truncated in the console.
      query_string_length = optional(number)

      # Record application tags (e.g. from sqlcommenter) with each query.
      record_application_tags = optional(bool, false)

      # Record the client IP address with each query.
      record_client_address = optional(bool, false)

      # Sampled execution plans captured per minute across all queries
      # (0 disables plan sampling; default 5).
      query_plans_per_minute = optional(number)

      # Enhanced Query Insights: longer retention and finer-grained query
      # telemetry than the standard tier.
      enhanced_query_insights_enabled = optional(bool, false)
    }))

    # Password complexity/rotation policy enforced by the instance for
    # built-in database users.
    password_validation_policy = optional(object({
      # Master switch for the policy. The other fields take effect only while
      # this is true.
      enable_password_policy = optional(bool, false)

      # Minimum password length.
      min_length = optional(number)

      # COMPLEXITY_DEFAULT requires a mix of lower/upper case, numbers, and
      # non-alphanumeric characters.
      complexity = optional(string, "")

      # Number of previous passwords that cannot be reused.
      reuse_interval = optional(number)

      # Disallow the username as a substring of the password.
      disallow_username_substring = optional(bool, false)

      # PostgreSQL only: minimum interval between password changes, as a
      # duration string, e.g. "3600s".
      password_change_interval = optional(string, "")
    }))

    # Enables the data cache (local SSD read caching). Enterprise Plus only;
    # delivers up to 4x read throughput for cache-friendly workloads.
    data_cache_enabled = optional(bool, false)

    # Managed connection pooling (built-in pooler in front of the engine).
    # Reduces connection-storm pressure without deploying PgBouncer/ProxySQL.
    connection_pooling = optional(object({
      # Whether managed connection pooling is enabled.
      enabled = optional(bool, false)

      # Pooler tuning flags (name → value), e.g. {"max_client_connections":
      # "1000"}. Permitted flags are engine-specific and validated by the API.
      flags = optional(map(string), {})
    }))

    # Engine configuration flags, e.g. {"max_connections": "500"} or
    # {"cloudsql.iam_authentication": "on"} (required on PostgreSQL before
    # creating IAM-type users). Flag names and permitted values are
    # engine-specific and validated by the API at deploy time.
    database_flags = optional(map(string), {})

    # SQL Server only: number of threads per physical core (1 or 2).
    # Tuning lever for SQL Server licensing/performance trade-offs.
    threads_per_core = optional(number)

    # SQL Server only: server time zone, e.g. "Pacific Standard Time".
    # Immutable in practice — changing it forces a maintenance restart.
    time_zone = optional(string, "")

    # SQL Server only: server-level collation, e.g.
    # "SQL_Latin1_General_CP1_CI_AS". Immutable (set at create time).
    # MySQL/PostgreSQL collation is configured per database on
    # GcpCloudSqlDatabase instead.
    collation = optional(string, "")

    # SQL Server only: SQLServer Audit — writes audit files to a GCS bucket.
    sql_server_audit_config = optional(object({
      # Destination bucket, e.g. "gs://my-audit-bucket". The instance's service
      # account (service_account_email output) needs write access to it.
      bucket = optional(string, "")

      # How long generated audit files are kept, e.g. "86400s" (1 day).
      retention_interval = optional(string, "")

      # How often audit files are uploaded to the bucket, e.g. "1800s".
      upload_interval = optional(string, "")
    }))

    # SQL Server only: the Active Directory the instance joins for Windows
    # authentication — a Managed Microsoft AD domain or (with mode
    # CUSTOMER_MANAGED_ACTIVE_DIRECTORY) a self-managed AD reached through
    # the listed domain controllers.
    active_directory = optional(object({
      # The AD domain to join, e.g. "ad.example.com". With the default
      # (managed) mode this is a Managed Microsoft AD domain in the project.
      domain = string

      # MANAGED_ACTIVE_DIRECTORY (default) joins a Managed Microsoft AD
      # domain; CUSTOMER_MANAGED_ACTIVE_DIRECTORY joins a self-managed AD
      # bootstrapped through dns_servers and the admin credential.
      mode = optional(string, "")

      # Customer-managed AD only: domain controller IPv4 addresses used to
      # bootstrap the join.
      dns_servers = optional(list(string), [])

      # Customer-managed AD only: the Secret Manager secret
      # (projects/{project}/secrets/{secret}) holding the AD administrator
      # credential used to join the domain. A secret NAME, not the credential
      # itself.
      admin_credential_secret_name = optional(string, "")

      # Customer-managed AD only: the organizational unit distinguished name
      # (full hierarchical path) the instance's computer account joins under.
      organizational_unit = optional(string, "")
    }))

    # Connection-path enforcement. REQUIRED rejects all direct connections,
    # admitting only Cloud SQL connectors / Auth Proxy traffic (which is
    # always TLS-encrypted and IAM-authenticated). NOT_REQUIRED (default)
    # admits direct connections too.
    connector_enforcement = optional(string, "")

    # Enables Vertex AI integration (e.g. ML predictions from SQL via
    # ml_integration). PostgreSQL and MySQL.
    enable_google_ml_integration = optional(bool, false)

    # Enables Dataplex integration for data cataloging/lineage.
    enable_dataplex_integration = optional(bool, false)

    # Customer-managed encryption key (CMEK) for the instance's storage.
    # Accepts a full crypto key path
    # (projects/.../locations/.../keyRings/.../cryptoKeys/...) or a reference
    # to a GcpKmsKey resource. The key MUST be in the same region as the
    # instance. Immutable: CMEK cannot be added or changed after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    encryption_key_name = optional(string, "")

    # Engine-side delete guard: when true, both IaC engines refuse to destroy
    # the instance (the plan/preview fails) until this is set back to false.
    # Protects against a bad manifest or an accidental destroy.
    deletion_protection = optional(bool, false)

    # API-side delete guard: when true, GCP itself rejects instance deletion
    # from EVERY surface — console, gcloud, API, and IaC. The strongest
    # protection; set both guards on production instances.
    deletion_protection_enabled = optional(bool, false)

    # When true, automated backups (and transaction logs for PITR) are
    # retained after the instance is deleted — the recovery path for
    # "deleted the instance, need the data back".
    retain_backups_on_delete = optional(bool, false)

    # Makes this instance a READ REPLICA of the named primary instance.
    # Accepts the primary's instance name or a reference to a GcpCloudSql
    # resource. Immutable — an existing primary cannot be converted in place.
    # The primary must have automated backups enabled (and binary logs on
    # MySQL). To promote a replica into a standalone primary, set
    # instance_type to CLOUD_SQL_INSTANCE and clear this field and
    # replica_configuration in the same change (the instance restarts).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    master_instance_name = optional(string, "")

    # Replica behavior (failover target, cascading) and — for replicas of an
    # external source — the replication channel credentials. Only meaningful
    # together with master_instance_name.
    replica_configuration = optional(object({
      # Designates this replica as the failover target promoted if the primary
      # fails. MySQL only (legacy HA); not supported for PostgreSQL — prefer
      # availability_type REGIONAL on the primary for modern HA.
      failover_target = optional(bool, false)

      # SQL Server only: allows this cross-region replica to carry replicas of
      # its own (cascading replication).
      cascadable_replica = optional(bool, false)

      # Replication username on the external source.
      username = optional(string, "")

      # Replication password on the external source. Write-only replication
      # channel material — never exported in outputs.
      password = optional(string, "")

      # PEM certificate of the external source's CA — public trust material,
      # not a secret.
      ca_certificate = optional(string, "")

      # PEM client certificate for mutual TLS with the external source —
      # public handshake material, not a secret.
      client_certificate = optional(string, "")

      # PEM private key matching client_certificate. Secret key material.
      client_key = optional(string, "")

      # Path to a SQL dump file in GCS (gs://...) used to seed the replica.
      dump_file_path = optional(string, "")

      # Seconds between connection retries to the source (MySQL only).
      connect_retry_interval = optional(number)

      # Interval in milliseconds between replication heartbeats (MySQL only).
      master_heartbeat_period = optional(number)

      # Permitted ciphers for the replication channel TLS (MySQL only).
      ssl_cipher = optional(string, "")

      # Whether to verify the external source's server certificate against
      # ca_certificate (MySQL only).
      verify_server_certificate = optional(bool, false)
    }))

    # Initial password for the engine's default admin user ("root" on MySQL,
    # "postgres" on PostgreSQL, "sqlserver" on SQL Server — where it is
    # REQUIRED). Write-only in GCP: never readable back from the API and
    # never exported in outputs. Create additional users as first-class
    # GcpCloudSqlUser resources.
    root_password = optional(string, "")

    # Engine-side teardown behavior. "DELETE" (default) destroys the
    # instance; "PREVENT" fails any plan that would destroy it; "ABANDON"
    # removes it from IaC management while leaving it running in GCP.
    # Distinct from deletion_protection (which blocks the destroy) — ABANDON
    # is the lever for handing an instance over to out-of-band management.
    deletion_policy = optional(string, "")

    # The instance's role. Usually derived by GCP (primary vs replica);
    # set explicitly for two workflows: READ_POOL_INSTANCE turns the
    # resource into a read pool (with node_count / read_pool_auto_scale),
    # and CLOUD_SQL_INSTANCE promotes an existing read replica into a
    # standalone primary (clear master_instance_name and
    # replica_configuration in the same change).
    instance_type = optional(string, "")

    # Read pools only: number of nodes serving reads behind the pool's
    # single endpoint. Read-only while read_pool_auto_scale is enabled (the
    # autoscaler owns it).
    node_count = optional(number)

    # Read pools only: automatic node-count scaling between min and max
    # bounds, driven by target metrics.
    read_pool_auto_scale = optional(object({
      # Whether auto scaling is active. While enabled, node_count is owned by
      # the autoscaler and read-only.
      enabled = optional(bool, false)

      # Lower bound of pool nodes; scale-in never goes below it.
      min_node_count = optional(number)

      # Upper bound of pool nodes; scale-out never exceeds it.
      max_node_count = optional(number)

      # Disables scale-in entirely: the pool only ever grows automatically.
      disable_scale_in = optional(bool, false)

      # Cooldown in seconds after a scale-in before another may run.
      scale_in_cooldown_seconds = optional(number)

      # Cooldown in seconds after a scale-out before another may run.
      scale_out_cooldown_seconds = optional(number)

      # Metrics the autoscaler steers by; nodes are added or removed to hold
      # each metric at its target value.
      target_metrics = optional(list(object({
        # Metric name, e.g. a CPU utilization metric.
        metric = string

        # Target value for the metric.
        target_value = optional(number, 0)
      })), [])
    }))

    # Creates this instance as a CLONE of another instance — a full copy of
    # its data at a point in time. A create-time source: changing it on an
    # existing instance triggers a new clone operation.
    clone = optional(object({
      # Name of the source instance to clone.
      source_instance_name = string

      # Project of the source instance, for cross-project clones. Defaults to
      # this instance's project.
      source_project = optional(string, "")

      # RFC 3339 timestamp to clone from (point-in-time clone). Requires PITR
      # (or binary logs on MySQL) on the source.
      point_in_time = optional(string, "")

      # PostgreSQL point-in-time clones only: the zone the clone lands in.
      # Defaults to the source's zone.
      preferred_zone = optional(string, "")

      # SQL Server point-in-time clones only: clone just the named databases.
      # Empty clones all databases.
      database_names = optional(list(string), [])

      # Private-IP clones: the allocated IP range name (RFC 1035) the clone's
      # private IP is drawn from.
      allocated_ip_range = optional(string, "")

      # Cloning a DELETED instance: the RFC 3339 timestamp of the source's
      # deletion.
      source_instance_deletion_time = optional(string, "")
    }))

    # Restores a specific backup run into this instance. An imperative
    # restore trigger expressed declaratively: adding or changing the block
    # runs the restore after the instance exists.
    restore_backup_context = optional(object({
      # The backup run ID to restore.
      backup_run_id = optional(number, 0)

      # The instance the backup was taken from. Defaults to this instance.
      instance_id = optional(string, "")

      # The full project ID of the source instance.
      project = optional(string, "")
    }))

    # Restores this instance from a Backup and DR datasource to a point in
    # time. Like restore_backup_context, adding or changing the block
    # triggers the restore.
    point_in_time_restore_context = optional(object({
      # The Backup and DR datasource URI to restore from.
      datasource = string

      # RFC 3339 timestamp to restore to.
      point_in_time = string

      # Name of the target instance receiving the restore.
      target_instance = optional(string, "")

      # Region of the target instance, e.g. "us-central1".
      region = optional(string, "")

      # The zone the restored instance lands in. Defaults to the source
      # primary's zone.
      preferred_zone = optional(string, "")

      # Private-IP restores: the allocated IP range name (RFC 1035) the
      # restored instance's private IP is drawn from.
      allocated_ip_range = optional(string, "")
    }))

    # Restores from a Backup and DR backup (the backup's full resource
    # name). The backup must be in active state. Adding or changing this
    # triggers the restore.
    backupdr_backup = optional(string, "")

    # Pins the instance's maintenance (patch) version. Cannot be set at
    # creation; updating it restarts the instance. Values older than the
    # running version are ignored by the API.
    maintenance_version = optional(string, "")

    # Declares the instance's read replicas by name from the primary's side.
    # Most compositions leave this to GCP (replicas declare their primary
    # via master_instance_name instead).
    replica_names = optional(list(string), [])

    # MySQL/PostgreSQL disaster recovery: names this primary's DR replica
    # ("project:instance" or plain instance name), enabling switchover /
    # replica failover between regions.
    failover_dr_replica_name = optional(string, "")

    # MySQL 8.0 only: opts the instance into automatic minor-version
    # upgrades. The database_version must be an eligible MYSQL_8_0 minor.
    auto_upgrade_enabled = optional(bool, false)

    # Whether the ExecuteSql API may connect to this instance.
    # DISALLOW_DATA_API (default) rejects it; ALLOW_DATA_API admits it —
    # on private-IP instances this allows authorized users to reach the
    # instance from the public internet through the API.
    data_api_access = optional(string, "")

    # A final backup taken automatically when the instance is deleted — the
    # safety net that survives the teardown itself.
    final_backup = optional(object({
      # Whether a final backup is taken on delete.
      enabled = optional(bool, false)

      # Days the final backup is retained.
      retention_days = optional(number)

      # Description recorded on the final backup.
      description = optional(string, "")
    }))

    # SQL Server only: Microsoft Entra ID (Azure AD) authentication for the
    # instance.
    entra_id = optional(object({
      # The Entra ID application (client) ID.
      application_id = string

      # The Entra ID tenant (directory) ID.
      tenant_id = string
    }))

    # Opt-in that lets the instance move point-in-time-recovery transaction
    # logs from the data disk to Cloud Storage, freeing disk space and
    # allowing longer transaction-log retention windows. An input-only
    # instruction: Cloud SQL acts on it but never stores it, so it is sent
    # exactly as written and never read back.
    switch_transaction_logs_to_cloud_storage_enabled = optional(bool, false)

    # Opt-in that upgrades this primary's read replicas in place, together
    # with the primary, when database_version moves to a new major version.
    # Without it a major-version upgrade leaves replicas on the old version
    # to be upgraded (or recreated) separately. Input-only: consulted only
    # during a major-version upgrade and never stored by the API.
    include_replicas_for_major_version_upgrade = optional(bool, false)

    # Irreversible opt-in to Cloud SQL's new network architecture for an
    # instance created in a project that predates it (projects created after
    # August 2021 already use it). Required before features such as PSC
    # and outbound network attachments on those older projects. Once true
    # it cannot be set back to false. Leave unset to let Cloud SQL report
    # the instance's current architecture; sent only when set because the
    # API fills the value itself.
    enforce_new_sql_network_architecture = optional(bool)

    # Read replicas only: the replication lag, in seconds, beyond which the
    # replica recreates itself. The lag must persist for at least five
    # minutes before recreation triggers. Between 300 (five minutes) and
    # 31536000 (one year). Leave unset for no automatic recreation; sent
    # only when set because the API fills the value itself.
    replication_lag_max_seconds = optional(number)
  })
}
