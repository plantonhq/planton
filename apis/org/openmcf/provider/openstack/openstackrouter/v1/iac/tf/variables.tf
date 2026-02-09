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
  description = "OpenStackRouterSpec defines the configuration for a Neutron router"
  type = object({
    # (Optional) The ID of the external network for the router's gateway.
    # Supports StringValueOrRef pattern - use {value: "network-id"} for literal values.
    external_network_id = optional(object({
      value = string
    }))

    # (Optional) Administrative state of the router. Default: true.
    admin_state_up = optional(bool, true)

    # (Optional) Enable SNAT on the external gateway. Requires external_network_id.
    enable_snat = optional(bool)

    # (Optional) Enable Distributed Virtual Router (DVR) mode. Create-time only.
    distributed = optional(bool)

    # (Optional) Fixed IP addresses on the external network.
    external_fixed_ips = optional(list(object({
      subnet_id  = optional(string, "")
      ip_address = optional(string, "")
    })), [])

    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Tags to associate with the router.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
