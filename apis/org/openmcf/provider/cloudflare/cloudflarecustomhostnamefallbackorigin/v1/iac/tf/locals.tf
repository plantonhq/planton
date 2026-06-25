locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-custom-hostname-fallback-origin")

  # zone_id and origin are StringValueOrRef flattened to plain strings.
  zone_id = try(var.spec.zone_id, "")
  origin  = try(var.spec.origin, "")
}
