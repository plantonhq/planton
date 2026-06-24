# Cloudflare R2 Bucket
# R2 is Cloudflare's S3-compatible object storage with zero egress fees.
# The location hint is omitted when "auto" so Cloudflare selects the region;
# jurisdiction and storage_class fall back to the provider defaults when null.
resource "cloudflare_r2_bucket" "main" {
  account_id    = local.account_id
  name          = local.bucket_name
  location      = local.location_hint
  jurisdiction  = local.jurisdiction
  storage_class = local.storage_class
}

# Managed public access over the Cloudflare r2.dev domain. Created only when
# public_access is true. The r2.dev domain is rate-limited (development-grade);
# custom domains below are the production path.
resource "cloudflare_r2_managed_domain" "main" {
  count = local.public_access ? 1 : 0

  account_id   = local.account_id
  bucket_name  = cloudflare_r2_bucket.main.name
  jurisdiction = local.jurisdiction
  enabled      = true
}

# Custom domains serving the bucket over your own hostnames. One resource per
# enabled custom domain.
resource "cloudflare_r2_custom_domain" "main" {
  for_each = local.custom_domains_enabled

  account_id   = local.account_id
  bucket_name  = cloudflare_r2_bucket.main.name
  jurisdiction = local.jurisdiction
  zone_id      = each.value.zone_id
  domain       = each.value.domain
  enabled      = true
  min_tls      = try(each.value.min_tls, null)
  ciphers      = try(each.value.ciphers, null)
}

# CORS configuration. Created only when at least one rule is provided.
resource "cloudflare_r2_bucket_cors" "main" {
  count = length(local.cors_rules) > 0 ? 1 : 0

  account_id   = local.account_id
  bucket_name  = cloudflare_r2_bucket.main.name
  jurisdiction = local.jurisdiction

  rules = [for r in local.cors_rules : {
    allowed = {
      methods = r.allowed.methods
      origins = r.allowed.origins
      headers = try(r.allowed.headers, null)
    }
    id              = try(r.id, "") != "" ? r.id : null
    expose_headers  = try(r.expose_headers, null)
    max_age_seconds = try(r.max_age_seconds, null)
  }]
}

# Object lifecycle configuration. Created only when at least one rule is provided.
# The abort-multipart transition is always an "Age" condition; storage-class
# transitions always target Infrequent Access (the only supported target class).
resource "cloudflare_r2_bucket_lifecycle" "main" {
  count = length(local.lifecycle_rules) > 0 ? 1 : 0

  account_id   = local.account_id
  bucket_name  = cloudflare_r2_bucket.main.name
  jurisdiction = local.jurisdiction

  rules = [for r in local.lifecycle_rules : {
    id         = r.id
    enabled    = r.enabled
    conditions = { prefix = try(r.conditions.prefix, "") }

    abort_multipart_uploads_transition = try(r.abort_multipart_uploads_transition, null) != null ? {
      condition = {
        max_age = r.abort_multipart_uploads_transition.max_age_seconds
        type    = "Age"
      }
    } : null

    delete_objects_transition = try(r.delete_objects_transition, null) != null ? {
      condition = {
        type    = r.delete_objects_transition.condition.type
        max_age = try(r.delete_objects_transition.condition.max_age_seconds, null)
        date    = try(r.delete_objects_transition.condition.date, null)
      }
    } : null

    storage_class_transitions = [for t in try(r.storage_class_transitions, []) : {
      condition = {
        type    = t.condition.type
        max_age = try(t.condition.max_age_seconds, null)
        date    = try(t.condition.date, null)
      }
      storage_class = "InfrequentAccess"
    }]
  }]
}

# Object lock (retention) configuration. Created only when at least one rule is provided.
resource "cloudflare_r2_bucket_lock" "main" {
  count = length(local.lock_rules) > 0 ? 1 : 0

  account_id   = local.account_id
  bucket_name  = cloudflare_r2_bucket.main.name
  jurisdiction = local.jurisdiction

  rules = [for r in local.lock_rules : {
    id      = r.id
    enabled = r.enabled
    prefix  = try(r.prefix, "") != "" ? r.prefix : null
    condition = {
      type            = r.condition.type
      max_age_seconds = try(r.condition.max_age_seconds, null)
      date            = try(r.condition.date, null)
    }
  }]
}
