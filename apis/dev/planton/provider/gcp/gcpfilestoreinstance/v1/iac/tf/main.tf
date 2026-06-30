resource "google_filestore_instance" "this" {
  name     = local.instance_name
  project  = local.project_id
  location = local.location
  tier     = local.tier

  description  = local.description
  protocol     = local.protocol
  kms_key_name = local.kms_key_name

  deletion_protection_enabled = var.spec.deletion_protection_enabled
  deletion_protection_reason  = var.spec.deletion_protection_reason != "" ? var.spec.deletion_protection_reason : null

  labels = local.labels

  file_shares {
    name        = var.spec.file_share.name
    capacity_gb = var.spec.file_share.capacity_gb

    dynamic "nfs_export_options" {
      for_each = var.spec.file_share.nfs_export_options
      content {
        ip_ranges   = length(nfs_export_options.value.ip_ranges) > 0 ? nfs_export_options.value.ip_ranges : null
        access_mode = nfs_export_options.value.access_mode != "" ? nfs_export_options.value.access_mode : null
        squash_mode = nfs_export_options.value.squash_mode != "" ? nfs_export_options.value.squash_mode : null
        anon_uid    = nfs_export_options.value.anon_uid
        anon_gid    = nfs_export_options.value.anon_gid
      }
    }
  }

  networks {
    network           = local.network
    modes             = ["MODE_IPV4"]
    connect_mode      = local.connect_mode
    reserved_ip_range = local.reserved_ip_range
  }

  dynamic "performance_config" {
    for_each = var.spec.performance_config != null ? [var.spec.performance_config] : []
    content {
      dynamic "fixed_iops" {
        for_each = performance_config.value.fixed_iops != null ? [performance_config.value.fixed_iops] : []
        content {
          max_iops = fixed_iops.value.max_iops
        }
      }
      dynamic "iops_per_tb" {
        for_each = performance_config.value.iops_per_tb != null ? [performance_config.value.iops_per_tb] : []
        content {
          max_iops_per_tb = iops_per_tb.value.max_iops_per_tb
        }
      }
    }
  }
}
