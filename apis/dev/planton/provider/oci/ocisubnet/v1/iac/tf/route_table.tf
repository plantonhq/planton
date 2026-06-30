resource "oci_core_route_table" "this" {
  count = length(var.spec.route_rules) > 0 ? 1 : 0

  compartment_id = var.spec.compartment_id.value
  vcn_id         = var.spec.vcn_id.value
  display_name   = "${local.display_name}-rt"
  freeform_tags  = local.freeform_tags

  dynamic "route_rules" {
    for_each = var.spec.route_rules
    content {
      destination       = route_rules.value.destination
      destination_type  = lookup(local.destination_type_map, route_rules.value.destination_type, "CIDR_BLOCK")
      network_entity_id = route_rules.value.network_entity_id.value
      description       = route_rules.value.description != "" ? route_rules.value.description : null
    }
  }
}
