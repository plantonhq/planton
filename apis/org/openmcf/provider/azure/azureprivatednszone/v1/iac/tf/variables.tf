variable "metadata" {
  description = "Resource metadata"
  type = object({
    name    = string
    id      = optional(string, "")
    org     = optional(string, "")
    env     = optional(string, "")
    labels  = optional(map(string), {})
    tags    = optional(map(string), {})
    version = optional(string, "")
  })
}

variable "spec" {
  description = "AzurePrivateDnsZone spec"
  type = object({
    resource_group       = string
    name                 = string
    vnet_id              = string
    registration_enabled = optional(bool, false)
  })
}
