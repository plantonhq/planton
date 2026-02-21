resource "alicloud_pvtz_zone_record" "records" {
  for_each = local.records_map

  zone_id  = alicloud_pvtz_zone.main.id
  rr       = each.value.rr
  type     = each.value.type
  value    = each.value.value
  ttl      = each.value.ttl
  priority = each.value.type == "MX" ? each.value.priority : null
  remark   = each.value.remark != "" ? each.value.remark : null
}
