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
  description = "OpenStackFloatingIpSpec defines the configuration for a Neutron floating IP"
  type = object({
    # (Required) The ID of the external network to allocate the floating IP from.
    # Supports StringValueOrRef pattern - use {value: "network-id"} for literal values.
    floating_network_id = object({
      value = string
    })

    # (Optional) The ID of a port to associate the floating IP with.
    # Supports StringValueOrRef pattern - use {value: "port-id"} for literal values.
    port_id = optional(object({
      value = string
    }))

    # (Optional) Fixed IP address on the port to associate with.
    # Only relevant when port_id is set and port has multiple IPs.
    fixed_ip = optional(string, "")

    # (Optional) Subnet within the external network to allocate from.
    subnet_id = optional(string, "")

    # (Optional) Request a specific floating IP address.
    address = optional(string, "")

    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Tags to associate with the floating IP.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
