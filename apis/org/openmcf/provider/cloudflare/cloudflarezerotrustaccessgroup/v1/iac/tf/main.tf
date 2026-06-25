# Cloudflare Zero Trust Access group: a reusable, named bundle of access rules
# (include/exclude/require) referenced by Access policies and other groups.
resource "cloudflare_zero_trust_access_group" "main" {
  account_id = local.account_id != "" ? local.account_id : null
  zone_id    = local.zone_id != "" ? local.zone_id : null

  name = var.spec.name

  include = local.include
  exclude = length(local.exclude) > 0 ? local.exclude : null
  require = length(local.require) > 0 ? local.require : null

  is_default = var.spec.is_default
}
