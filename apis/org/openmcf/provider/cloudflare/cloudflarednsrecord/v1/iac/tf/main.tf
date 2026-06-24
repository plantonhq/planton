# main.tf

# Create the Cloudflare DNS Record
resource "cloudflare_dns_record" "main" {
  zone_id = var.spec.zone_id
  name    = var.spec.name
  type    = local.record_type
  content = var.spec.value
  proxied = local.proxied
  ttl     = var.spec.ttl

  # Priority is only used for MX and SRV records
  priority = local.requires_priority ? var.spec.priority : null

  # Comment for documentation
  comment = var.spec.comment != "" ? var.spec.comment : null
}
