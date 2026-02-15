resource "google_memorystore_instance" "this" {
  instance_id = local.instance_name
  project     = local.project_id
  location    = local.location
  shard_count = local.shard_count

  mode           = local.mode
  node_type      = local.node_type
  engine_version = local.engine_version
  engine_configs = length(var.spec.engine_configs) > 0 ? var.spec.engine_configs : null
  replica_count  = local.replica_count

  authorization_mode      = local.authorization_mode
  transit_encryption_mode = local.transit_encryption_mode
  kms_key                 = local.kms_key

  deletion_protection_enabled = local.deletion_protection_enabled

  labels = local.labels

  # PSC auto-created endpoints for VPC connectivity.
  dynamic "desired_auto_created_endpoints" {
    for_each = var.spec.psc_auto_connections
    content {
      network    = desired_auto_created_endpoints.value.network.value
      project_id = desired_auto_created_endpoints.value.project_id.value
    }
  }

  # Persistence configuration (RDB or AOF).
  dynamic "persistence_config" {
    for_each = var.spec.persistence_config != null ? [var.spec.persistence_config] : []
    content {
      mode = persistence_config.value.mode

      dynamic "rdb_config" {
        for_each = persistence_config.value.rdb_config != null ? [persistence_config.value.rdb_config] : []
        content {
          rdb_snapshot_period     = rdb_config.value.rdb_snapshot_period
          rdb_snapshot_start_time = rdb_config.value.rdb_snapshot_start_time != "" ? rdb_config.value.rdb_snapshot_start_time : null
        }
      }

      dynamic "aof_config" {
        for_each = persistence_config.value.aof_config != null ? [persistence_config.value.aof_config] : []
        content {
          append_fsync = aof_config.value.append_fsync
        }
      }
    }
  }

  # Zone distribution configuration.
  dynamic "zone_distribution_config" {
    for_each = var.spec.zone_distribution_config != null ? [var.spec.zone_distribution_config] : []
    content {
      mode = zone_distribution_config.value.mode
      zone = zone_distribution_config.value.zone != "" ? zone_distribution_config.value.zone : null
    }
  }

  # Maintenance policy with weekly maintenance window.
  dynamic "maintenance_policy" {
    for_each = var.spec.maintenance_policy != null ? [var.spec.maintenance_policy] : []
    content {
      weekly_maintenance_window {
        day = maintenance_policy.value.weekly_maintenance_window.day
        start_time {
          hours = maintenance_policy.value.weekly_maintenance_window.hour
        }
      }
    }
  }

  # Automated backup configuration.
  dynamic "automated_backup_config" {
    for_each = var.spec.automated_backup_config != null ? [var.spec.automated_backup_config] : []
    content {
      retention = automated_backup_config.value.retention

      fixed_frequency_schedule {
        start_time {
          hours = automated_backup_config.value.start_hour
        }
      }
    }
  }
}
