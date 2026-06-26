# records.tf

# Create DNS records within the zone
# Include index to ensure uniqueness when multiple records have same name and type
resource "cloudflare_dns_record" "records" {
  for_each = { for idx, record in var.spec.records : "${record.name}-${record.type}-${idx}" => record }

  zone_id = cloudflare_zone.main.id
  name    = each.value.name
  type    = each.value.type
  content = each.value.content
  ttl     = each.value.ttl

  # proxied is only applicable to A, AAAA, and CNAME records
  proxied = contains(["A", "AAAA", "CNAME"], each.value.type) ? each.value.proxied : false

  # priority is only used for MX and SRV records
  priority = contains(["MX", "SRV"], each.value.type) ? each.value.priority : null

  # comment for the DNS record
  comment = each.value.comment != "" ? each.value.comment : null

  depends_on = [cloudflare_zone.main]
}
