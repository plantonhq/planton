locals {
  # ── Resource identity ──────────────────────────────────────────────
  bucket_name = var.metadata.name
  region      = var.spec.region

  # ── Configuration ──────────────────────────────────────────────────
  versioning_enabled  = var.spec.versioning_enabled
  object_lock_enabled = var.spec.object_lock_enabled
  force_destroy       = var.spec.force_destroy

  # ── Lifecycle ──────────────────────────────────────────────────────
  lifecycle_rules     = var.spec.lifecycle_rules
  has_lifecycle_rules = length(local.lifecycle_rules) > 0

  # ── CORS ───────────────────────────────────────────────────────────
  cors_rules     = var.spec.cors_rules
  has_cors_rules = length(local.cors_rules) > 0

  # ── Standard Planton tags ──────────────────────────────────────────
  # NOTE: Object Storage uses key-value map tags (S3-compatible),
  # NOT flat "key=value" string tags like other Scaleway resources.
  standard_tags = merge(
    {
      "planton-ai_resource" = "true"
      "planton-ai_name"     = var.metadata.name
      "planton-ai_kind"     = "ScalewayObjectBucket"
    },
    var.metadata.org != null ? { "planton-ai_org" = var.metadata.org } : {},
    var.metadata.env != null ? { "planton-ai_env" = var.metadata.env } : {},
    var.metadata.id != null ? { "planton-ai_id" = var.metadata.id } : {},
  )
}
