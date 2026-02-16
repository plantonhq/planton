resource "google_redis_instance" "this" {
  name           = local.instance_name
  project        = local.project_id
  region         = local.region
  tier           = local.tier
  memory_size_gb = local.memory_size_gb

  redis_version           = local.redis_version
  display_name            = local.display_name
  location_id             = local.location_id
  authorized_network      = local.authorized_network
  connect_mode            = local.connect_mode
  reserved_ip_range       = local.reserved_ip_range
  auth_enabled            = var.spec.auth_enabled
  transit_encryption_mode = local.transit_encryption_mode
  redis_configs           = length(var.spec.redis_configs) > 0 ? var.spec.redis_configs : null
  read_replicas_mode      = local.read_replicas_mode
  replica_count           = local.replica_count
  customer_managed_key = local.customer_managed_key

  labels = local.labels

  # Note: deletion_protection is handled at the Pulumi level.
  # For Terraform, use lifecycle { prevent_destroy = true } in your root module.

  dynamic "maintenance_policy" {
    for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
    content {
      weekly_maintenance_window {
        day = maintenance_policy.value.day
        start_time {
          hours = maintenance_policy.value.hour
        }
      }
    }
  }

  dynamic "persistence_config" {
    for_each = var.spec.persistence_config != null ? [var.spec.persistence_config] : []
    content {
      persistence_mode    = persistence_config.value.persistence_mode
      rdb_snapshot_period = persistence_config.value.rdb_snapshot_period != "" ? persistence_config.value.rdb_snapshot_period : null
    }
  }
}
