# Cloudflare R2 Bucket
# R2 is Cloudflare's S3-compatible object storage with zero egress fees.
# The location hint is omitted when "auto" so Cloudflare selects the region.
resource "cloudflare_r2_bucket" "main" {
  account_id = local.account_id
  name       = local.bucket_name
  location   = local.location_hint
}

# Public access via the managed r2.dev URL is configured outside this module
# (it has its own lifecycle); custom domains are the production-grade path below.

# Custom domain for the R2 bucket. Created only when custom_domain.enabled is true.
resource "cloudflare_r2_custom_domain" "main" {
  count = local.custom_domain_enabled ? 1 : 0

  account_id  = local.account_id
  bucket_name = cloudflare_r2_bucket.main.name
  zone_id     = local.custom_domain_zone_id
  domain      = local.custom_domain_name
  enabled     = true
}
