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
  description = "Azure VPC specification"
  type = object({
    # Azure region for deployment
    region = string

    # The Azure Resource Group where the VPC will be created
    resource_group = string

    # The CIDR block defining the address space for the Virtual Network
    address_space_cidr = string

    # The CIDR block for the primary subnet for nodes
    nodes_subnet_cidr = string

    # Toggle to enable a NAT Gateway for the nodes subnet
    is_nat_gateway_enabled = optional(bool, false)

    # List of Azure Private DNS zone resource IDs to link to this VNet
    dns_private_zone_links = optional(list(string), [])

    # Arbitrary tags to apply to the Virtual Network
    tags = optional(map(string), {})
  })

  validation {
    condition     = can(cidrhost(var.spec.address_space_cidr, 0))
    error_message = "address_space_cidr must be a valid CIDR block."
  }

  validation {
    condition     = can(cidrhost(var.spec.nodes_subnet_cidr, 0))
    error_message = "nodes_subnet_cidr must be a valid CIDR block."
  }
}
