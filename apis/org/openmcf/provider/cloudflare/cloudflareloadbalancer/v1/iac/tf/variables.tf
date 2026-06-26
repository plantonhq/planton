variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string
    id      = optional(string)
    org     = optional(string)
    env     = optional(string)
    labels  = optional(map(string))
    tags    = optional(list(string))
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareLoadBalancerSpec — a zone-scoped load balancer over account-scoped pools"
  # NOTE: StringValueOrRef fields (zone_id, pool references) flatten to plain
  # strings (and repeated ones to list(string)); enums flatten to their string
  # names; unset proto3 `optional` scalars arrive as null.
  type = object({
    # (Required) DNS hostname for the load balancer.
    hostname = string

    # (Required) Zone that owns the hostname (StringValueOrRef -> string).
    zone_id = optional(string, "")

    # (Required) Ordered default pools and the fallback pool (pool IDs/refs).
    default_pools = list(string)
    fallback_pool = string

    # Enum names; "none"/"off"/"" mean the respective defaults.
    session_affinity = optional(string, "")
    steering_policy  = optional(string, "")

    proxied = optional(bool)
    enabled = optional(bool)

    ttl                  = optional(number, 0)
    session_affinity_ttl = optional(number, 0)

    description = optional(string, "")

    session_affinity_attributes = optional(object({
      drain_duration         = optional(number, 0)
      headers                = optional(list(string), [])
      require_all_headers    = optional(bool, false)
      samesite               = optional(string, "")
      secure                 = optional(string, "")
      zero_downtime_failover = optional(string, "")
    }))

    region_pools = optional(list(object({
      code     = string
      pool_ids = list(string)
    })), [])
    country_pools = optional(list(object({
      code     = string
      pool_ids = list(string)
    })), [])
    pop_pools = optional(list(object({
      code     = string
      pool_ids = list(string)
    })), [])

    adaptive_routing = optional(object({
      failover_across_pools = optional(bool, false)
    }))

    location_strategy = optional(object({
      mode       = optional(string, "")
      prefer_ecs = optional(string, "")
    }))

    random_steering = optional(object({
      default_weight = optional(number, 0)
      pool_weights   = optional(map(number), {})
    }))
  })
}
