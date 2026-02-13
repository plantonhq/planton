variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Azure Subnet specification"
  type = object({
    # The Azure Resource Group name
    resource_group = string

    # The Azure Resource Manager ID of the parent VNet
    vnet_id = string

    # The name of the subnet
    name = string

    # The IPv4 CIDR block for the subnet
    address_prefix = string

    # Azure service endpoints to enable
    service_endpoints = optional(list(string))

    # Service delegation
    delegation = optional(object({
      name         = string
      service_name = string
      actions      = optional(list(string))
    }))

    # Private endpoint network policies (Disabled, Enabled, NetworkSecurityGroupEnabled, RouteTableEnabled)
    private_endpoint_network_policies = optional(string, "Disabled")

    # Private Link Service network policies enabled
    private_link_service_network_policies_enabled = optional(bool, true)
  })
}
