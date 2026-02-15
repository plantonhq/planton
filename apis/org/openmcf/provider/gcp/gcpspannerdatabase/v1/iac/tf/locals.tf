locals {
  project_id    = var.spec.project_id.value
  instance_name = var.spec.instance.value
  database_name = var.spec.database_name

  database_dialect         = var.spec.database_dialect != "" ? var.spec.database_dialect : null
  version_retention_period = var.spec.version_retention_period != "" ? var.spec.version_retention_period : null
  default_time_zone        = var.spec.default_time_zone != "" ? var.spec.default_time_zone : null
  kms_key_name             = var.spec.kms_key_name != null ? var.spec.kms_key_name.value : null
}
