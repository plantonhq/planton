resource "oci_core_network_security_group" "this" {
  compartment_id = var.spec.compartment_id.value
  vcn_id         = var.spec.vcn_id.value
  display_name   = local.display_name
  freeform_tags  = local.freeform_tags
}
