locals {
  key_ring_id = var.spec.key_ring_id.value
  key_name    = var.spec.key_name
  purpose     = var.spec.purpose != "" ? var.spec.purpose : null
  rotation_period = var.spec.rotation_period != "" ? var.spec.rotation_period : null
  destroy_scheduled_duration = var.spec.destroy_scheduled_duration != "" ? var.spec.destroy_scheduled_duration : null
  skip_initial_version_creation = var.spec.skip_initial_version_creation
  version_template = var.spec.version_template
}
