resource "alicloud_vpn_gateway" "main" {
  vpn_gateway_name  = var.spec.vpn_gateway_name
  vpc_id            = var.spec.vpc_id
  vswitch_id        = var.spec.vswitch_id
  bandwidth         = var.spec.bandwidth
  description       = var.spec.description != "" ? var.spec.description : null
  payment_type      = var.spec.payment_type
  enable_ssl        = var.spec.enable_ssl
  ssl_connections   = var.spec.ssl_connections
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags              = local.final_tags
}
