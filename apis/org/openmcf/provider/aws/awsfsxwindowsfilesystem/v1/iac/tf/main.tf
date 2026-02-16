resource "aws_fsx_windows_file_system" "this" {
  deployment_type    = var.spec.deployment_type
  storage_capacity   = var.spec.storage_capacity_gib
  storage_type       = var.spec.storage_type
  throughput_capacity = var.spec.throughput_capacity
  subnet_ids         = var.spec.subnet_ids
  preferred_subnet_id = var.spec.preferred_subnet_id
  security_group_ids = length(var.spec.security_group_ids) > 0 ? var.spec.security_group_ids : null
  kms_key_id         = var.spec.kms_key_id
  active_directory_id = var.spec.active_directory_id
  aliases            = length(var.spec.aliases) > 0 ? var.spec.aliases : null

  dynamic "self_managed_active_directory" {
    for_each = var.spec.self_managed_active_directory != null ? [var.spec.self_managed_active_directory] : []
    content {
      domain_name                            = self_managed_active_directory.value.domain_name
      dns_ips                                = self_managed_active_directory.value.dns_ips
      username                               = self_managed_active_directory.value.username
      password                               = self_managed_active_directory.value.password
      domain_join_service_account_secret     = self_managed_active_directory.value.domain_join_service_account_secret_arn
      file_system_administrators_group       = self_managed_active_directory.value.file_system_administrators_group
      organizational_unit_distinguished_name = self_managed_active_directory.value.organizational_unit_distinguished_name
    }
  }

  dynamic "audit_log_configuration" {
    for_each = var.spec.audit_log_configuration != null ? [var.spec.audit_log_configuration] : []
    content {
      file_access_audit_log_level       = audit_log_configuration.value.file_access_audit_log_level
      file_share_access_audit_log_level = audit_log_configuration.value.file_share_access_audit_log_level
      audit_log_destination             = audit_log_configuration.value.audit_log_destination
    }
  }

  dynamic "disk_iops_configuration" {
    for_each = var.spec.disk_iops_configuration != null ? [var.spec.disk_iops_configuration] : []
    content {
      mode = disk_iops_configuration.value.mode
      iops = disk_iops_configuration.value.iops
    }
  }

  automatic_backup_retention_days   = var.spec.automatic_backup_retention_days
  daily_automatic_backup_start_time = var.spec.daily_automatic_backup_start_time
  copy_tags_to_backups              = var.spec.copy_tags_to_backups
  skip_final_backup                 = var.spec.skip_final_backup
  weekly_maintenance_start_time     = var.spec.weekly_maintenance_start_time

  tags = local.tags

  lifecycle {
    ignore_changes = [tags["CreatedAt"]]
  }
}
