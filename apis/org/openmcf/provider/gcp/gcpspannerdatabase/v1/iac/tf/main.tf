resource "google_spanner_database" "this" {
  instance = local.instance_name
  name     = local.database_name
  project  = local.project_id

  database_dialect         = local.database_dialect
  version_retention_period = local.version_retention_period
  enable_drop_protection   = var.spec.enable_drop_protection
  default_time_zone        = local.default_time_zone

  ddl = var.spec.ddl

  # Terraform-level deletion protection. Set to false to allow destroy.
  deletion_protection = false

  dynamic "encryption_config" {
    for_each = local.kms_key_name != null ? [local.kms_key_name] : []
    content {
      kms_key_name = encryption_config.value
    }
  }
}
