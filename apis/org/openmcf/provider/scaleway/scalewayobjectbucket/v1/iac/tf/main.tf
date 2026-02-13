# ScalewayObjectBucket Terraform Module
#
# This module provisions a Scaleway Object Storage bucket with optional
# versioning, lifecycle rules, and CORS configuration.
#
# Object Storage buckets are REGIONAL resources (not zonal).
# Available regions: fr-par, nl-ams, pl-waw
#
# Tags use key-value map format (S3-compatible), unlike other Scaleway
# resources that use flat string tags.

resource "scaleway_object_bucket" "bucket" {
  name   = local.bucket_name
  region = local.region
  tags   = local.standard_tags

  # Object Lock must be enabled at creation time and cannot be removed.
  # Requires versioning to be enabled (enforced by proto CEL validation).
  object_lock_enabled = local.object_lock_enabled

  # Force destroy: when true, deletes all objects before destroying bucket.
  force_destroy = local.force_destroy

  # ── Versioning ─────────────────────────────────────────────────────
  dynamic "versioning" {
    for_each = local.versioning_enabled ? [1] : []
    content {
      enabled = true
    }
  }

  # ── Lifecycle Rules ────────────────────────────────────────────────
  dynamic "lifecycle_rule" {
    for_each = local.lifecycle_rules
    content {
      id      = lifecycle_rule.value.id
      enabled = lifecycle_rule.value.enabled
      prefix  = lifecycle_rule.value.prefix != "" ? lifecycle_rule.value.prefix : null

      # Tag filter (optional).
      tags = length(lifecycle_rule.value.tags) > 0 ? lifecycle_rule.value.tags : null

      # Expiration (optional).
      dynamic "expiration" {
        for_each = lifecycle_rule.value.expiration_days > 0 ? [1] : []
        content {
          days = lifecycle_rule.value.expiration_days
        }
      }

      # Storage class transitions (optional).
      dynamic "transition" {
        for_each = lifecycle_rule.value.transitions
        content {
          days          = transition.value.days
          storage_class = transition.value.storage_class
        }
      }

      # Abort incomplete multipart uploads (optional).
      abort_incomplete_multipart_upload_days = (
        lifecycle_rule.value.abort_incomplete_multipart_upload_days > 0
        ? lifecycle_rule.value.abort_incomplete_multipart_upload_days
        : null
      )
    }
  }

  # ── CORS Rules ─────────────────────────────────────────────────────
  dynamic "cors_rule" {
    for_each = local.cors_rules
    content {
      allowed_methods = cors_rule.value.allowed_methods
      allowed_origins = cors_rule.value.allowed_origins
      allowed_headers = length(cors_rule.value.allowed_headers) > 0 ? cors_rule.value.allowed_headers : null
      expose_headers  = length(cors_rule.value.expose_headers) > 0 ? cors_rule.value.expose_headers : null
      max_age_seconds = cors_rule.value.max_age_seconds > 0 ? cors_rule.value.max_age_seconds : null
    }
  }
}
