resource "alicloud_nas_access_group" "custom" {
  count = local.has_access_rules ? 1 : 0

  access_group_name = "${var.metadata.name}-ag"
  access_group_type = "Vpc"
  file_system_type  = var.spec.file_system_type
  description       = "Access group for NAS file system ${var.metadata.name}"
}

resource "alicloud_nas_access_rule" "rules" {
  for_each = local.access_rules_map

  access_group_name = alicloud_nas_access_group.custom[0].access_group_name
  source_cidr_ip    = each.value.source_cidr_ip
  rw_access_type    = each.value.rw_access_type
  user_access_type  = each.value.user_access_type
  priority          = each.value.priority
  file_system_type  = var.spec.file_system_type
}
