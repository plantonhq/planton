resource "oci_logging_log" "this" {
  for_each = { for l in var.spec.logs : l.display_name => l }

  display_name       = each.value.display_name
  log_group_id       = oci_logging_log_group.this.id
  log_type           = local.log_type_map[each.value.log_type]
  is_enabled         = each.value.is_enabled
  retention_duration = each.value.retention_duration
  freeform_tags      = local.freeform_tags

  dynamic "configuration" {
    for_each = each.value.configuration != null ? [each.value.configuration] : []
    content {
      compartment_id = try(configuration.value.compartment_id.value, null)
      source {
        source_type = "OCISERVICE"
        service     = configuration.value.service
        resource    = configuration.value.resource.value
        category    = configuration.value.category
        parameters  = length(configuration.value.parameters) > 0 ? configuration.value.parameters : null
      }
    }
  }
}
