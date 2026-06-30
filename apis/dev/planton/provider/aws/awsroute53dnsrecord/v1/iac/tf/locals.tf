locals {
  # Extract zone_id from StringValueOrRef structure
  zone_id = var.spec.zone_id != null ? var.spec.zone_id.value : ""

  # Determine if this is an alias record by checking alias_target.dns_name
  is_alias = (
    var.spec.alias_target != null &&
    var.spec.alias_target.dns_name != null &&
    var.spec.alias_target.dns_name.value != null &&
    var.spec.alias_target.dns_name.value != ""
  )

  # Extract alias target values from StringValueOrRef structure
  alias_dns_name = local.is_alias ? var.spec.alias_target.dns_name.value : null
  alias_zone_id  = local.is_alias && var.spec.alias_target.zone_id != null ? var.spec.alias_target.zone_id.value : null

  # TTL is only applicable for non-alias records
  ttl = local.is_alias ? null : var.spec.ttl

  # Records are only for non-alias records
  records = local.is_alias ? null : var.spec.values

  # Determine routing policy type
  has_weighted    = var.spec.routing_policy != null ? var.spec.routing_policy.weighted != null : false
  has_latency     = var.spec.routing_policy != null ? var.spec.routing_policy.latency != null : false
  has_failover    = var.spec.routing_policy != null ? var.spec.routing_policy.failover != null : false
  has_geolocation = var.spec.routing_policy != null ? var.spec.routing_policy.geolocation != null : false
}
