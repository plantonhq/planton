locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-zero-trust-access-group")

  # Scope: exactly one of account_id or zone_id is set (enforced by the spec).
  account_id = try(var.spec.account_id, "")
  zone_id    = try(var.spec.zone_id, "")

  # Access rules pass straight through to the provider: each element already
  # carries exactly one variant (the proto oneof), and the proto field names match
  # the provider's attribute names 1:1.
  include = var.spec.include
  exclude = try(var.spec.exclude, [])
  require = try(var.spec.require, [])
}
