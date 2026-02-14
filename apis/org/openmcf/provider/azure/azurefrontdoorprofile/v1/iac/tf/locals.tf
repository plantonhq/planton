locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_front_door_profile"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Flatten origins nested within origin groups for for_each iteration.
  # Each origin is keyed as "{origin_group_name}/{origin_name}".
  origins_flat = merge([
    for group in var.spec.origin_groups : {
      for origin in group.origins :
      "${group.name}/${origin.name}" => merge(origin, { origin_group_name = group.name })
    }
  ]...)
}
