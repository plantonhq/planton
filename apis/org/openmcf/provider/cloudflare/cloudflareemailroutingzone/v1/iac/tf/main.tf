# Enabling Email Routing on the zone. Creating this resource turns Email Routing
# on and provisions the zone's required MX/SPF/DKIM records automatically.
resource "cloudflare_email_routing_settings" "main" {
  zone_id = var.spec.zone_id
}

# The single per-zone catch-all rule (folded), created only when configured.
resource "cloudflare_email_routing_catch_all" "main" {
  count   = local.catch_all != null ? 1 : 0
  zone_id = var.spec.zone_id
  enabled = try(local.catch_all.enabled, false)

  matchers = [{
    type = "all"
  }]

  actions = [{
    type  = local.catch_all.type
    value = length(local.catch_all_values) > 0 ? local.catch_all_values : null
  }]

  depends_on = [cloudflare_email_routing_settings.main]
}

# Optionally lock the Email Routing DNS records so they cannot be modified
# out-of-band.
resource "cloudflare_email_routing_dns" "main" {
  count   = try(var.spec.lock_dns_records, false) ? 1 : 0
  zone_id = var.spec.zone_id

  depends_on = [cloudflare_email_routing_settings.main]
}
