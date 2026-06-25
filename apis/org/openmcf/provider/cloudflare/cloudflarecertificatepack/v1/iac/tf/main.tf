# An advanced certificate pack for a zone. cloudflare_branding is sent only when
# true (null otherwise) so the provider applies its own default, matching the
# Pulumi module byte-for-byte.
resource "cloudflare_certificate_pack" "main" {
  zone_id               = local.zone_id
  certificate_authority = var.spec.certificate_authority
  type                  = local.cert_type
  validation_method     = var.spec.validation_method
  validity_days         = var.spec.validity_days
  hosts                 = var.spec.hosts
  cloudflare_branding   = try(var.spec.cloudflare_branding, false) ? true : null
}
