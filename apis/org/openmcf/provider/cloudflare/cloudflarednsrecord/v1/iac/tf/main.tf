# main.tf

# Create the Cloudflare DNS Record
resource "cloudflare_dns_record" "main" {
  zone_id = var.spec.zone_id
  name    = var.spec.name
  type    = local.record_type
  proxied = local.proxied
  ttl     = var.spec.ttl

  # Simple record types carry their value in content; structured types use data.
  content = var.spec.content != "" ? var.spec.content : null
  data    = local.record_data

  # Priority is only used for MX records
  priority = local.requires_priority ? var.spec.priority : null

  # Comment for documentation
  comment = var.spec.comment != "" ? var.spec.comment : null

  # Custom tags
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Record-level settings (only affect proxied records)
  settings = var.spec.settings
}
