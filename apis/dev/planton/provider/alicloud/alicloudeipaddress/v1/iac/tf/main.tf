resource "alicloud_eip_address" "main" {
  address_name         = var.spec.address_name != "" ? var.spec.address_name : null
  description          = var.spec.description != "" ? var.spec.description : null
  bandwidth            = tostring(var.spec.bandwidth)
  internet_charge_type = var.spec.internet_charge_type
  isp                  = var.spec.isp
  resource_group_id    = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags                 = local.final_tags
}
