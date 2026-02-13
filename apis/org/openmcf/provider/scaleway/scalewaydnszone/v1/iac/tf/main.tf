# ScalewayDnsZone Terraform Module
#
# Composite resource: creates a Scaleway DNS zone with optional inline
# DNS records.
#
# Resources created:
#   - scaleway_domain_zone (1x) -- the zone itself
#   - scaleway_domain_record (0..Nx) -- one per inline record entry
#
# NOTE: Scaleway DNS zones and records do not support tags. Unlike most
# other Scaleway resources, the DNS API does not accept tags/labels.

# ── DNS Zone ─────────────────────────────────────────────────────────

resource "scaleway_domain_zone" "zone" {
  domain    = local.domain
  subdomain = local.subdomain
}

# ── Inline DNS Records ───────────────────────────────────────────────

resource "scaleway_domain_record" "records" {
  for_each = local.dns_records

  dns_zone = local.zone_name
  name     = each.value.name
  type     = each.value.type
  data     = each.value.data
  ttl      = each.value.ttl

  # Priority is only meaningful for MX and SRV records.
  # Set to null for other types to avoid unnecessary API calls.
  priority = (
    each.value.type == "MX" || each.value.type == "SRV"
    ? each.value.priority
    : null
  )

  depends_on = [scaleway_domain_zone.zone]
}
