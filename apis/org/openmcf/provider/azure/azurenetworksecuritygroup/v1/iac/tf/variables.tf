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
  description = "Azure Network Security Group specification"
  type = object({
    # The Azure region where the NSG will be created
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The name of the Network Security Group
    name = string

    # Security rules
    security_rules = optional(list(object({
      # Rule name (unique within NSG)
      name = string

      # Optional description (max 140 chars)
      description = optional(string)

      # Priority (100-4096, lower = evaluated first)
      priority = number

      # Direction: "Inbound" or "Outbound"
      direction = string

      # Access decision: "Allow" or "Deny"
      access = string

      # Protocol: "Tcp", "Udp", "Icmp", or "*"
      protocol = string

      # Source port range (default "*")
      source_port_range = optional(string, "*")

      # Destination port range (required)
      destination_port_range = string

      # Source address prefix (default "*")
      source_address_prefix = optional(string, "*")

      # Destination address prefix (default "*")
      destination_address_prefix = optional(string, "*")

      # Source address prefixes (overrides singular if non-empty)
      source_address_prefixes = optional(list(string))

      # Destination address prefixes (overrides singular if non-empty)
      destination_address_prefixes = optional(list(string))
    })), [])
  })
}
