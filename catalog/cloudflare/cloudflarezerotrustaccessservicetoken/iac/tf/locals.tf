locals {
  # Scope: exactly one of account_id or zone_id is set (enforced by the spec).
  account_id = try(var.spec.account_id, "")
  zone_id    = try(var.spec.zone_id, "")
}
