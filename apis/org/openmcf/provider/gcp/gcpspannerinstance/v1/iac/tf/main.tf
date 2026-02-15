resource "google_spanner_instance" "this" {
  name         = local.instance_name
  project      = local.project_id
  config       = local.config
  display_name = local.display_name

  num_nodes        = local.num_nodes
  processing_units = local.processing_units

  instance_type                = local.instance_type
  edition                      = local.edition
  default_backup_schedule_type = local.default_backup_schedule_type
  force_destroy                = var.spec.force_destroy

  labels = local.labels

  dynamic "autoscaling_config" {
    for_each = var.spec.autoscaling_config != null ? [var.spec.autoscaling_config] : []
    content {
      autoscaling_limits {
        min_nodes            = autoscaling_config.value.autoscaling_limits.min_nodes > 0 ? autoscaling_config.value.autoscaling_limits.min_nodes : null
        max_nodes            = autoscaling_config.value.autoscaling_limits.max_nodes > 0 ? autoscaling_config.value.autoscaling_limits.max_nodes : null
        min_processing_units = autoscaling_config.value.autoscaling_limits.min_processing_units > 0 ? autoscaling_config.value.autoscaling_limits.min_processing_units : null
        max_processing_units = autoscaling_config.value.autoscaling_limits.max_processing_units > 0 ? autoscaling_config.value.autoscaling_limits.max_processing_units : null
      }

      dynamic "autoscaling_targets" {
        for_each = autoscaling_config.value.autoscaling_targets != null ? [autoscaling_config.value.autoscaling_targets] : []
        content {
          high_priority_cpu_utilization_percent = autoscaling_targets.value.high_priority_cpu_utilization_percent > 0 ? autoscaling_targets.value.high_priority_cpu_utilization_percent : null
          storage_utilization_percent           = autoscaling_targets.value.storage_utilization_percent > 0 ? autoscaling_targets.value.storage_utilization_percent : null
        }
      }
    }
  }
}
