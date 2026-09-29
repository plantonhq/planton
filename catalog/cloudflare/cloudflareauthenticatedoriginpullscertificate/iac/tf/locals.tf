locals {
  # The scope decides which of Cloudflare's two upload surfaces receives the
  # certificate. Exactly one of the two resources below is created.
  is_zone_scope = coalesce(var.spec.scope, "zone") == "zone"
}
