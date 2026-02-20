resource "oci_core_volume_backup_policy_assignment" "this" {
  count = var.spec.backup_policy_id != null ? 1 : 0

  asset_id       = oci_core_volume.this.id
  policy_id      = var.spec.backup_policy_id.value
}
