locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciQueue"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  capabilities = concat(
    var.spec.is_large_messages_enabled == true ? [{
      type                                                  = "LARGE_MESSAGES"
      is_primary_consumer_group_enabled                     = null
      primary_consumer_group_dead_letter_queue_delivery_count = null
      primary_consumer_group_display_name                   = null
    }] : [],
    var.spec.consumer_group_config != null ? [{
      type                                                  = "CONSUMER_GROUPS"
      is_primary_consumer_group_enabled                     = var.spec.consumer_group_config.is_primary_enabled
      primary_consumer_group_dead_letter_queue_delivery_count = var.spec.consumer_group_config.primary_dead_letter_queue_delivery_count
      primary_consumer_group_display_name                   = var.spec.consumer_group_config.primary_display_name != "" ? var.spec.consumer_group_config.primary_display_name : null
    }] : []
  )
}
