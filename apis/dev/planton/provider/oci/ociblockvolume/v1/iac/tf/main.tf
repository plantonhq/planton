resource "oci_core_volume" "this" {
  compartment_id      = var.spec.compartment_id.value
  availability_domain = var.spec.availability_domain
  freeform_tags       = local.freeform_tags

  display_name            = local.display_name != "" ? local.display_name : null
  size_in_gbs             = tostring(var.spec.size_in_gbs)
  vpus_per_gb             = var.spec.vpus_per_gb != null ? tostring(var.spec.vpus_per_gb) : null
  kms_key_id              = var.spec.kms_key_id != null ? var.spec.kms_key_id.value : null
  is_reservations_enabled = var.spec.is_reservations_enabled ? true : null
  xrc_kms_key_id          = var.spec.xrc_kms_key_id != null ? var.spec.xrc_kms_key_id.value : null

  block_volume_replicas_deletion = length(var.spec.block_volume_replicas) == 0

  dynamic "autotune_policies" {
    for_each = var.spec.autotune_policies
    content {
      autotune_type = lookup(local.autotune_type_map, autotune_policies.value.autotune_type, autotune_policies.value.autotune_type)
      max_vpus_per_gb = autotune_policies.value.max_vpus_per_gb > 0 ? tostring(autotune_policies.value.max_vpus_per_gb) : null
    }
  }

  dynamic "block_volume_replicas" {
    for_each = var.spec.block_volume_replicas
    content {
      availability_domain = block_volume_replicas.value.availability_domain
      display_name        = block_volume_replicas.value.display_name != "" ? block_volume_replicas.value.display_name : null
      xrr_kms_key_id      = block_volume_replicas.value.xrr_kms_key_id != null ? block_volume_replicas.value.xrr_kms_key_id.value : null
    }
  }
}
