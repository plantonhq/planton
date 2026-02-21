resource "oci_core_subnet" "this" {
  compartment_id = var.spec.compartment_id.value
  vcn_id         = var.spec.vcn_id.value
  cidr_block     = var.spec.cidr_block
  display_name   = local.display_name
  freeform_tags  = local.freeform_tags

  dns_label               = var.spec.dns_label != "" ? var.spec.dns_label : null
  availability_domain     = var.spec.availability_domain != "" ? var.spec.availability_domain : null
  prohibit_public_ip_on_vnic = var.spec.prohibit_public_ip_on_vnic
  prohibit_internet_ingress  = var.spec.prohibit_internet_ingress
  ipv6cidr_block          = var.spec.ipv6_cidr_block != "" ? var.spec.ipv6_cidr_block : null

  dhcp_options_id = var.spec.dhcp_options_id != null ? var.spec.dhcp_options_id.value : null

  route_table_id = (
    var.spec.route_table_id != null
    ? var.spec.route_table_id.value
    : length(var.spec.route_rules) > 0
      ? oci_core_route_table.this[0].id
      : null
  )

  security_list_ids = length(var.spec.security_list_ids) > 0 ? [
    for sl in var.spec.security_list_ids : sl.value
  ] : null
}
