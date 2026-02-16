# ---------------------------------------------------------------------------
# Tags and Resource Naming
# ---------------------------------------------------------------------------

locals {
  resource_name = coalesce(var.resource_name, "awsfsxlustrefilesystem")

  tags = merge({
    "Name" = local.resource_name
  }, var.labels)

  # ---------------------------------------------------------------------------
  # Optional string fields — null when not specified
  # ---------------------------------------------------------------------------
  kms_key_id               = var.kms_key_id != "" ? var.kms_key_id : null
  file_system_type_version = var.file_system_type_version != "" ? var.file_system_type_version : null
  import_path              = var.import_path != "" ? var.import_path : null
  export_path              = var.export_path != "" ? var.export_path : null

  # ---------------------------------------------------------------------------
  # Per-unit storage throughput — null when not set (SCRATCH types)
  # ---------------------------------------------------------------------------
  per_unit_storage_throughput = var.per_unit_storage_throughput > 0 ? var.per_unit_storage_throughput : null

  # ---------------------------------------------------------------------------
  # Backup — null when disabled
  # ---------------------------------------------------------------------------
  automatic_backup_retention_days    = var.automatic_backup_retention_days > 0 ? var.automatic_backup_retention_days : null
  daily_automatic_backup_start_time  = var.daily_automatic_backup_start_time != "" ? var.daily_automatic_backup_start_time : null
  weekly_maintenance_start_time      = var.weekly_maintenance_start_time != "" ? var.weekly_maintenance_start_time : null

  # ---------------------------------------------------------------------------
  # Logging — build block only when destination is provided
  # ---------------------------------------------------------------------------
  has_log_configuration = var.log_destination != ""

  # ---------------------------------------------------------------------------
  # Metadata configuration — build block only for PERSISTENT_2 with mode set
  # ---------------------------------------------------------------------------
  has_metadata_configuration = var.metadata_mode != ""
}
