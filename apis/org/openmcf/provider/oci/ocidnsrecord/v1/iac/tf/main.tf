resource "oci_dns_rrset" "this" {
  zone_name_or_id = var.spec.zone_name_or_id.value
  domain          = var.spec.domain
  rtype           = var.spec.rtype
  view_id         = var.spec.view_id != null ? var.spec.view_id.value : null

  dynamic "items" {
    for_each = var.spec.items
    content {
      domain = var.spec.domain
      rtype  = var.spec.rtype
      rdata  = items.value.rdata
      ttl    = items.value.ttl
    }
  }
}
