# The zone's fallback origin — the default backend all custom hostnames route to.
# One per zone (the zone_id is the resource id).
resource "cloudflare_custom_hostname_fallback_origin" "main" {
  zone_id = local.zone_id
  origin  = local.origin
}
