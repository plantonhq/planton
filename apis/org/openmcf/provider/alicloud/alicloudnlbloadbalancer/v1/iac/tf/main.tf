resource "alicloud_nlb_load_balancer" "main" {
  load_balancer_name = local.load_balancer_name
  vpc_id             = var.spec.vpc_id
  address_type       = var.spec.address_type
  load_balancer_type = "Network"
  payment_type       = "PayAsYouGo"
  cross_zone_enabled = var.spec.cross_zone_enabled
  resource_group_id  = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags               = local.final_tags

  dynamic "zone_mappings" {
    for_each = var.spec.zone_mappings
    content {
      zone_id       = zone_mappings.value.zone_id
      vswitch_id    = zone_mappings.value.vswitch_id
      allocation_id = zone_mappings.value.allocation_id != "" ? zone_mappings.value.allocation_id : null
    }
  }
}
