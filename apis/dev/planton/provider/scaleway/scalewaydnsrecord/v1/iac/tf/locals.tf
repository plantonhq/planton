locals {
  # ── Record configuration ──────────────────────────────────────────────

  # Extract zone name from StringValueOrRef (use resolved value).
  zone_name = var.spec.zone_name.value

  # Record fields.
  name = var.spec.name
  type = var.spec.type
  data = var.spec.data.value

  ttl             = coalesce(var.spec.ttl, 3600)
  priority        = coalesce(var.spec.priority, 0)
  keep_empty_zone = coalesce(var.spec.keep_empty_zone, true)

  # NOTE: Scaleway DNS records do not support tags.
  # Unlike most other Scaleway resources, the DNS API does not accept
  # tags/labels. Standard Planton metadata tags are not applied.
}
