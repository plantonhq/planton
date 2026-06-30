resource "oci_file_storage_mount_target" "this" {
  compartment_id      = var.spec.compartment_id.value
  availability_domain = var.spec.availability_domain
  subnet_id           = var.spec.mount_target.subnet_id.value
  freeform_tags       = local.freeform_tags

  display_name         = var.spec.mount_target.display_name != "" ? var.spec.mount_target.display_name : null
  hostname_label       = var.spec.mount_target.hostname_label != "" ? var.spec.mount_target.hostname_label : null
  ip_address           = var.spec.mount_target.ip_address != "" ? var.spec.mount_target.ip_address : null
  nsg_ids              = length(var.spec.mount_target.nsg_ids) > 0 ? [for n in var.spec.mount_target.nsg_ids : n.value] : null
  requested_throughput = var.spec.mount_target.requested_throughput > 0 ? tostring(var.spec.mount_target.requested_throughput) : null
}

resource "oci_file_storage_export_set" "this" {
  count = (var.spec.mount_target.max_fs_stat_bytes > 0 || var.spec.mount_target.max_fs_stat_files > 0) ? 1 : 0

  mount_target_id  = oci_file_storage_mount_target.this.id
  max_fs_stat_bytes = var.spec.mount_target.max_fs_stat_bytes > 0 ? tostring(var.spec.mount_target.max_fs_stat_bytes) : null
  max_fs_stat_files = var.spec.mount_target.max_fs_stat_files > 0 ? tostring(var.spec.mount_target.max_fs_stat_files) : null
}
