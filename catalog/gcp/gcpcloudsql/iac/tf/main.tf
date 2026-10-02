# Enable the Cloud SQL Admin API so a fresh project can host the instance.
# disable_on_destroy is false: tearing down one instance must never disable
# the API for everything else in the project.
resource "google_project_service" "sqladmin_api" {
  project = local.project_id
  service = "sqladmin.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Cloud SQL instance — a managed MySQL, PostgreSQL, or SQL Server server.
# One resource is one instance: a primary, or (with master_instance_name) a
# read replica. Databases and users inside it are separate resources
# (google_sql_database / google_sql_user), composed by instance name.
#
# Lifecycle notes the API enforces:
#   - name/region/CMEK/disk type are immutable; deleted names stay reserved
#     for ~1 week.
#   - database_version upgrades and tier/edition changes apply IN PLACE
#     (with a restart) — no replacement.
#   - disks grow but never shrink; private_network can be set or changed but
#     never removed in place.
#   - a private-IP instance requires the VPC to already carry a service
#     networking connection; the provider prechecks and fails fast because a
#     failed create still burns the reserved instance name.
resource "google_sql_database_instance" "this" {
  name             = var.spec.instance_name
  project          = local.project_id
  region           = var.spec.region
  database_version = var.spec.database_version

  # Set only on read replicas: the primary this instance replicates from.
  master_instance_name = local.master_instance_name

  # CMEK: the key must live in the instance's region. Immutable.
  encryption_key_name = local.encryption_key_name

  # Write-only in GCP — never readable back, never in outputs. Required for
  # SQL Server (spec CEL enforces pre-deploy).
  root_password = local.root_password

  # Engine-side destroy guard: refuses `destroy` at plan time while true.
  # Distinct from settings.deletion_protection_enabled (the API-side guard).
  deletion_protection = var.spec.deletion_protection

  # Engine-side teardown behavior (DELETE / PREVENT / ABANDON) — the
  # ABANDON lever hands the instance to out-of-band management.
  deletion_policy = local.deletion_policy

  # Set READ_POOL_INSTANCE for read pools; CLOUD_SQL_INSTANCE promotes a
  # replica to standalone (with master_instance_name cleared).
  instance_type = local.instance_type

  # Read pools: nodes behind the pool endpoint (autoscaler-owned while
  # read_pool_auto_scale is enabled).
  node_count = var.spec.node_count

  # Pinned patch version; updating restarts the instance.
  maintenance_version = local.maintenance_version

  # Replicas declared from the primary's side (normally left to GCP).
  replica_names = local.replica_names

  # Backup and DR restore trigger: adding or changing it after create runs
  # the restore.
  backupdr_backup = local.backupdr_backup

  # Description recorded on the final backup (final_backup.enabled only).
  final_backup_description = local.final_backup_description

  # Input-only instructions Cloud SQL acts on but never stores: sent as
  # written, never read back, so an explicit false is safe.
  switch_transaction_logs_to_cloud_storage_enabled = var.spec.switch_transaction_logs_to_cloud_storage_enabled
  include_replicas_for_major_version_upgrade       = var.spec.include_replicas_for_major_version_upgrade

  # Irreversible opt-in to the new network architecture. API-computed, so
  # sent only when the spec sets it — an explicit value where the API
  # reports its own would re-plan forever.
  enforce_new_sql_network_architecture = var.spec.enforce_new_sql_network_architecture

  # Create-time clone source: this instance is born as a copy of another.
  dynamic "clone" {
    for_each = var.spec.clone != null ? [var.spec.clone] : []
    content {
      source_instance_name          = clone.value.source_instance_name
      source_project                = clone.value.source_project != "" ? clone.value.source_project : null
      point_in_time                 = clone.value.point_in_time != "" ? clone.value.point_in_time : null
      preferred_zone                = clone.value.preferred_zone != "" ? clone.value.preferred_zone : null
      database_names                = length(clone.value.database_names) > 0 ? clone.value.database_names : null
      allocated_ip_range            = clone.value.allocated_ip_range != "" ? clone.value.allocated_ip_range : null
      source_instance_deletion_time = clone.value.source_instance_deletion_time != "" ? clone.value.source_instance_deletion_time : null
    }
  }

  # Backup-run restore trigger.
  dynamic "restore_backup_context" {
    for_each = var.spec.restore_backup_context != null ? [var.spec.restore_backup_context] : []
    content {
      backup_run_id = restore_backup_context.value.backup_run_id
      instance_id   = restore_backup_context.value.instance_id != "" ? restore_backup_context.value.instance_id : null
      project       = restore_backup_context.value.project != "" ? restore_backup_context.value.project : null
    }
  }

  # Backup and DR point-in-time restore trigger.
  dynamic "point_in_time_restore_context" {
    for_each = var.spec.point_in_time_restore_context != null ? [var.spec.point_in_time_restore_context] : []
    content {
      datasource         = point_in_time_restore_context.value.datasource
      point_in_time      = point_in_time_restore_context.value.point_in_time
      target_instance    = point_in_time_restore_context.value.target_instance != "" ? point_in_time_restore_context.value.target_instance : null
      region             = point_in_time_restore_context.value.region != "" ? point_in_time_restore_context.value.region : null
      preferred_zone     = point_in_time_restore_context.value.preferred_zone != "" ? point_in_time_restore_context.value.preferred_zone : null
      allocated_ip_range = point_in_time_restore_context.value.allocated_ip_range != "" ? point_in_time_restore_context.value.allocated_ip_range : null
    }
  }

  # Cross-region disaster-recovery pairing (MySQL/PostgreSQL): names this
  # primary's DR replica for switchover / replica failover.
  dynamic "replication_cluster" {
    for_each = local.failover_dr_replica_name != null ? [1] : []
    content {
      failover_dr_replica_name = local.failover_dr_replica_name
    }
  }

  settings {
    tier              = var.spec.tier
    edition           = var.spec.edition
    availability_type = var.spec.availability_type
    activation_policy = var.spec.activation_policy

    disk_type             = local.disk_type
    disk_size             = local.disk_size_gb
    disk_autoresize       = local.disk_auto_resize
    disk_autoresize_limit = local.disk_auto_resize_limit

    # API-side delete guard — blocks deletion from console/gcloud/API too.
    deletion_protection_enabled = var.spec.deletion_protection_enabled

    # Keep automated backups (and PITR logs) after the instance is deleted.
    retain_backups_on_delete = var.spec.retain_backups_on_delete

    connector_enforcement = local.connector_enforcement

    # SQL Server-only knobs (spec CEL restricts them to SQLSERVER engines).
    time_zone = local.time_zone
    collation = local.collation

    enable_google_ml_integration = var.spec.enable_google_ml_integration
    enable_dataplex_integration  = var.spec.enable_dataplex_integration

    # MySQL 8.0 automatic minor-version upgrades.
    auto_upgrade_enabled = var.spec.auto_upgrade_enabled

    # ExecuteSql API posture (ALLOW_DATA_API / DISALLOW_DATA_API).
    data_api_access = local.data_api_access

    # Read replicas: self-recreate past this lag. API-computed, so sent only
    # when the spec sets it (null otherwise).
    replication_lag_max_seconds = var.spec.replication_lag_max_seconds

    # HYPERDISK_BALANCED provisioned performance (spec CEL gates the disk
    # type).
    data_disk_provisioned_iops       = try(var.spec.disk.provisioned_iops, null)
    data_disk_provisioned_throughput = try(var.spec.disk.provisioned_throughput, null)

    user_labels = local.final_labels

    # Emitted only when enabled: the API rejects a data-cache stanza on
    # ENTERPRISE instances (spec CEL already forces ENTERPRISE_PLUS).
    dynamic "data_cache_config" {
      for_each = local.data_cache_enabled ? [1] : []
      content {
        data_cache_enabled = true
      }
    }

    # SQL Server licensing/performance dial.
    dynamic "advanced_machine_features" {
      for_each = var.spec.threads_per_core != null ? [1] : []
      content {
        threads_per_core = var.spec.threads_per_core
      }
    }

    ip_configuration {
      # An omitted spec network block resolves to public IPv4 with no
      # authorized networks — Auth Proxy / connector access only.
      ipv4_enabled    = local.ipv4_enabled
      private_network = local.private_network

      allocated_ip_range                            = local.allocated_ip_range
      enable_private_path_for_google_cloud_services = local.enable_private_path

      ssl_mode       = local.ssl_mode
      server_ca_mode = local.server_ca_mode
      server_ca_pool = local.server_ca_pool

      # Automatic server certificate rotation (CAS CA modes only; spec CEL
      # gates the pairing).
      server_certificate_rotation_mode = local.server_certificate_rotation_mode

      custom_subject_alternative_names = local.custom_sans

      dynamic "authorized_networks" {
        for_each = local.authorized_networks
        content {
          value           = authorized_networks.value.value
          name            = authorized_networks.value.name != "" ? authorized_networks.value.name : null
          expiration_time = authorized_networks.value.expiration_time != "" ? authorized_networks.value.expiration_time : null
        }
      }

      dynamic "psc_config" {
        for_each = local.psc != null ? [local.psc] : []
        content {
          psc_enabled               = true
          allowed_consumer_projects = psc_config.value.allowed_consumer_projects
          network_attachment_uri    = psc_config.value.network_attachment_uri != "" ? psc_config.value.network_attachment_uri : null

          # DNS automation for PSC endpoints (write-endpoint DNS is
          # Enterprise Plus only).
          psc_auto_dns_enabled           = psc_config.value.auto_dns_enabled
          psc_write_endpoint_dns_enabled = psc_config.value.write_endpoint_dns_enabled

          # Service Connection Policy for the auto connections. API-computed,
          # so sent only when the spec sets it.
          psc_auto_connection_policy_enabled = psc_config.value.auto_connection_policy_enabled

          dynamic "psc_auto_connections" {
            for_each = psc_config.value.auto_connections
            content {
              consumer_network            = psc_auto_connections.value.consumer_network
              consumer_service_project_id = psc_auto_connections.value.consumer_service_project_id != "" ? psc_auto_connections.value.consumer_service_project_id : null
            }
          }
        }
      }
    }

    dynamic "location_preference" {
      for_each = local.location_preference != null ? [local.location_preference] : []
      content {
        zone           = location_preference.value.zone != "" ? location_preference.value.zone : null
        secondary_zone = location_preference.value.secondary_zone != "" ? location_preference.value.secondary_zone : null
      }
    }

    dynamic "backup_configuration" {
      for_each = var.spec.backup != null ? [var.spec.backup] : []
      content {
        enabled    = backup_configuration.value.enabled
        start_time = backup_configuration.value.start_time != "" ? backup_configuration.value.start_time : null
        location   = backup_configuration.value.location != "" ? backup_configuration.value.location : null

        # MySQL PITR mechanism; also required for MySQL replicas and HA.
        binary_log_enabled = backup_configuration.value.binary_log_enabled

        # PostgreSQL / SQL Server PITR mechanism.
        point_in_time_recovery_enabled = backup_configuration.value.point_in_time_recovery_enabled

        transaction_log_retention_days = backup_configuration.value.transaction_log_retention_days

        dynamic "backup_retention_settings" {
          for_each = local.backup_retention_settings
          content {
            retained_backups = backup_configuration.value.retained_backups
            retention_unit   = backup_configuration.value.retention_unit != "" ? backup_configuration.value.retention_unit : "COUNT"
          }
        }
      }
    }

    dynamic "maintenance_window" {
      for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
      content {
        day          = maintenance_window.value.day
        hour         = maintenance_window.value.hour
        update_track = maintenance_window.value.update_track != "" ? maintenance_window.value.update_track : null
      }
    }

    dynamic "deny_maintenance_period" {
      for_each = var.spec.deny_maintenance_period != null ? [var.spec.deny_maintenance_period] : []
      content {
        start_date = deny_maintenance_period.value.start_date
        end_date   = deny_maintenance_period.value.end_date
        time       = deny_maintenance_period.value.time
      }
    }

    dynamic "insights_config" {
      # Emitted when either telemetry tier is on. try() because HCL's &&
      # does not short-circuit on the nullable block.
      for_each = (
        try(var.spec.insights_config.query_insights_enabled, false) || try(var.spec.insights_config.enhanced_query_insights_enabled, false)
      ) ? [var.spec.insights_config] : []
      content {
        query_insights_enabled          = insights_config.value.query_insights_enabled
        enhanced_query_insights_enabled = insights_config.value.enhanced_query_insights_enabled
        query_string_length             = insights_config.value.query_string_length
        record_application_tags         = insights_config.value.record_application_tags
        record_client_address           = insights_config.value.record_client_address
        query_plans_per_minute          = insights_config.value.query_plans_per_minute
      }
    }

    dynamic "password_validation_policy" {
      for_each = var.spec.password_validation_policy != null ? [var.spec.password_validation_policy] : []
      content {
        enable_password_policy      = password_validation_policy.value.enable_password_policy
        min_length                  = password_validation_policy.value.min_length
        complexity                  = password_validation_policy.value.complexity != "" ? password_validation_policy.value.complexity : null
        reuse_interval              = password_validation_policy.value.reuse_interval
        disallow_username_substring = password_validation_policy.value.disallow_username_substring
        password_change_interval    = password_validation_policy.value.password_change_interval != "" ? password_validation_policy.value.password_change_interval : null
      }
    }

    dynamic "connection_pool_config" {
      # try() because HCL's && does not short-circuit on the nullable block.
      for_each = try(var.spec.connection_pooling.enabled, false) ? [var.spec.connection_pooling] : []
      content {
        connection_pooling_enabled = true

        dynamic "flags" {
          for_each = local.connection_pool_flags_list
          content {
            name  = flags.value.name
            value = flags.value.value
          }
        }
      }
    }

    dynamic "sql_server_audit_config" {
      for_each = var.spec.sql_server_audit_config != null ? [var.spec.sql_server_audit_config] : []
      content {
        bucket             = sql_server_audit_config.value.bucket != "" ? sql_server_audit_config.value.bucket : null
        retention_interval = sql_server_audit_config.value.retention_interval != "" ? sql_server_audit_config.value.retention_interval : null
        upload_interval    = sql_server_audit_config.value.upload_interval != "" ? sql_server_audit_config.value.upload_interval : null
      }
    }

    # SQL Server Active Directory join — managed AD by default; the
    # customer-managed mode bootstraps through domain controllers and a
    # Secret Manager admin credential.
    dynamic "active_directory_config" {
      for_each = var.spec.active_directory != null ? [var.spec.active_directory] : []
      content {
        domain                       = active_directory_config.value.domain
        mode                         = active_directory_config.value.mode != "" ? active_directory_config.value.mode : null
        dns_servers                  = length(active_directory_config.value.dns_servers) > 0 ? active_directory_config.value.dns_servers : null
        admin_credential_secret_name = active_directory_config.value.admin_credential_secret_name != "" ? active_directory_config.value.admin_credential_secret_name : null
        organizational_unit          = active_directory_config.value.organizational_unit != "" ? active_directory_config.value.organizational_unit : null
      }
    }

    # SQL Server Microsoft Entra ID authentication (paired IDs).
    dynamic "entraid_config" {
      for_each = var.spec.entra_id != null ? [var.spec.entra_id] : []
      content {
        application_id = entraid_config.value.application_id
        tenant_id      = entraid_config.value.tenant_id
      }
    }

    # Final backup on delete — the safety net that survives the teardown.
    dynamic "final_backup_config" {
      # try() because HCL's && does not short-circuit on the nullable block.
      for_each = try(var.spec.final_backup.enabled, false) ? [var.spec.final_backup] : []
      content {
        enabled        = true
        retention_days = final_backup_config.value.retention_days
      }
    }

    # Read pool auto scaling between node-count bounds.
    dynamic "read_pool_auto_scale_config" {
      for_each = var.spec.read_pool_auto_scale != null ? [var.spec.read_pool_auto_scale] : []
      content {
        enabled                    = read_pool_auto_scale_config.value.enabled
        min_node_count             = read_pool_auto_scale_config.value.min_node_count
        max_node_count             = read_pool_auto_scale_config.value.max_node_count
        disable_scale_in           = read_pool_auto_scale_config.value.disable_scale_in
        scale_in_cooldown_seconds  = read_pool_auto_scale_config.value.scale_in_cooldown_seconds
        scale_out_cooldown_seconds = read_pool_auto_scale_config.value.scale_out_cooldown_seconds

        dynamic "target_metrics" {
          for_each = read_pool_auto_scale_config.value.target_metrics
          content {
            metric       = target_metrics.value.metric
            target_value = target_metrics.value.target_value
          }
        }
      }
    }

    dynamic "database_flags" {
      for_each = local.database_flags_list
      content {
        name  = database_flags.value.name
        value = database_flags.value.value
      }
    }
  }

  # Replica behavior + external-source replication channel. Only meaningful
  # with master_instance_name (spec CEL enforces the pairing).
  dynamic "replica_configuration" {
    for_each = var.spec.replica_configuration != null ? [var.spec.replica_configuration] : []
    content {
      failover_target           = replica_configuration.value.failover_target
      cascadable_replica        = replica_configuration.value.cascadable_replica
      username                  = replica_configuration.value.username != "" ? replica_configuration.value.username : null
      password                  = replica_configuration.value.password != "" ? replica_configuration.value.password : null
      ca_certificate            = replica_configuration.value.ca_certificate != "" ? replica_configuration.value.ca_certificate : null
      client_certificate        = replica_configuration.value.client_certificate != "" ? replica_configuration.value.client_certificate : null
      client_key                = replica_configuration.value.client_key != "" ? replica_configuration.value.client_key : null
      dump_file_path            = replica_configuration.value.dump_file_path != "" ? replica_configuration.value.dump_file_path : null
      connect_retry_interval    = replica_configuration.value.connect_retry_interval
      master_heartbeat_period   = replica_configuration.value.master_heartbeat_period
      ssl_cipher                = replica_configuration.value.ssl_cipher != "" ? replica_configuration.value.ssl_cipher : null
      verify_server_certificate = replica_configuration.value.verify_server_certificate
    }
  }

  depends_on = [google_project_service.sqladmin_api]
}
