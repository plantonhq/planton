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
  description = "OpenStackNetworkSpec defines the configuration for a Neutron network"
  type = object({
    # (Optional) Human-readable description of the network.
    description = optional(string, "")

    # (Optional) Administrative state. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) Whether shared across all tenants. Default: false.
    shared = optional(bool, false)

    # (Optional) Whether this is an external/provider network. Default: false.
    external = optional(bool, false)

    # (Optional) Maximum Transmission Unit in bytes. 0 = OpenStack default.
    mtu = optional(number, 0)

    # (Optional) DNS domain for the network. Must end with a dot if set.
    dns_domain = optional(string, "")

    # (Optional) Whether port security is enforced. null = deployment default.
    port_security_enabled = optional(bool)

    # (Optional) Tags to associate with the network.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
