resource "oci_core_vcn" "this" {
  compartment_id = var.spec.compartment_id.value
  cidr_blocks    = var.spec.cidr_blocks
  display_name   = local.display_name
  dns_label      = var.spec.dns_label != "" ? var.spec.dns_label : null
  is_ipv6enabled = var.spec.is_ipv6_enabled
  freeform_tags  = local.freeform_tags
}
