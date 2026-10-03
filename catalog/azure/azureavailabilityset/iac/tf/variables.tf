variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AzureAvailabilitySet specification"
  type = object({
    resource_group                = string
    name                          = string
    region                        = string
    platform_update_domain_count  = optional(number)
    platform_fault_domain_count   = optional(number)
    managed                       = optional(bool)
    proximity_placement_group_id  = optional(string, "")
    tags                          = optional(map(string), {})
  })
}
