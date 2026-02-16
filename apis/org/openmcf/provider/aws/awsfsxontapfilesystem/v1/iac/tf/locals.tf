locals {
  resource_name = coalesce(var.resource_name, "awsfsxontapfilesystem")

  tags = merge({
    "Name" = local.resource_name
  }, var.labels)

  kms_key_id                = var.kms_key_id != "" ? var.kms_key_id : null
  preferred_subnet_id       = var.preferred_subnet_id != "" ? var.preferred_subnet_id : var.subnet_ids[0]
  endpoint_ip_address_range = var.endpoint_ip_address_range != "" ? var.endpoint_ip_address_range : null
  fsx_admin_password        = var.fsx_admin_password != "" ? var.fsx_admin_password : null

  automatic_backup_retention_days   = var.automatic_backup_retention_days > 0 ? var.automatic_backup_retention_days : null
  daily_automatic_backup_start_time = var.daily_automatic_backup_start_time != "" ? var.daily_automatic_backup_start_time : null
  weekly_maintenance_start_time     = var.weekly_maintenance_start_time != "" ? var.weekly_maintenance_start_time : null

  has_disk_iops_configuration = var.disk_iops_mode != ""
}
