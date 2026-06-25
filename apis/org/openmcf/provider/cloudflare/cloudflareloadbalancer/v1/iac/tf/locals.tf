locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-load-balancer")

  # Zone (StringValueOrRef flattened to a plain string by the converter).
  zone_id = try(var.spec.zone_id, "")

  # Enums flatten to their string names. Omit the default values so behavior
  # matches the Pulumi module (which omits none/off) and lets the provider decide.
  session_affinity = (
    try(var.spec.session_affinity, "") == "" || var.spec.session_affinity == "none"
  ) ? null : var.spec.session_affinity
  steering_policy = (
    try(var.spec.steering_policy, "") == "" || var.spec.steering_policy == "off"
  ) ? null : var.spec.steering_policy

  # Geo-pool maps: list-of-{code,pool_ids} -> provider's { code => pool_ids }.
  region_pools  = { for e in try(var.spec.region_pools, []) : e.code => e.pool_ids }
  country_pools = { for e in try(var.spec.country_pools, []) : e.code => e.pool_ids }
  pop_pools     = { for e in try(var.spec.pop_pools, []) : e.code => e.pool_ids }

  saa = try(var.spec.session_affinity_attributes, null) == null ? null : {
    drain_duration         = try(var.spec.session_affinity_attributes.drain_duration, 0) > 0 ? var.spec.session_affinity_attributes.drain_duration : null
    headers                = length(try(var.spec.session_affinity_attributes.headers, [])) > 0 ? var.spec.session_affinity_attributes.headers : null
    require_all_headers    = try(var.spec.session_affinity_attributes.require_all_headers, false)
    samesite               = try(var.spec.session_affinity_attributes.samesite, "") != "" ? var.spec.session_affinity_attributes.samesite : null
    secure                 = try(var.spec.session_affinity_attributes.secure, "") != "" ? var.spec.session_affinity_attributes.secure : null
    zero_downtime_failover = try(var.spec.session_affinity_attributes.zero_downtime_failover, "") != "" ? var.spec.session_affinity_attributes.zero_downtime_failover : null
  }

  adaptive_routing = try(var.spec.adaptive_routing, null) == null ? null : {
    failover_across_pools = try(var.spec.adaptive_routing.failover_across_pools, false)
  }

  location_strategy = (
    try(var.spec.location_strategy, null) == null ||
    (try(var.spec.location_strategy.mode, "") == "" && try(var.spec.location_strategy.prefer_ecs, "") == "")
  ) ? null : {
    mode       = try(var.spec.location_strategy.mode, "") != "" ? var.spec.location_strategy.mode : null
    prefer_ecs = try(var.spec.location_strategy.prefer_ecs, "") != "" ? var.spec.location_strategy.prefer_ecs : null
  }

  random_steering = try(var.spec.random_steering, null) == null ? null : {
    default_weight = try(var.spec.random_steering.default_weight, 0) > 0 ? var.spec.random_steering.default_weight : null
    pool_weights   = length(try(var.spec.random_steering.pool_weights, {})) > 0 ? var.spec.random_steering.pool_weights : null
  }
}
