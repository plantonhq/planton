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
  description = "OpenStackNetworkPortSpec defines the configuration for a Neutron port"
  type = object({
    # (Required) The ID of the network to create this port on.
    # Supports StringValueOrRef pattern - use {value: "network-id"} for literal values.
    network_id = object({
      value = string
    })

    # (Optional) IP address allocations for the port.
    # Each entry references a subnet (via StringValueOrRef) and optionally a specific IP.
    fixed_ips = optional(list(object({
      subnet_id = optional(object({
        value = string
      }))
      ip_address = optional(string, "")
    })), [])

    # (Optional) Security group IDs to apply to this port.
    # Supports repeated StringValueOrRef pattern - each entry uses {value: "sg-id"}.
    security_group_ids = optional(list(object({
      value = string
    })), [])

    # (Optional) Explicitly remove all security groups including the default.
    no_security_groups = optional(bool, false)

    # (Optional) Administrative state of the port. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) Specific MAC address for the port. ForceNew.
    mac_address = optional(string, "")

    # (Optional) Whether port security is enforced. Inherits from network if omitted.
    port_security_enabled = optional(bool, null)

    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Tags to associate with the port.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
