resource "oci_identity_policy" "this" {
  compartment_id = var.spec.compartment_id.value
  name           = local.name
  description    = var.spec.description
  statements     = var.spec.statements
  version_date   = var.spec.version_date != "" ? var.spec.version_date : null
  freeform_tags  = local.freeform_tags
}
