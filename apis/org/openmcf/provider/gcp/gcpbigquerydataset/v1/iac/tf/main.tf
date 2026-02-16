resource "google_bigquery_dataset" "this" {
  dataset_id = local.dataset_id
  project    = local.project_id
  location   = local.location

  friendly_name                  = local.friendly_name
  description                    = local.description
  default_table_expiration_ms    = local.default_table_expiration_ms
  default_partition_expiration_ms = local.default_partition_expiration_ms
  max_time_travel_hours          = local.max_time_travel_hours
  is_case_insensitive            = var.spec.is_case_insensitive
  default_collation              = local.default_collation
  storage_billing_model          = local.storage_billing_model
  delete_contents_on_destroy     = var.spec.delete_contents_on_destroy

  dynamic "default_encryption_configuration" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = default_encryption_configuration.value
    }
  }

  dynamic "access" {
    for_each = var.spec.access
    content {
      role           = access.value.role != "" ? access.value.role : null
      user_by_email  = access.value.user_by_email != "" ? access.value.user_by_email : null
      group_by_email = access.value.group_by_email != "" ? access.value.group_by_email : null
      domain         = access.value.domain != "" ? access.value.domain : null
      special_group  = access.value.special_group != "" ? access.value.special_group : null
      iam_member     = access.value.iam_member != "" ? access.value.iam_member : null

      dynamic "view" {
        for_each = access.value.view != null ? [access.value.view] : []
        content {
          project_id = view.value.project_id
          dataset_id = view.value.dataset_id
          table_id   = view.value.table_id
        }
      }
    }
  }
}
