resource "oci_load_balancer_load_balancer" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = local.display_name
  shape          = var.spec.shape
  subnet_ids     = [for s in var.spec.subnet_ids : s.value]
  freeform_tags  = local.freeform_tags
  is_private     = var.spec.is_private

  is_delete_protection_enabled = var.spec.is_delete_protection_enabled ? true : null
  ip_mode                      = var.spec.ip_mode != "" ? var.spec.ip_mode : null
  is_request_id_enabled        = var.spec.is_request_id_enabled ? true : null
  request_id_header            = var.spec.request_id_header != "" ? var.spec.request_id_header : null

  network_security_group_ids = length(var.spec.network_security_group_ids) > 0 ? [
    for n in var.spec.network_security_group_ids : n.value
  ] : null

  dynamic "shape_details" {
    for_each = var.spec.shape_details != null ? [var.spec.shape_details] : []
    content {
      minimum_bandwidth_in_mbps = shape_details.value.minimum_bandwidth_in_mbps
      maximum_bandwidth_in_mbps = shape_details.value.maximum_bandwidth_in_mbps
    }
  }

  dynamic "reserved_ips" {
    for_each = var.spec.reserved_ips
    content {
      id = reserved_ips.value.id
    }
  }
}
