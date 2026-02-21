locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  load_balancer_name = (
    var.spec.load_balancer_name != ""
    ? var.spec.load_balancer_name
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "alicloud_alb_load_balancer"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.spec.tags)

  server_groups_map = {
    for sg in var.spec.server_groups : sg.name => sg
  }

  listeners_map = {
    for l in var.spec.listeners : "${l.listener_port}-${l.listener_protocol}" => l
  }
}
