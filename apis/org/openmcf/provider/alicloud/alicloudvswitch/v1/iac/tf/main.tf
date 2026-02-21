resource "alicloud_vswitch" "main" {
  vpc_id              = var.spec.vpc_id
  zone_id             = var.spec.zone_id
  cidr_block          = var.spec.cidr_block
  vswitch_name        = var.spec.vswitch_name
  description         = var.spec.description != "" ? var.spec.description : null
  enable_ipv6         = var.spec.enable_ipv6
  ipv6_cidr_block_mask = var.spec.ipv6_cidr_block_mask != 0 ? var.spec.ipv6_cidr_block_mask : null
  tags                = local.final_tags
}
