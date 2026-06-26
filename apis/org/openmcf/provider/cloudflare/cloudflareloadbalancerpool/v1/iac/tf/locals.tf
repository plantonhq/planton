locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-load-balancer-pool")

  monitor       = try(var.spec.monitor, "")
  check_regions = try(var.spec.check_regions, [])

  # Build the provider's origin objects, converting host_header into the
  # header.host list and omitting unset optional values (null lets the provider
  # apply its own defaults: weight 1, enabled true, flatten_cname true).
  origins = [
    for o in var.spec.origins : {
      name               = o.name
      address            = o.address
      weight             = o.weight
      enabled            = o.enabled
      flatten_cname      = o.flatten_cname
      port               = try(o.port, 0) > 0 ? o.port : null
      virtual_network_id = try(o.virtual_network_id, "") != "" ? o.virtual_network_id : null
      header             = try(o.host_header, "") != "" ? { host = [o.host_header] } : null
    }
  ]

  load_shedding = try(var.spec.load_shedding, null) == null ? null : {
    default_percent = try(var.spec.load_shedding.default_percent, 0)
    default_policy  = try(var.spec.load_shedding.default_policy, "") != "" ? var.spec.load_shedding.default_policy : null
    session_percent = try(var.spec.load_shedding.session_percent, 0)
    session_policy  = try(var.spec.load_shedding.session_policy, "") != "" ? var.spec.load_shedding.session_policy : null
  }

  origin_steering = (
    try(var.spec.origin_steering, null) == null || try(var.spec.origin_steering.policy, "") == ""
  ) ? null : { policy = var.spec.origin_steering.policy }

  notification_filter = try(var.spec.notification_filter, null) == null ? null : {
    origin = try(var.spec.notification_filter.origin, null) == null ? null : {
      disable = try(var.spec.notification_filter.origin.disable, false)
      healthy = try(var.spec.notification_filter.origin.healthy, null)
    }
    pool = try(var.spec.notification_filter.pool, null) == null ? null : {
      disable = try(var.spec.notification_filter.pool.disable, false)
      healthy = try(var.spec.notification_filter.pool.healthy, null)
    }
  }
}
