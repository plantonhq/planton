locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_event_hub_namespace"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Flatten consumer groups for for_each iteration.
  # Each consumer group is keyed as "{eventhub_name}/{consumer_group_name}".
  consumer_groups = flatten([
    for eh in var.spec.event_hubs : [
      for cg in eh.consumer_groups : {
        key           = "${eh.name}/${cg.name}"
        eventhub_name = eh.name
        name          = cg.name
        user_metadata = cg.user_metadata
      }
    ]
  ])

  consumer_groups_map = { for cg in local.consumer_groups : cg.key => cg }
}
