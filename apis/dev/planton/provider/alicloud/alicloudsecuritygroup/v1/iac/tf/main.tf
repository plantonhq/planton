resource "alicloud_security_group" "main" {
  security_group_name = var.spec.security_group_name
  description         = var.spec.description != "" ? var.spec.description : null
  vpc_id              = var.spec.vpc_id
  inner_access_policy = var.spec.inner_access_policy
  resource_group_id   = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags                = local.final_tags
}
