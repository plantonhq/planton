resource "oci_objectstorage_object_lifecycle_policy" "this" {
  count = length(var.spec.lifecycle_rules) > 0 ? 1 : 0

  bucket    = oci_objectstorage_bucket.this.name
  namespace = var.spec.namespace

  dynamic "rules" {
    for_each = var.spec.lifecycle_rules
    content {
      name        = rules.value.name
      action      = lookup(local.lifecycle_action_map, rules.value.action, rules.value.action)
      is_enabled  = rules.value.is_enabled
      time_amount = tostring(rules.value.time_amount)
      time_unit   = lookup(local.time_unit_map, rules.value.time_unit, "DAYS")
      target      = rules.value.target

      dynamic "object_name_filter" {
        for_each = rules.value.object_name_filter != null ? [rules.value.object_name_filter] : []
        content {
          inclusion_patterns = length(object_name_filter.value.inclusion_patterns) > 0 ? object_name_filter.value.inclusion_patterns : null
          inclusion_prefixes = length(object_name_filter.value.inclusion_prefixes) > 0 ? object_name_filter.value.inclusion_prefixes : null
          exclusion_patterns = length(object_name_filter.value.exclusion_patterns) > 0 ? object_name_filter.value.exclusion_patterns : null
        }
      }
    }
  }
}
