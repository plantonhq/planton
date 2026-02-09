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
  description = "OpenStackFloatingIpAssociateSpec defines the configuration for a floating IP association"
  type = object({
    # (Required) The floating IP address or UUID to associate.
    # Supports StringValueOrRef pattern - use {value: "203.0.113.42"} for literal values.
    floating_ip = object({
      value = string
    })

    # (Required) The ID of the port to associate the floating IP with.
    # Supports StringValueOrRef pattern - use {value: "port-id"} for literal values.
    port_id = object({
      value = string
    })

    # (Optional) Fixed IP on the port to map the floating IP to.
    # Only needed when port has multiple fixed IPs.
    fixed_ip = optional(string, "")

    # (Optional) Override the region from the provider config.
    region = optional(string, "")
  })
}
