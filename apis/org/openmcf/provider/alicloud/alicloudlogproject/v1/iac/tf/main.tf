resource "alicloud_log_project" "main" {
  project_name      = var.spec.project_name
  description       = var.spec.description
  resource_group_id = var.spec.resource_group_id != "" ? var.spec.resource_group_id : null
  tags              = local.final_tags
}

resource "alicloud_log_store" "stores" {
  for_each = local.log_stores_map

  project_name        = alicloud_log_project.main.project_name
  logstore_name       = each.value.name
  retention_period    = each.value.retention_days
  shard_count         = each.value.shard_count
  auto_split          = each.value.auto_split
  max_split_shard_count = each.value.max_split_shard_count
  append_meta         = each.value.append_meta
}

resource "alicloud_log_store_index" "indexes" {
  for_each = local.log_stores_with_index

  project  = alicloud_log_project.main.project_name
  logstore = alicloud_log_store.stores[each.key].logstore_name

  full_text {
    case_sensitive  = false
    include_chinese = false
    token           = ", '\";=()[]{}?@&<>/:\n\t\r"
  }

  depends_on = [alicloud_log_store.stores]
}
