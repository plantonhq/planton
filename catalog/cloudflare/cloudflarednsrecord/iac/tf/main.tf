# main.tf

# Create the Cloudflare DNS Record
resource "cloudflare_dns_record" "main" {
  zone_id = var.spec.zone_id
  name    = var.spec.name
  type    = local.record_type
  proxied = local.proxied
  ttl     = var.spec.ttl

  # Simple record types carry their value in content (as Cloudflare stores it,
  # see locals.tf); structured types use data.
  content = local.content
  data    = local.record_data

  # Top-level priority mirrors the API contract for MX/SRV/URI (see locals.tf).
  priority = local.top_priority

  # Comment for documentation
  comment = var.spec.comment != "" ? var.spec.comment : null

  # Custom tags
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Record-level settings (only affect proxied records)
  settings = var.spec.settings

  # Restrict the record to Cloudflare internal (private) routing when set.
  private_routing = var.spec.private_routing ? true : null
}
