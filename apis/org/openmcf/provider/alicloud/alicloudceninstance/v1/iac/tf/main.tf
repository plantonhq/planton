resource "alicloud_cen_instance" "main" {
  cen_instance_name = var.spec.cen_instance_name
  description       = var.spec.description != "" ? var.spec.description : null
  protection_level  = var.spec.protection_level != "" ? var.spec.protection_level : null
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags              = local.final_tags
}

resource "alicloud_cen_instance_attachment" "attachments" {
  for_each = local.attachments_map

  instance_id             = alicloud_cen_instance.main.id
  child_instance_id       = each.value.child_instance_id
  child_instance_type     = each.value.child_instance_type
  child_instance_region_id = each.value.child_instance_region_id
}
