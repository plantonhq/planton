# ---------------------------------------------------------------------------
# AWS FSx for ONTAP File System
# ---------------------------------------------------------------------------
# Enterprise NAS/SAN file system built on NetApp ONTAP. Supports multi-protocol
# access (NFS, SMB, iSCSI), scale-out HA pairs, SnapMirror replication, and
# features like instant snapshots, cloning, compression, and deduplication.
# ForceNew attributes: deployment_type, subnet_ids, security_group_ids,
# kms_key_id, storage_type, preferred_subnet_id, endpoint_ip_address_range.
# ---------------------------------------------------------------------------

resource "aws_fsx_ontap_file_system" "this" {
  # Core
  deployment_type                 = var.deployment_type
  storage_capacity                 = var.storage_capacity_gib
  storage_type                     = var.storage_type
  throughput_capacity_per_ha_pair  = var.throughput_capacity_per_ha_pair
  ha_pairs                         = var.ha_pairs

  # Networking (ForceNew)
  subnet_ids                = var.subnet_ids
  security_group_ids        = length(var.security_group_ids) > 0 ? var.security_group_ids : null
  preferred_subnet_id       = local.preferred_subnet_id
  endpoint_ip_address_range = local.endpoint_ip_address_range
  route_table_ids           = length(var.route_table_ids) > 0 ? var.route_table_ids : null

  # Encryption
  kms_key_id = local.kms_key_id

  # ONTAP administration
  fsx_admin_password = local.fsx_admin_password

  # Disk IOPS configuration
  dynamic "disk_iops_configuration" {
    for_each = local.has_disk_iops_configuration ? [1] : []
    content {
      mode = var.disk_iops_mode
      iops = var.disk_iops > 0 ? var.disk_iops : null
    }
  }

  # Backup (copy_tags_to_backups and skip_final_backup not supported by aws_fsx_ontap_file_system)
  automatic_backup_retention_days   = local.automatic_backup_retention_days
  daily_automatic_backup_start_time = local.daily_automatic_backup_start_time

  # Maintenance
  weekly_maintenance_start_time = local.weekly_maintenance_start_time

  tags = local.tags
}
