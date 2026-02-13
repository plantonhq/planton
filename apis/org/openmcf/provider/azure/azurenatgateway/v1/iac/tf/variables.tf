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
  description = "Azure NAT Gateway specification"
  type = object({
    # Reference to the subnet to attach this NAT Gateway to
    subnet_id = string

    # Idle timeout in minutes for TCP connections (4-120)
    idle_timeout_minutes = optional(number, 4)

    # Optional prefix length for Public IP Prefix (28-31)
    # If set, creates a Public IP Prefix instead of individual IP
    public_ip_prefix_length = optional(number)

    # Optional tags to assign to the NAT Gateway resource
    tags = optional(map(string), {})

    # The Azure region where the NAT Gateway will be deployed
    region = string

    # The Azure Resource Group where the NAT Gateway will be created
    resource_group = string
  })

  validation {
    condition     = var.spec.idle_timeout_minutes >= 4 && var.spec.idle_timeout_minutes <= 120
    error_message = "idle_timeout_minutes must be between 4 and 120 (inclusive)."
  }

  validation {
    condition = (
      var.spec.public_ip_prefix_length == null ||
      (var.spec.public_ip_prefix_length >= 28 && var.spec.public_ip_prefix_length <= 31)
    )
    error_message = "public_ip_prefix_length, if specified, must be between 28 and 31 (inclusive)."
  }
}
