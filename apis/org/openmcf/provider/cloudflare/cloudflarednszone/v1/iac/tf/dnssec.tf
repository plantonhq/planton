# dnssec.tf

# DNSSEC for the zone, provisioned only when the spec enables it. The DS material
# Cloudflare computes is surfaced through outputs for entry at the registrar.
resource "cloudflare_zone_dnssec" "main" {
  count   = local.has_dnssec ? 1 : 0
  zone_id = cloudflare_zone.main.id
  status  = "active"

  dnssec_multi_signer = var.spec.dnssec.multi_signer
  dnssec_presigned    = var.spec.dnssec.presigned
  dnssec_use_nsec3    = var.spec.dnssec.use_nsec3
}
