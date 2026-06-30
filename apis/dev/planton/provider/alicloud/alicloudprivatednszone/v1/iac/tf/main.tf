resource "alicloud_pvtz_zone" "main" {
  zone_name         = var.spec.zone_name
  remark            = var.spec.remark != "" ? var.spec.remark : null
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags              = local.final_tags
}

resource "alicloud_pvtz_zone_attachment" "main" {
  zone_id = alicloud_pvtz_zone.main.id

  dynamic "vpcs" {
    for_each = var.spec.vpc_attachments
    content {
      vpc_id    = vpcs.value.vpc_id
      region_id = vpcs.value.region_id != "" ? vpcs.value.region_id : null
    }
  }
}
