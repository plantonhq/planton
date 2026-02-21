resource "oci_kms_key" "this" {
  compartment_id      = var.spec.compartment_id.value
  display_name        = local.display_name
  management_endpoint = var.spec.management_endpoint.value
  freeform_tags       = local.freeform_tags

  key_shape {
    algorithm = lookup(local.algorithm_map, var.spec.key_shape.algorithm, var.spec.key_shape.algorithm)
    length    = var.spec.key_shape.length
    curve_id  = var.spec.key_shape.curve_id != "" ? lookup(local.curve_id_map, var.spec.key_shape.curve_id, var.spec.key_shape.curve_id) : null
  }

  protection_mode          = var.spec.protection_mode != "" ? lookup(local.protection_mode_map, var.spec.protection_mode, var.spec.protection_mode) : null
  is_auto_rotation_enabled = var.spec.is_auto_rotation_enabled ? true : null

  dynamic "auto_key_rotation_details" {
    for_each = var.spec.auto_key_rotation_details != null ? [var.spec.auto_key_rotation_details] : []
    content {
      rotation_interval_in_days = auto_key_rotation_details.value.rotation_interval_in_days > 0 ? auto_key_rotation_details.value.rotation_interval_in_days : null
      time_of_schedule_start    = auto_key_rotation_details.value.time_of_schedule_start != "" ? auto_key_rotation_details.value.time_of_schedule_start : null
    }
  }

  dynamic "external_key_reference" {
    for_each = var.spec.external_key_reference != null ? [var.spec.external_key_reference] : []
    content {
      external_key_id = external_key_reference.value.external_key_id
    }
  }
}
