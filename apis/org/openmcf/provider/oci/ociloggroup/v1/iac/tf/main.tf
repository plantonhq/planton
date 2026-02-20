resource "oci_logging_log_group" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = var.metadata.name
  description    = var.spec.description != "" ? var.spec.description : null
  freeform_tags  = local.freeform_tags
}
