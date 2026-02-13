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
  description = "Azure Public IP specification"
  type = object({
    # The Azure region where the Public IP will be created
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The name of the Public IP resource
    name = string

    # An optional DNS label for domain name mapping
    domain_name_label = optional(string)

    # Availability zones for the Public IP
    zones = optional(list(string))

    # Idle timeout in minutes (4-30, default 4)
    idle_timeout_in_minutes = optional(number, 4)
  })
}
