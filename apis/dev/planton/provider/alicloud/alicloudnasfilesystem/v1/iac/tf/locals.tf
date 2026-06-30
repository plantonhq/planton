locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "alicloud_nas_file_system"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.spec.tags)

  has_access_rules = length(var.spec.access_rules) > 0

  access_rules_map = {
    for idx, rule in var.spec.access_rules :
    "${var.metadata.name}-ag-rule-${idx}" => rule
  }
}
