resource "alicloud_alidns_domain" "main" {
  domain_name       = var.spec.domain_name
  group_id          = var.spec.group_id != "" ? var.spec.group_id : null
  remark            = var.spec.remark != "" ? var.spec.remark : null
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags              = local.final_tags
}
