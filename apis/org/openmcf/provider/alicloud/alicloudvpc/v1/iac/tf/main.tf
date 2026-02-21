resource "alicloud_vpc" "main" {
  vpc_name         = var.spec.vpc_name
  cidr_block       = var.spec.cidr_block
  description      = var.spec.description != "" ? var.spec.description : null
  enable_ipv6      = var.spec.enable_ipv6
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags             = local.final_tags
}
