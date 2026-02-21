resource "alicloud_polardb_database" "databases" {
  for_each = local.databases_map

  db_cluster_id      = alicloud_polardb_cluster.main.id
  db_name            = each.value.db_name
  character_set_name = each.value.character_set_name != "" ? each.value.character_set_name : lookup(local.default_charset, var.spec.db_type, "utf8")
  db_description     = each.value.db_description != "" ? each.value.db_description : null
  collate            = each.value.collate != "" ? each.value.collate : null
  ctype              = each.value.ctype != "" ? each.value.ctype : null
}
