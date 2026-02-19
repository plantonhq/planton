resource "oci_file_storage_file_system" "this" {
  compartment_id      = var.spec.compartment_id.value
  availability_domain = var.spec.availability_domain
  freeform_tags       = local.freeform_tags

  display_name                   = local.display_name != "" ? local.display_name : null
  kms_key_id                     = var.spec.kms_key_id != null ? var.spec.kms_key_id.value : null
  filesystem_snapshot_policy_id  = var.spec.filesystem_snapshot_policy_id != null ? var.spec.filesystem_snapshot_policy_id.value : null
}
