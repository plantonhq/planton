# ---------------------------------------------------------------------------
# AWS FSx for OpenZFS File System
# ---------------------------------------------------------------------------
# Fully managed NFS file system built on OpenZFS. Supports NFSv3/v4,
# snapshots, cloning, ZSTD/LZ4 compression, and Multi-AZ deployments.
# ForceNew attributes: deployment_type, subnet_ids, security_group_ids,
# kms_key_id, preferred_subnet_id, endpoint_ip_address_range.
# ---------------------------------------------------------------------------

resource "aws_fsx_openzfs_file_system" "this" {
  # Core
  deployment_type    = var.deployment_type
  storage_capacity   = var.storage_capacity_gib
  throughput_capacity = var.throughput_capacity

  # Networking (ForceNew)
  subnet_ids                = var.subnet_ids
  security_group_ids        = length(var.security_group_ids) > 0 ? var.security_group_ids : null
  preferred_subnet_id       = local.preferred_subnet_id
  endpoint_ip_address_range = local.endpoint_ip_address_range
  route_table_ids           = length(var.route_table_ids) > 0 ? var.route_table_ids : null

  # Encryption
  kms_key_id = local.kms_key_id

  # Disk IOPS configuration
  dynamic "disk_iops_configuration" {
    for_each = local.has_disk_iops_configuration ? [1] : []
    content {
      mode = var.disk_iops_mode
      iops = var.disk_iops > 0 ? var.disk_iops : null
    }
  }

  # Root volume configuration
  root_volume_configuration {
    data_compression_type = var.root_data_compression_type
    read_only             = var.root_read_only
    record_size_kib       = var.root_record_size_kib
    copy_tags_to_snapshots = var.root_copy_tags_to_snapshots

    dynamic "nfs_exports" {
      for_each = local.has_root_nfs_exports ? [1] : []
      content {
        dynamic "client_configurations" {
          for_each = var.root_nfs_client_configurations
          content {
            clients = client_configurations.value.clients
            options = client_configurations.value.options
          }
        }
      }
    }

    dynamic "user_and_group_quotas" {
      for_each = var.root_user_and_group_quotas
      content {
        id                        = user_and_group_quotas.value.id
        storage_capacity_quota_gib = user_and_group_quotas.value.storage_capacity_quota_gib
        type                      = user_and_group_quotas.value.type
      }
    }
  }

  # Backup
  automatic_backup_retention_days   = local.automatic_backup_retention_days
  daily_automatic_backup_start_time = local.daily_automatic_backup_start_time
  copy_tags_to_backups              = var.copy_tags_to_backups
  copy_tags_to_volumes              = var.copy_tags_to_volumes
  skip_final_backup                 = var.skip_final_backup

  # Maintenance
  weekly_maintenance_start_time = local.weekly_maintenance_start_time

  tags = local.tags
}
