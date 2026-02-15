resource "google_alloydb_cluster" "cluster" {
  cluster_id = var.spec.cluster_name
  location   = var.spec.location
  project    = var.spec.project_id
  labels     = local.labels

  network_config {
    network = var.spec.network
    allocated_ip_range = var.spec.allocated_ip_range != "" ? var.spec.allocated_ip_range : null
  }

  database_version = var.spec.database_version != "" ? var.spec.database_version : null
  display_name     = var.spec.display_name != "" ? var.spec.display_name : null

  dynamic "initial_user" {
    for_each = var.spec.initial_user != null ? [var.spec.initial_user] : []
    content {
      password = initial_user.value.password
      user     = initial_user.value.user != "" ? initial_user.value.user : null
    }
  }

  dynamic "automated_backup_policy" {
    for_each = var.spec.automated_backup_policy != null ? [var.spec.automated_backup_policy] : []
    content {
      enabled       = automated_backup_policy.value.enabled
      backup_window = automated_backup_policy.value.backup_window != "" ? automated_backup_policy.value.backup_window : null
      location      = automated_backup_policy.value.location != "" ? automated_backup_policy.value.location : null

      dynamic "quantity_based_retention" {
        for_each = automated_backup_policy.value.quantity_based_retention_count > 0 ? [1] : []
        content {
          count = automated_backup_policy.value.quantity_based_retention_count
        }
      }

      dynamic "time_based_retention" {
        for_each = automated_backup_policy.value.time_based_retention_period != "" ? [1] : []
        content {
          retention_period = automated_backup_policy.value.time_based_retention_period
        }
      }

      dynamic "weekly_schedule" {
        for_each = automated_backup_policy.value.weekly_schedule != null ? [automated_backup_policy.value.weekly_schedule] : []
        content {
          days_of_week = length(weekly_schedule.value.days_of_week) > 0 ? weekly_schedule.value.days_of_week : null

          start_times {
            hours = weekly_schedule.value.start_hour
          }
        }
      }

      dynamic "encryption_config" {
        for_each = automated_backup_policy.value.encryption_kms_key_name != "" ? [1] : []
        content {
          kms_key_name = automated_backup_policy.value.encryption_kms_key_name
        }
      }
    }
  }

  dynamic "continuous_backup_config" {
    for_each = var.spec.continuous_backup_config != null ? [var.spec.continuous_backup_config] : []
    content {
      enabled              = continuous_backup_config.value.enabled
      recovery_window_days = continuous_backup_config.value.recovery_window_days > 0 ? continuous_backup_config.value.recovery_window_days : null

      dynamic "encryption_config" {
        for_each = continuous_backup_config.value.encryption_kms_key_name != "" ? [1] : []
        content {
          kms_key_name = continuous_backup_config.value.encryption_kms_key_name
        }
      }
    }
  }

  dynamic "encryption_config" {
    for_each = var.spec.kms_key_name != "" ? [1] : []
    content {
      kms_key_name = var.spec.kms_key_name
    }
  }

  dynamic "maintenance_update_policy" {
    for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
    content {
      maintenance_windows {
        day = maintenance_update_policy.value.day
        start_time {
          hours = maintenance_update_policy.value.start_hour
        }
      }
    }
  }
}

resource "google_alloydb_instance" "primary" {
  cluster       = google_alloydb_cluster.cluster.name
  instance_id   = var.spec.primary_instance.instance_id
  instance_type = "PRIMARY"
  labels        = local.labels

  depends_on = [google_alloydb_cluster.cluster]

  dynamic "machine_config" {
    for_each = var.spec.primary_instance.cpu_count > 0 || var.spec.primary_instance.machine_type != "" ? [1] : []
    content {
      cpu_count    = var.spec.primary_instance.cpu_count > 0 ? var.spec.primary_instance.cpu_count : null
      machine_type = var.spec.primary_instance.machine_type != "" ? var.spec.primary_instance.machine_type : null
    }
  }

  availability_type = var.spec.primary_instance.availability_type != "" ? var.spec.primary_instance.availability_type : null
  database_flags    = length(var.spec.primary_instance.database_flags) > 0 ? var.spec.primary_instance.database_flags : null
  display_name      = var.spec.primary_instance.display_name != "" ? var.spec.primary_instance.display_name : null

  dynamic "query_insights_config" {
    for_each = var.spec.primary_instance.query_insights_config != null ? [var.spec.primary_instance.query_insights_config] : []
    content {
      query_plans_per_minute  = query_insights_config.value.query_plans_per_minute
      query_string_length     = query_insights_config.value.query_string_length
      record_application_tags = query_insights_config.value.record_application_tags
      record_client_address   = query_insights_config.value.record_client_address
    }
  }

  dynamic "client_connection_config" {
    for_each = var.spec.primary_instance.require_connectors || var.spec.primary_instance.ssl_mode != "" ? [1] : []
    content {
      require_connectors = var.spec.primary_instance.require_connectors

      dynamic "ssl_config" {
        for_each = var.spec.primary_instance.ssl_mode != "" ? [1] : []
        content {
          ssl_mode = var.spec.primary_instance.ssl_mode
        }
      }
    }
  }
}
