###############################################################################
# Cloud SQL Database Instance
###############################################################################
resource "google_sql_database_instance" "instance" {
  name             = var.metadata.name
  project          = local.project_id
  region           = var.spec.region
  database_version = var.spec.database_version

  settings {
    tier              = var.spec.tier
    disk_size         = var.spec.storage_gb
    disk_type         = "PD_SSD"
    disk_autoresize   = var.spec.disk_auto_resize
    availability_type = local.availability_type
    user_labels       = local.final_gcp_labels

    # IP configuration
    ip_configuration {
      ipv4_enabled    = local.private_ip_enabled ? var.spec.network.ipv4_enabled : true
      private_network = local.private_ip_enabled ? local.vpc_id : null

      # Authorized networks for public IP access
      dynamic "authorized_networks" {
        for_each = var.spec.network != null ? var.spec.network.authorized_networks : []
        content {
          name  = "authorized-network-${authorized_networks.key}"
          value = authorized_networks.value
        }
      }
    }

    # Backup configuration
    backup_configuration {
      enabled = local.backup_enabled
      start_time = local.backup_enabled && var.spec.backup.start_time != null ? (
        var.spec.backup.start_time
      ) : null
      point_in_time_recovery_enabled = local.backup_enabled

      dynamic "backup_retention_settings" {
        for_each = local.backup_enabled ? [1] : []
        content {
          retained_backups = var.spec.backup.retention_days
        }
      }
    }

    # Database flags
    dynamic "database_flags" {
      for_each = local.database_flags_list
      content {
        name  = database_flags.value.name
        value = database_flags.value.value
      }
    }
  }

  # Root password
  root_password = var.spec.root_password

  deletion_protection = var.spec.deletion_protection
}

