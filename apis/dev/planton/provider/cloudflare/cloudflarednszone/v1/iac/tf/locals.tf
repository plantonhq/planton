# locals.tf

locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "cloudflare_dns_zone"
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null &&
    try(var.metadata.env, "") != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment labels
  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Zone type defaults to "full" (guard the proto's unspecified zero value).
  zone_type = (
    var.spec.type == null || var.spec.type == "" || var.spec.type == "zone_type_unspecified"
    ? "full"
    : var.spec.type
  )

  # Zone mode: drop the proto's unspecified zero value.
  zone_mode = (
    try(var.spec.dns_settings.zone_mode, null) == "zone_mode_unspecified"
    ? null
    : try(var.spec.dns_settings.zone_mode, null)
  )

  has_dns_settings = var.spec.dns_settings != null
  has_dnssec       = var.spec.dnssec != null ? var.spec.dnssec.enabled : false
}

