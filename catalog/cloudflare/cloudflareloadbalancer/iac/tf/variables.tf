variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "CloudflareLoadBalancer specification"
  type = object({
    hostname             = string
    zone_id              = string
    proxied              = optional(bool)
    session_affinity     = optional(string, "")
    steering_policy      = optional(string, "")
    default_pools        = list(string)
    fallback_pool        = string
    description          = optional(string, "")
    enabled              = optional(bool)
    ttl                  = optional(number, 0)
    session_affinity_ttl = optional(number, 0)
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
    rules = optional(list(object({
      name       = optional(string, "")
      condition  = optional(string, "")
      priority   = optional(number)
      disabled   = optional(bool, false)
      terminates = optional(bool, false)
      fixed_response = optional(object({
        content_type = optional(string, "")
        location     = optional(string, "")
        message_body = optional(string, "")
        status_code  = optional(number, 0)
      }))
      overrides = optional(object({
        adaptive_routing = optional(object({
          failover_across_pools = optional(bool, false)
        }))
        country_pools = optional(list(object({
          code     = string
          pool_ids = list(string)
        })), [])
        default_pools = optional(list(string), [])
        fallback_pool = optional(string, "")
        location_strategy = optional(object({
          mode       = optional(string, "")
          prefer_ecs = optional(string, "")
        }))
        pop_pools = optional(list(object({
          code     = string
          pool_ids = list(string)
        })), [])
        random_steering = optional(object({
          default_weight = optional(number, 0)
          pool_weights   = optional(map(number), {})
        }))
        region_pools = optional(list(object({
          code     = string
          pool_ids = list(string)
        })), [])
        session_affinity = optional(string)
        session_affinity_attributes = optional(object({
          drain_duration         = optional(number, 0)
          headers                = optional(list(string), [])
          require_all_headers    = optional(bool, false)
          samesite               = optional(string, "")
          secure                 = optional(string, "")
          zero_downtime_failover = optional(string, "")
        }))
        session_affinity_ttl = optional(number, 0)
        steering_policy      = optional(string)
        ttl                  = optional(number, 0)
      }))
    })), [])
    networks = optional(list(string), [])
  })
}
