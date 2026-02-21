resource "hcloud_zone" "this" {
  name              = var.spec.domain_name
  mode              = var.spec.mode
  labels            = local.standard_labels
  delete_protection = var.spec.delete_protection != null ? var.spec.delete_protection : false

  ttl = var.spec.ttl

  primary_nameservers = var.spec.primary_nameservers != null ? [
    for ns in var.spec.primary_nameservers : {
      address        = ns.address
      port           = ns.port
      tsig_algorithm = ns.tsig_algorithm
      tsig_key       = ns.tsig_key
    }
  ] : null
}

resource "hcloud_zone_rrset" "this" {
  for_each = local.record_sets

  zone = hcloud_zone.this.id
  name = each.value.name
  type = each.value.type
  ttl  = each.value.ttl

  records = [
    for rec in each.value.records : {
      value   = rec.value
      comment = rec.comment
    }
  ]
}
