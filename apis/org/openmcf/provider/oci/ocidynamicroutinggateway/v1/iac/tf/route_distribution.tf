resource "oci_core_drg_route_distribution" "this" {
  for_each = local.route_distributions_map

  drg_id            = oci_core_drg.this.id
  distribution_type = lookup(local.distribution_type_map, each.value.distribution_type, upper(each.value.distribution_type))
  display_name      = each.value.display_name
  freeform_tags     = local.freeform_tags
}

resource "oci_core_drg_route_distribution_statement" "this" {
  for_each = local.distribution_statements_map

  drg_route_distribution_id = oci_core_drg_route_distribution.this[each.value.dist_name].id
  action                    = "ACCEPT"
  priority                  = each.value.priority

  match_criteria {
    match_type      = lookup(local.match_type_map, each.value.match_type, upper(each.value.match_type))
    attachment_type = each.value.attachment_type != "" ? upper(each.value.attachment_type) : null
    drg_attachment_id = each.value.drg_attachment_name != "" ? oci_core_drg_attachment.this[each.value.drg_attachment_name].id : null
  }

  depends_on = [oci_core_drg_attachment.this]
}
