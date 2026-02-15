resource "google_firestore_database" "this" {
  name        = local.database_name
  location_id = var.spec.location_id
  type        = var.spec.type
  project     = local.project_id

  concurrency_mode                  = local.concurrency_mode
  point_in_time_recovery_enablement = local.pitr_enablement
  delete_protection_state           = var.spec.delete_protection_state
  database_edition                  = local.database_edition

  # Always DELETE to allow IaC lifecycle management. Without this, the
  # default "ABANDON" would leave the database behind on destroy.
  deletion_policy = "DELETE"

  dynamic "cmek_config" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = cmek_config.value
    }
  }
}
