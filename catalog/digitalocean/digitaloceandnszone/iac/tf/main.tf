# The DNS zone (a DigitalOcean "domain"). Adding a domain does not require
# owning it; it resolves once the registrar delegates to DigitalOcean's name
# servers. ip_address is a create-only convenience that seeds an initial apex
# A record DigitalOcean never tracks afterwards — prefer declaring records.
resource "digitalocean_domain" "dns_zone" {
  name       = var.spec.domain_name
  ip_address = var.spec.ip_address != "" ? var.spec.ip_address : null

  # ip_address is applied at creation ONLY and never read back by the API,
  # and the provider marks it ForceNew. Left unguarded, adopting an existing
  # zone whose manifest still carries it -- or editing it later -- would plan
  # a destroy-and-recreate of the whole zone, taking every record and the
  # domain's resolution with it. The seed record it created keeps living in
  # the zone regardless, so later changes are ignored here, exactly as the
  # Pulumi module's IgnoreChanges does; the spec field comment tells manifest
  # authors the same.
  lifecycle {
    ignore_changes = [ip_address]
  }
}

# The zone's managed records, one resource per record value.
resource "digitalocean_record" "dns_records" {
  for_each = { for record in local.dns_records : record.key => record }

  domain = digitalocean_domain.dns_zone.id
  type   = each.value.type
  name   = each.value.name
  value  = each.value.value
  ttl    = each.value.ttl

  priority = each.value.priority
  weight   = each.value.weight
  port     = each.value.port
  flags    = each.value.flags
  tag      = each.value.tag
}
