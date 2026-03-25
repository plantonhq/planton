locals {
  zone_id      = try(var.spec.zone_id.value, "")
  account_id   = var.spec.account_id
  ruleset_kind = var.spec.ruleset_kind
  phase        = var.spec.phase
}
