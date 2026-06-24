locals {
  # Resource naming
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-r2-bucket")

  # Labels/tags
  labels = merge({
    "name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Bucket configuration
  bucket_name = var.spec.bucket_name
  account_id  = var.spec.account_id

  # Location hint. The enum value already matches Cloudflare's expected string;
  # "auto" means "no hint", which the provider expresses by omitting the field.
  location      = coalesce(try(var.spec.location, null), "auto")
  location_hint = local.location == "auto" ? null : local.location

  # Public access via the managed r2.dev URL is handled outside this module.
  public_access = coalesce(try(var.spec.public_access, null), false)

  # Path-style S3 API URL for the bucket.
  bucket_url = "https://${local.account_id}.r2.cloudflarestorage.com/${local.bucket_name}"

  # Custom domain configuration. zone_id is resolved to a plain string before
  # this module runs, so it is read directly.
  custom_domain_enabled = try(var.spec.custom_domain.enabled, false)
  custom_domain_zone_id = try(var.spec.custom_domain.zone_id, "")
  custom_domain_name    = try(var.spec.custom_domain.domain, "")
}

