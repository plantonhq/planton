resource "oci_nosql_index" "this" {
  for_each = { for idx in var.spec.indexes : idx.name => idx }

  table_name_or_id = oci_nosql_table.this.id
  name             = each.value.name

  dynamic "keys" {
    for_each = each.value.keys
    content {
      column_name     = keys.value.column_name
      json_field_type = keys.value.json_field_type != "" ? keys.value.json_field_type : null
      json_path       = keys.value.json_path != "" ? keys.value.json_path : null
    }
  }
}
