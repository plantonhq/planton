resource "alicloud_cr_ee_namespace" "namespaces" {
  for_each = local.namespaces_map

  instance_id        = alicloud_cr_ee_instance.main.id
  name               = each.value.name
  auto_create        = each.value.auto_create
  default_visibility = each.value.default_visibility
}
