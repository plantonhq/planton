resource "oci_identity_compartment" "this" {
  compartment_id = var.spec.compartment_id.value
  name           = local.name
  description    = var.spec.description
  enable_delete  = var.spec.enable_delete
  freeform_tags  = local.freeform_tags
}
