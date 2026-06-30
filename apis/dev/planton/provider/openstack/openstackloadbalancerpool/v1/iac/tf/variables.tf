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
  description = "OpenStackLoadBalancerPoolSpec defines the configuration for an Octavia backend pool"
  type = object({
    # (Required) The listener this pool is the default pool for.
    # Supports StringValueOrRef pattern - use {value: "listener-id"} for literal values.
    listener_id = object({
      value = string
    })

    # (Required) The protocol used by pool members to receive traffic.
    # Valid values: HTTP, HTTPS, TCP, UDP, PROXY
    protocol = string

    # (Required) The load-balancing algorithm.
    # Valid values: ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP, SOURCE_IP_PORT
    lb_method = string

    # (Optional) Session persistence configuration.
    persistence = optional(object({
      type        = string
      cookie_name = optional(string, "")
    }))

    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Administrative state of the pool. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) Tags to associate with the pool.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
