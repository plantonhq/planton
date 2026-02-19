resource "oci_core_drg" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = local.display_name
  freeform_tags  = local.freeform_tags
}
