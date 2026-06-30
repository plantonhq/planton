resource "oci_streaming_stream_pool" "this" {
  compartment_id = var.spec.compartment_id.value
  name           = var.metadata.name
  freeform_tags  = local.freeform_tags

  dynamic "kafka_settings" {
    for_each = var.spec.kafka_settings != null ? [var.spec.kafka_settings] : []
    content {
      auto_create_topics_enable = kafka_settings.value.auto_create_topics_enable
      log_retention_hours       = kafka_settings.value.log_retention_hours
      num_partitions            = kafka_settings.value.num_partitions
    }
  }

  dynamic "custom_encryption_key" {
    for_each = var.spec.kms_key_id != null ? [var.spec.kms_key_id] : []
    content {
      kms_key_id = custom_encryption_key.value.value
    }
  }

  dynamic "private_endpoint_settings" {
    for_each = var.spec.private_endpoint_settings != null ? [var.spec.private_endpoint_settings] : []
    content {
      subnet_id           = private_endpoint_settings.value.subnet_id.value
      nsg_ids             = [for n in private_endpoint_settings.value.nsg_ids : n.value]
      private_endpoint_ip = private_endpoint_settings.value.private_endpoint_ip != "" ? private_endpoint_settings.value.private_endpoint_ip : null
    }
  }
}
