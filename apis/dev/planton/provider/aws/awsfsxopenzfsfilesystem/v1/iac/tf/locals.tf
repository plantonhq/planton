# ---------------------------------------------------------------------------
# Tags and Resource Naming
# ---------------------------------------------------------------------------

locals {
  resource_name = coalesce(var.resource_name, "awsfsxopenzfsfilesystem")

  tags = merge({
    "Name" = local.resource_name
  }, var.labels)

  # ---------------------------------------------------------------------------
  # Optional string fields — null when not specified
  # ---------------------------------------------------------------------------
  kms_key_id                = var.kms_key_id != "" ? var.kms_key_id : null
  preferred_subnet_id       = var.preferred_subnet_id != "" ? var.preferred_subnet_id : null
  endpoint_ip_address_range = var.endpoint_ip_address_range != "" ? var.endpoint_ip_address_range : null

  # ---------------------------------------------------------------------------
  # Backup — null when disabled
  # ---------------------------------------------------------------------------
  automatic_backup_retention_days   = var.automatic_backup_retention_days > 0 ? var.automatic_backup_retention_days : null
  daily_automatic_backup_start_time = var.daily_automatic_backup_start_time != "" ? var.daily_automatic_backup_start_time : null
  weekly_maintenance_start_time     = var.weekly_maintenance_start_time != "" ? var.weekly_maintenance_start_time : null

  # ---------------------------------------------------------------------------
  # Disk IOPS — build block only when mode is provided
  # ---------------------------------------------------------------------------
  has_disk_iops_configuration = var.disk_iops_mode != ""

  # ---------------------------------------------------------------------------
  # Root volume — always build the block (uses defaults when omitted)
  # ---------------------------------------------------------------------------
  has_root_nfs_exports       = length(var.root_nfs_client_configurations) > 0
  has_root_user_group_quotas = length(var.root_user_and_group_quotas) > 0
}
