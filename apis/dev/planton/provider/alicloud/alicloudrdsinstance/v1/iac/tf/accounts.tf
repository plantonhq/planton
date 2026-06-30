resource "alicloud_rds_account" "accounts" {
  for_each = local.accounts_map

  db_instance_id      = alicloud_db_instance.main.id
  account_name        = each.value.account_name
  account_password    = each.value.account_password
  account_type        = each.value.account_type
  account_description = each.value.account_description != "" ? each.value.account_description : null

  depends_on = [alicloud_db_database.databases]
}

resource "alicloud_db_account_privilege" "privileges" {
  for_each = local.privileges_map

  instance_id  = alicloud_db_instance.main.id
  account_name = alicloud_rds_account.accounts[each.value.account_name].account_name
  privilege    = each.value.privilege
  db_names     = each.value.database_names

  depends_on = [alicloud_db_database.databases, alicloud_rds_account.accounts]
}
