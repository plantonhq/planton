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
  description = "CloudflareLoadBalancerPoolSpec defines an account-scoped pool of origins"
  # NOTE: StringValueOrRef fields (origin address, monitor) flatten to plain
  # strings, and check_regions enums flatten to their string names, via the
  # proto->tfvars converter. Unset proto3 `optional` scalars arrive as null.
  type = object({
    # (Required) The Cloudflare account ID that owns this pool.
    account_id = string

    # (Required) Short pool name (tag).
    name = string

    # (Required) Origin servers in the pool.
    origins = list(object({
      name               = string
      address            = string
      weight             = optional(number)
      enabled            = optional(bool)
      flatten_cname      = optional(bool)
      port               = optional(number, 0)
      host_header        = optional(string, "")
      virtual_network_id = optional(string, "")
    }))

    # (Optional) Monitor ID (or reference) that health-checks the origins.
    monitor = optional(string, "")

    # (Optional) Regions to run health checks from (empty = everywhere).
    check_regions = optional(list(string), [])

    description     = optional(string, "")
    enabled         = optional(bool)
    minimum_origins = optional(number, 0)
    latitude        = optional(number)
    longitude       = optional(number)

    load_shedding = optional(object({
      default_percent = optional(number, 0)
      default_policy  = optional(string, "")
      session_percent = optional(number, 0)
      session_policy  = optional(string, "")
    }))

    origin_steering = optional(object({
      policy = optional(string, "")
    }))

    notification_filter = optional(object({
      origin = optional(object({
        disable = optional(bool, false)
        healthy = optional(bool)
      }))
      pool = optional(object({
        disable = optional(bool, false)
        healthy = optional(bool)
      }))
    }))
  })
}
