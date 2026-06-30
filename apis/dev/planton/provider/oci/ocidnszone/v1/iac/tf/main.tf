resource "oci_dns_zone" "this" {
  compartment_id = var.spec.compartment_id.value
  name           = var.metadata.name
  zone_type      = local.zone_type_map[var.spec.zone_type]
  freeform_tags  = local.freeform_tags

  scope        = local.scope_value
  view_id      = var.spec.view_id != null ? var.spec.view_id.value : null
  dnssec_state = local.dnssec_state

  dynamic "external_masters" {
    for_each = var.spec.external_masters
    content {
      address     = external_masters.value.address
      port        = external_masters.value.port
      tsig_key_id = external_masters.value.tsig_key_id != "" ? external_masters.value.tsig_key_id : null
    }
  }

  dynamic "external_downstreams" {
    for_each = var.spec.external_downstreams
    content {
      address     = external_downstreams.value.address
      port        = external_downstreams.value.port
      tsig_key_id = external_downstreams.value.tsig_key_id != "" ? external_downstreams.value.tsig_key_id : null
    }
  }
}
