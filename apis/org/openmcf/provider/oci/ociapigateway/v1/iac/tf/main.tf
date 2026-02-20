resource "oci_apigateway_gateway" "this" {
  compartment_id = var.spec.compartment_id.value
  endpoint_type  = lookup(local.endpoint_type_map, var.spec.endpoint_type, null)
  subnet_id      = var.spec.subnet_id.value
  display_name   = local.display_name
  freeform_tags  = local.freeform_tags

  certificate_id = var.spec.certificate_id != "" ? var.spec.certificate_id : null

  network_security_group_ids = length(var.spec.network_security_group_ids) > 0 ? [
    for n in var.spec.network_security_group_ids : n.value
  ] : null
}
