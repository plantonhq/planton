locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  instance_name = (
    var.spec.instance_name != ""
    ? var.spec.instance_name
    : var.metadata.name
  )

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "alicloud_rocketmq_instance"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.spec.tags)

  internet_enabled = (
    var.spec.internet_info != null
    ? var.spec.internet_info.enabled
    : false
  )

  internet_spec = local.internet_enabled ? "enable" : "disable"

  flow_out_type = (
    local.internet_enabled
    ? (var.spec.internet_info.flow_out_type != "" ? var.spec.internet_info.flow_out_type : "payByTraffic")
    : "uninvolved"
  )

  commodity_code = (
    var.spec.sub_series_code == "serverless"
    ? "ons_rmqsrvlesspost_public_cn"
    : (var.spec.payment_type == "Subscription"
      ? "ons_rmqsub_public_cn"
      : "ons_rmqpost_public_cn"
    )
  )

  topics_map = {
    for t in var.spec.topics : t.topic_name => t
  }

  consumer_groups_map = {
    for cg in var.spec.consumer_groups : cg.consumer_group_id => cg
  }
}
