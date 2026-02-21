locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciLogGroup"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  log_type_map = {
    "custom"  = "CUSTOM"
    "service" = "SERVICE"
  }
}
