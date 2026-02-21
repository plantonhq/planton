resource "oci_identity_dynamic_group" "this" {
  compartment_id = var.spec.compartment_id.value
  name           = local.name
  description    = var.spec.description
  matching_rule  = var.spec.matching_rule
  freeform_tags  = local.freeform_tags
}
