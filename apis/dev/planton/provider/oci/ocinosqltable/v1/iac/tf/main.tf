resource "oci_nosql_table" "this" {
  compartment_id    = var.spec.compartment_id.value
  name              = var.spec.name
  ddl_statement     = var.spec.ddl_statement
  is_auto_reclaimable = var.spec.is_auto_reclaimable
  freeform_tags     = local.freeform_tags

  table_limits {
    capacity_mode    = var.spec.table_limits.capacity_mode != "" ? lookup(local.capacity_mode_map, var.spec.table_limits.capacity_mode, null) : null
    max_read_units   = var.spec.table_limits.max_read_units
    max_write_units  = var.spec.table_limits.max_write_units
    max_storage_in_gbs = var.spec.table_limits.max_storage_in_gbs
  }
}
