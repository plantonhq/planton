locals {
  project_id                     = var.spec.project_id.value
  dataset_id                     = var.spec.dataset_id
  location                       = var.spec.location
  friendly_name                  = var.spec.friendly_name != "" ? var.spec.friendly_name : null
  description                    = var.spec.description != "" ? var.spec.description : null
  default_table_expiration_ms    = var.spec.default_table_expiration_ms > 0 ? var.spec.default_table_expiration_ms : null
  default_partition_expiration_ms = var.spec.default_partition_expiration_ms > 0 ? var.spec.default_partition_expiration_ms : null
  max_time_travel_hours          = var.spec.max_time_travel_hours > 0 ? tostring(var.spec.max_time_travel_hours) : null
  default_collation              = var.spec.default_collation != "" ? var.spec.default_collation : null
  storage_billing_model          = var.spec.storage_billing_model != "" ? var.spec.storage_billing_model : null
  kms_key_name                   = var.spec.kms_key_name != null ? var.spec.kms_key_name.value : null
}
