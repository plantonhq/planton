variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Cloud SQL instance"
  type = object({
    # The GCP project that owns the instance. The CLI's tfvars converter
    # resolves StringValueOrRef fields to their literal string before the
    # module runs, so this arrives as a plain string.
    # If empty, the provider's default project is used (see locals.tf).
    project_id = optional(string, "")

    # Name of the instance in GCP. Immutable; reserved ~1 week post-delete.
    instance_name = string

    # Region hosting the instance (e.g. us-central1). Immutable.
    region = string

    # MYSQL, POSTGRESQL, or SQLSERVER. Drives spec-level validation only —
    # the API derives the engine from database_version.
    database_engine = string

    # Exact engine version, e.g. POSTGRES_16. Mutable (in-place upgrade).
    database_version = string

    # Machine type, e.g. db-custom-4-15360. Mutable (in-place resize).
    tier = string

    # ENTERPRISE (middleware default) or ENTERPRISE_PLUS. Mutable.
    edition = optional(string, "ENTERPRISE")

    # ZONAL (middleware default) or REGIONAL (HA with automatic failover).
    availability_type = optional(string, "ZONAL")

    # ALWAYS (middleware default), NEVER (stopped, storage retained), or
    # ON_DEMAND (legacy).
    activation_policy = optional(string, "ALWAYS")

    # Data disk. Null means 10 GB PD_SSD with auto-resize.
    disk = optional(object({
      type              = optional(string, "PD_SSD")
      size_gb           = optional(number, 10)
      auto_resize       = optional(bool, true)
      auto_resize_limit = optional(number, 0)
      # HYPERDISK_BALANCED only: provisioned performance dials.
      provisioned_iops       = optional(number, null)
      provisioned_throughput = optional(number, null)
    }), null)

    # Connectivity. Null means public IPv4 with no authorized networks
    # (Auth Proxy / connector access only).
    network = optional(object({
      # VPC for private IP (projects/.../global/networks/... form); arrives
      # as a plain string after ref resolution. Setting it enables private
      # IP. Never removable in place (ForceNew on unset).
      private_network = optional(string, "")

      ipv4_enabled = optional(bool, false)

      authorized_networks = optional(list(object({
        value           = string
        name            = optional(string, "")
        expiration_time = optional(string, "")
      })), [])

      # Named allocated range (a VPC_PEERING global address) to draw the
      # private IP from; empty lets GCP pick.
      allocated_ip_range = optional(string, "")

      enable_private_path_for_google_cloud_services = optional(bool, false)

      # TLS posture for direct connections.
      ssl_mode = optional(string, "")

      # Server certificate CA hierarchy; immutable.
      server_ca_mode = optional(string, "")
      server_ca_pool = optional(string, "")

      custom_subject_alternative_names = optional(list(string), [])

      # Private Service Connect exposure.
      psc = optional(object({
        enabled                   = optional(bool, false)
        allowed_consumer_projects = optional(list(string), [])
        network_attachment_uri    = optional(string, "")
        auto_connections = optional(list(object({
          consumer_network            = string
          consumer_service_project_id = optional(string, "")
        })), [])
        auto_dns_enabled           = optional(bool, false)
        write_endpoint_dns_enabled = optional(bool, false)
        # Also create a Service Connection Policy for the auto connections.
        # Sent only when set (API-computed).
        auto_connection_policy_enabled = optional(bool, null)
      }), null)

      # Automatic server certificate rotation (CAS CA modes only).
      server_certificate_rotation_mode = optional(string, "")
    }), null)

    # Preferred primary/standby zones.
    location_preference = optional(object({
      zone           = optional(string, "")
      secondary_zone = optional(string, "")
    }), null)

    # Automated backups + point-in-time recovery.
    backup = optional(object({
      enabled                        = optional(bool, false)
      start_time                     = optional(string, "")
      location                       = optional(string, "")
      binary_log_enabled             = optional(bool, false)
      point_in_time_recovery_enabled = optional(bool, false)
      transaction_log_retention_days = optional(number, null)
      retained_backups               = optional(number, null)
      retention_unit                 = optional(string, "")
    }), null)

    # Weekly one-hour maintenance window.
    maintenance_window = optional(object({
      day          = number
      hour         = optional(number, null)
      update_track = optional(string, "")
    }), null)

    # Maintenance freeze window (max 90 days).
    deny_maintenance_period = optional(object({
      start_date = string
      end_date   = string
      time       = string
    }), null)

    # Query Insights telemetry.
    insights_config = optional(object({
      query_insights_enabled          = optional(bool, false)
      query_string_length             = optional(number, null)
      record_application_tags         = optional(bool, false)
      record_client_address           = optional(bool, false)
      query_plans_per_minute          = optional(number, null)
      enhanced_query_insights_enabled = optional(bool, false)
    }), null)

    # Password rules for built-in users.
    password_validation_policy = optional(object({
      enable_password_policy      = optional(bool, false)
      min_length                  = optional(number, null)
      complexity                  = optional(string, "")
      reuse_interval              = optional(number, null)
      disallow_username_substring = optional(bool, false)
      password_change_interval    = optional(string, "")
    }), null)

    # Enterprise Plus local-SSD read cache.
    data_cache_enabled = optional(bool, false)

    # Managed connection pooling.
    connection_pooling = optional(object({
      enabled = optional(bool, false)
      flags   = optional(map(string), {})
    }), null)

    # Engine flags, e.g. {"max_connections" = "500"}.
    database_flags = optional(map(string), {})

    # SQL Server only.
    threads_per_core = optional(number, null)
    time_zone        = optional(string, "")
    collation        = optional(string, "")
    sql_server_audit_config = optional(object({
      bucket             = optional(string, "")
      retention_interval = optional(string, "")
      upload_interval    = optional(string, "")
    }), null)

    # SQL Server only: Active Directory join (managed or customer-managed).
    active_directory = optional(object({
      domain                       = string
      mode                         = optional(string, "")
      dns_servers                  = optional(list(string), [])
      admin_credential_secret_name = optional(string, "")
      organizational_unit          = optional(string, "")
    }), null)

    # SQL Server only: Microsoft Entra ID authentication.
    entra_id = optional(object({
      application_id = string
      tenant_id      = string
    }), null)

    # NOT_REQUIRED or REQUIRED (connectors-only access).
    connector_enforcement = optional(string, "")

    enable_google_ml_integration = optional(bool, false)
    enable_dataplex_integration  = optional(bool, false)

    # CMEK crypto key path; arrives as a plain string after ref resolution.
    # Immutable.
    encryption_key_name = optional(string, "")

    # Engine-side destroy guard (plan-time refusal).
    deletion_protection = optional(bool, false)

    # API-side delete guard (blocks deletion from every surface).
    deletion_protection_enabled = optional(bool, false)

    # Keep backups after instance deletion.
    retain_backups_on_delete = optional(bool, false)

    # Primary instance name when this instance is a read replica; arrives
    # as a plain string after ref resolution. Immutable.
    master_instance_name = optional(string, "")

    # Replica behavior + external-source replication channel.
    replica_configuration = optional(object({
      failover_target           = optional(bool, false)
      cascadable_replica        = optional(bool, false)
      username                  = optional(string, "")
      password                  = optional(string, "")
      ca_certificate            = optional(string, "")
      client_certificate        = optional(string, "")
      client_key                = optional(string, "")
      dump_file_path            = optional(string, "")
      connect_retry_interval    = optional(number, null)
      master_heartbeat_period   = optional(number, null)
      ssl_cipher                = optional(string, "")
      verify_server_certificate = optional(bool, false)
    }), null)

    # Initial admin password (root/postgres/sqlserver user). Write-only in
    # GCP; required for SQL Server.
    root_password = optional(string, "")

    # Engine-side teardown behavior: DELETE, PREVENT, or ABANDON.
    deletion_policy = optional(string, "")

    # Instance role; set READ_POOL_INSTANCE for read pools, or
    # CLOUD_SQL_INSTANCE to promote a replica to standalone.
    instance_type = optional(string, "")

    # Read pools: node count behind the pool endpoint.
    node_count = optional(number, null)

    # Read pools: automatic node-count scaling.
    read_pool_auto_scale = optional(object({
      enabled                    = optional(bool, false)
      min_node_count             = optional(number, null)
      max_node_count             = optional(number, null)
      disable_scale_in           = optional(bool, false)
      scale_in_cooldown_seconds  = optional(number, null)
      scale_out_cooldown_seconds = optional(number, null)
      target_metrics = optional(list(object({
        metric       = string
        target_value = optional(number, null)
      })), [])
    }), null)

    # Create-time clone source.
    clone = optional(object({
      source_instance_name          = string
      source_project                = optional(string, "")
      point_in_time                 = optional(string, "")
      preferred_zone                = optional(string, "")
      database_names                = optional(list(string), [])
      allocated_ip_range            = optional(string, "")
      source_instance_deletion_time = optional(string, "")
    }), null)

    # Backup-run restore trigger.
    restore_backup_context = optional(object({
      backup_run_id = number
      instance_id   = optional(string, "")
      project       = optional(string, "")
    }), null)

    # Backup and DR point-in-time restore trigger.
    point_in_time_restore_context = optional(object({
      datasource         = string
      point_in_time      = string
      target_instance    = optional(string, "")
      region             = optional(string, "")
      preferred_zone     = optional(string, "")
      allocated_ip_range = optional(string, "")
    }), null)

    # Backup and DR backup name restore trigger.
    backupdr_backup = optional(string, "")

    # Pinned maintenance (patch) version; updating restarts the instance.
    maintenance_version = optional(string, "")

    # Replica names declared from the primary's side.
    replica_names = optional(list(string), [])

    # DR replica pairing for cross-region switchover (MySQL/PostgreSQL).
    failover_dr_replica_name = optional(string, "")

    # MySQL 8.0 automatic minor-version upgrades.
    auto_upgrade_enabled = optional(bool, false)

    # ExecuteSql API posture: ALLOW_DATA_API or DISALLOW_DATA_API.
    data_api_access = optional(string, "")

    # Final backup taken when the instance is deleted.
    final_backup = optional(object({
      enabled        = optional(bool, false)
      retention_days = optional(number, null)
      description    = optional(string, "")
    }), null)

    # Input-only opt-in: move PITR transaction logs from the data disk to
    # Cloud Storage (longer retention, freed disk). Never stored by the API.
    switch_transaction_logs_to_cloud_storage_enabled = optional(bool, false)

    # Input-only opt-in: upgrade read replicas in place alongside the
    # primary on a major database_version change.
    include_replicas_for_major_version_upgrade = optional(bool, false)

    # Irreversible opt-in to the new SQL network architecture for
    # pre-August-2021 projects. Sent only when set (API-computed).
    enforce_new_sql_network_architecture = optional(bool, null)

    # Read replicas only: replication lag (seconds, 300..31536000) beyond
    # which the replica recreates itself. Sent only when set (API-computed).
    replication_lag_max_seconds = optional(number, null)
  })

  validation {
    condition     = var.spec.replication_lag_max_seconds == null || (var.spec.replication_lag_max_seconds >= 300 && var.spec.replication_lag_max_seconds <= 31536000)
    error_message = "replication_lag_max_seconds must be between 300 (five minutes) and 31536000 (one year)."
  }
}
