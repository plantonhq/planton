resource "oci_core_drg_route_table" "this" {
  for_each = local.route_tables_map

  drg_id        = oci_core_drg.this.id
  display_name  = each.value.display_name
  freeform_tags = local.freeform_tags

  is_ecmp_enabled = each.value.is_ecmp_enabled ? true : null

  import_drg_route_distribution_id = (
    each.value.import_drg_route_distribution_name != ""
    ? oci_core_drg_route_distribution.this[each.value.import_drg_route_distribution_name].id
    : null
  )
}

resource "oci_core_drg_route_table_route_rule" "this" {
  for_each = local.static_route_rules_map

  drg_route_table_id         = oci_core_drg_route_table.this[each.value.rt_name].id
  destination                = each.value.destination
  destination_type           = "CIDR_BLOCK"
  next_hop_drg_attachment_id = oci_core_drg_attachment.this[each.value.next_hop_attachment_name].id

  depends_on = [oci_core_drg_attachment.this]
}
