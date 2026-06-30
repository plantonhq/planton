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
  description = "CivoVpc specification"
  type = object({
    # The ID of the Civo credential to use for this network (required)
    civo_credential_id = string

    # The name of the network - DNS-friendly label (required)
    network_name = string

    # The Civo region where this network will be created (required)
    # Valid values: LON1, NYC1, FRA1, PHX1, SIN1
    region = string

    # The IPv4 CIDR range for the network (optional, max /24)
    # If omitted, Civo auto-allocates an available range
    # Example: "10.10.1.0/24"
    ip_range_cidr = optional(string, "")

    # Whether this network should be the default for the region (optional)
    # Note: Only one default network per region is allowed
    is_default_for_region = optional(bool, false)

    # An optional description for the network (optional, max 100 characters)
    # Note: The Civo provider doesn't currently expose description field
    description = optional(string, "")
  })
}