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
  description = "OpenStackLoadBalancerMemberSpec defines the configuration for an Octavia pool member"
  type = object({
    # (Required) The pool to add this member to.
    # Supports StringValueOrRef pattern - use {value: "pool-id"} for literal values.
    pool_id = object({
      value = string
    })

    # (Required) The IP address of the backend server.
    address = string

    # (Required) The port on the backend server that accepts traffic.
    # Must be between 1 and 65535.
    protocol_port = number

    # (Optional) The subnet where the member resides.
    # Supports StringValueOrRef pattern.
    subnet_id = optional(object({
      value = string
    }))

    # (Optional) Weight of this member for weighted load balancing.
    # Valid range: 0-256. Default: 1 (set by Octavia).
    weight = optional(number)

    # (Optional) Administrative state of the member. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) Tags to associate with the member.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
