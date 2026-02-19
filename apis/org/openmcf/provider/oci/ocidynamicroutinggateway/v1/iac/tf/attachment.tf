resource "oci_core_drg_attachment" "this" {
  for_each = local.attachments_map

  drg_id        = oci_core_drg.this.id
  display_name  = each.value.display_name
  freeform_tags = local.freeform_tags

  network_details {
    type           = lookup(local.network_type_map, each.value.network_details.type, upper(each.value.network_details.type))
    id             = each.value.network_details.id.value
    route_table_id = each.value.network_details.route_table_id != "" ? each.value.network_details.route_table_id : null
    vcn_route_type = each.value.network_details.vcn_route_type != "" ? lookup(local.vcn_route_type_map, each.value.network_details.vcn_route_type, upper(each.value.network_details.vcn_route_type)) : null
  }

  drg_route_table_id = (
    each.value.drg_route_table_name != ""
    ? oci_core_drg_route_table.this[each.value.drg_route_table_name].id
    : null
  )

  export_drg_route_distribution_id = (
    each.value.export_drg_route_distribution_name != ""
    ? oci_core_drg_route_distribution.this[each.value.export_drg_route_distribution_name].id
    : null
  )

  depends_on = [
    oci_core_drg_route_table.this,
    oci_core_drg_route_distribution.this,
  ]
}
