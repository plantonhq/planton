resource "oci_objectstorage_bucket" "this" {
  compartment_id = var.spec.compartment_id.value
  namespace      = var.spec.namespace
  name           = var.spec.name
  freeform_tags  = local.freeform_tags

  access_type           = var.spec.access_type != "" ? lookup(local.access_type_map, var.spec.access_type, null) : null
  storage_tier          = var.spec.storage_tier != "" ? lookup(local.storage_tier_map, var.spec.storage_tier, null) : null
  versioning            = var.spec.versioning != "" ? lookup(local.versioning_map, var.spec.versioning, null) : null
  auto_tiering          = var.spec.auto_tiering != "" ? lookup(local.auto_tiering_map, var.spec.auto_tiering, null) : null
  object_events_enabled = var.spec.object_events_enabled
  kms_key_id            = var.spec.kms_key_id != null ? var.spec.kms_key_id.value : null
  metadata              = length(var.spec.metadata) > 0 ? var.spec.metadata : null

  dynamic "retention_rules" {
    for_each = var.spec.retention_rules
    content {
      display_name = retention_rules.value.display_name

      dynamic "duration" {
        for_each = retention_rules.value.duration != null ? [retention_rules.value.duration] : []
        content {
          time_amount = tostring(duration.value.time_amount)
          time_unit   = lookup(local.time_unit_map, duration.value.time_unit, "DAYS")
        }
      }

      time_rule_locked = retention_rules.value.time_rule_locked != "" ? retention_rules.value.time_rule_locked : null
    }
  }
}
