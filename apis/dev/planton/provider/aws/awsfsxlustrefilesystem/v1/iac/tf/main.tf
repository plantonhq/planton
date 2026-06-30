# ---------------------------------------------------------------------------
# AWS FSx Lustre File System
# ---------------------------------------------------------------------------
# High-performance parallel file system for compute-intensive workloads.
# ForceNew attributes: deployment_type, storage_type, subnet_ids,
# security_group_ids, kms_key_id, import_path, export_path,
# file_system_type_version, copy_tags_to_backups.
# ---------------------------------------------------------------------------

resource "aws_fsx_lustre_file_system" "this" {
  # Core
  deployment_type          = var.deployment_type
  storage_capacity         = var.storage_capacity_gib
  storage_type             = var.storage_type
  per_unit_storage_throughput = local.per_unit_storage_throughput
  data_compression_type    = var.data_compression_type
  file_system_type_version = local.file_system_type_version

  # Networking (ForceNew — single AZ)
  subnet_ids         = [var.subnet_id]
  security_group_ids = length(var.security_group_ids) > 0 ? var.security_group_ids : null

  # Encryption
  kms_key_id = local.kms_key_id

  # S3 data repository (legacy, ForceNew)
  import_path = local.import_path
  export_path = local.export_path

  # Logging
  dynamic "log_configuration" {
    for_each = local.has_log_configuration ? [1] : []
    content {
      destination = var.log_destination
      level       = var.log_level
    }
  }

  # Backup (PERSISTENT only)
  automatic_backup_retention_days    = local.automatic_backup_retention_days
  daily_automatic_backup_start_time  = local.daily_automatic_backup_start_time
  copy_tags_to_backups               = var.copy_tags_to_backups
  skip_final_backup                  = var.skip_final_backup

  # Maintenance
  weekly_maintenance_start_time = local.weekly_maintenance_start_time

  # Metadata configuration (PERSISTENT_2 only)
  dynamic "metadata_configuration" {
    for_each = local.has_metadata_configuration ? [1] : []
    content {
      mode = var.metadata_mode
      iops = var.metadata_iops > 0 ? var.metadata_iops : null
    }
  }

  tags = local.tags
}
