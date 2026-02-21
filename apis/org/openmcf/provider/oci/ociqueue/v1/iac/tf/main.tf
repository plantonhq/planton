resource "oci_queue_queue" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = var.metadata.name
  freeform_tags  = local.freeform_tags

  custom_encryption_key_id         = var.spec.custom_encryption_key_id != null ? var.spec.custom_encryption_key_id.value : null
  dead_letter_queue_delivery_count = var.spec.dead_letter_queue_delivery_count
  retention_in_seconds             = var.spec.retention_in_seconds
  timeout_in_seconds               = var.spec.timeout_in_seconds
  visibility_in_seconds            = var.spec.visibility_in_seconds
  channel_consumption_limit        = var.spec.channel_consumption_limit

  dynamic "capabilities" {
    for_each = local.capabilities
    content {
      type                                                  = capabilities.value.type
      is_primary_consumer_group_enabled                     = capabilities.value.is_primary_consumer_group_enabled
      primary_consumer_group_dead_letter_queue_delivery_count = capabilities.value.primary_consumer_group_dead_letter_queue_delivery_count
      primary_consumer_group_display_name                   = capabilities.value.primary_consumer_group_display_name
    }
  }
}
