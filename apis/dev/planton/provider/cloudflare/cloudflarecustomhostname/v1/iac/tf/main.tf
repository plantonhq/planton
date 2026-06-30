# A Cloudflare for SaaS custom hostname. The ssl object (with defaults coalesced and
# unset optionals omitted) is assembled in locals so both engines send identical
# values and rely on the provider's defaults for everything left unset.
resource "cloudflare_custom_hostname" "main" {
  zone_id              = local.zone_id
  hostname             = var.spec.hostname
  custom_origin_server = local.custom_origin_server
  custom_origin_sni    = local.custom_origin_sni
  custom_metadata      = local.custom_metadata
  ssl                  = local.ssl
}
