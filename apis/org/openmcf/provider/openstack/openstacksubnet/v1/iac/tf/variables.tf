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
  description = "OpenStackSubnetSpec defines the configuration for a Neutron subnet"
  type = object({
    # (Required) The ID of the network to which this subnet belongs.
    # Supports StringValueOrRef pattern - use {value: "network-id"} for literal values.
    network_id = object({
      value = string
    })

    # (Required) IP address range in CIDR notation.
    cidr = string

    # (Optional) IP protocol version: 4 or 6. Default: 4.
    ip_version = optional(number, 4)

    # (Optional) Gateway IP address. Mutually exclusive with no_gateway.
    gateway_ip = optional(string, "")

    # (Optional) Disable gateway. Mutually exclusive with gateway_ip.
    no_gateway = optional(bool, false)

    # (Optional) Enable DHCP on the subnet. Default: true.
    enable_dhcp = optional(bool, true)

    # (Optional) DNS server IP addresses pushed to instances via DHCP.
    dns_nameservers = optional(list(string), [])

    # (Optional) IP allocation sub-ranges within the CIDR.
    allocation_pools = optional(list(object({
      start = string
      end   = string
    })), [])

    # (Optional) Human-readable description.
    description = optional(string, "")

    # (Optional) Tags to associate with the subnet.
    tags = optional(list(string), [])

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
