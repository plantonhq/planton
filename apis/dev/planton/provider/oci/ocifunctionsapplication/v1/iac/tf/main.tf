resource "oci_functions_application" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = local.display_name
  subnet_ids     = [for s in var.spec.subnet_ids : s.value]
  freeform_tags  = local.freeform_tags

  shape = local.shape

  config = length(var.spec.config) > 0 ? var.spec.config : null

  network_security_group_ids = length(var.spec.network_security_group_ids) > 0 ? [
    for n in var.spec.network_security_group_ids : n.value
  ] : null

  syslog_url = var.spec.syslog_url != "" ? var.spec.syslog_url : null

  dynamic "image_policy_config" {
    for_each = var.spec.image_policy_config != null ? [var.spec.image_policy_config] : []
    content {
      is_policy_enabled = image_policy_config.value.is_policy_enabled

      dynamic "key_details" {
        for_each = image_policy_config.value.key_details
        content {
          kms_key_id = key_details.value.kms_key_id.value
        }
      }
    }
  }

  dynamic "trace_config" {
    for_each = var.spec.trace_config != null ? [var.spec.trace_config] : []
    content {
      is_enabled = trace_config.value.is_enabled
      domain_id  = trace_config.value.domain_id != "" ? trace_config.value.domain_id : null
    }
  }
}
