locals {
  project_id    = var.spec.project_id.value
  database_name = var.spec.database_name

  concurrency_mode = var.spec.concurrency_mode != "" ? var.spec.concurrency_mode : null
  pitr_enablement  = var.spec.point_in_time_recovery_enablement != "" ? var.spec.point_in_time_recovery_enablement : null
  database_edition = var.spec.database_edition != "" ? var.spec.database_edition : null
  kms_key_name     = var.spec.kms_key_name != null ? var.spec.kms_key_name.value : null
}
