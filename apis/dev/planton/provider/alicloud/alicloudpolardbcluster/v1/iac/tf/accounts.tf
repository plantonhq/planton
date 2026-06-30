resource "alicloud_polardb_account" "accounts" {
  for_each = local.accounts_map

  db_cluster_id       = alicloud_polardb_cluster.main.id
  account_name        = each.value.account_name
  account_password    = each.value.account_password
  account_type        = each.value.account_type
  account_description = each.value.account_description != "" ? each.value.account_description : null

  depends_on = [alicloud_polardb_database.databases]
}

resource "alicloud_polardb_account_privilege" "privileges" {
  for_each = local.privileges_map

  db_cluster_id    = alicloud_polardb_cluster.main.id
  account_name     = alicloud_polardb_account.accounts[each.value.account_name].account_name
  account_privilege = each.value.account_privilege
  db_names         = each.value.db_names

  depends_on = [alicloud_polardb_database.databases, alicloud_polardb_account.accounts]
}
