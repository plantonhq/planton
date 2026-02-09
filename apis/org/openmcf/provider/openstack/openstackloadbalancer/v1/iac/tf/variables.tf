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
  description = "OpenStackLoadBalancerSpec defines the configuration for an Octavia load balancer"
  type = object({
    # (Required) The subnet on which to allocate the VIP address.
    # Supports StringValueOrRef pattern - use {value: "subnet-id"} for literal values.
    vip_subnet_id = object({
      value = string
    })

    # (Optional) A specific IP address to request for the VIP.
    vip_address = optional(string, "")

    # (Optional) Human-readable description of the load balancer.
    description = optional(string, "")

    # (Optional) Administrative state of the load balancer. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) The ID of an Octavia flavor to use for the load balancer.
    flavor_id = optional(string, "")

    # (Optional) Tags applied to the load balancer in OpenStack.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
