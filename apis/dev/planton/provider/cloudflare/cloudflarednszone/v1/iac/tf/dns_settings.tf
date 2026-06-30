# dns_settings.tf

# Zone-wide DNS settings, provisioned only when the spec supplies a dns_settings block.
resource "cloudflare_zone_dns_settings" "main" {
  count   = local.has_dns_settings ? 1 : 0
  zone_id = cloudflare_zone.main.id

  flatten_all_cnames  = var.spec.dns_settings.flatten_all_cnames
  foundation_dns      = var.spec.dns_settings.foundation_dns
  multi_provider      = var.spec.dns_settings.multi_provider
  secondary_overrides = var.spec.dns_settings.secondary_overrides
  ns_ttl              = var.spec.dns_settings.ns_ttl
  zone_mode           = local.zone_mode

  soa          = try(var.spec.dns_settings.soa, null)
  nameservers  = try(var.spec.dns_settings.nameservers, null)
  internal_dns = try(var.spec.dns_settings.internal_dns, null)
}
