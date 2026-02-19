resource "alicloud_db_database" "databases" {
  for_each = local.databases_map

  instance_id    = alicloud_db_instance.main.id
  data_base_name = each.value.name
  character_set  = each.value.character_set != "" ? each.value.character_set : lookup(local.default_charset, var.spec.engine, "utf8")
  description    = each.value.description != "" ? each.value.description : null
}
