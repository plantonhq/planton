locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciBlockVolume"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  autotune_type_map = {
    "detached_volume"  = "DETACHED_VOLUME"
    "performance_based" = "PERFORMANCE_BASED"
  }
}
