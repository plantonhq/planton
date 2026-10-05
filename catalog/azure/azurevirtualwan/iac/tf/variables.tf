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
  description = "AzureVirtualWan specification"
  type = object({
    region                            = string
    resource_group                    = string
    name                              = string
    disable_vpn_encryption            = optional(bool, false)
    allow_branch_to_branch_traffic    = optional(bool)
    office365_local_breakout_category = optional(string)
    type                              = optional(string)
    tags                              = optional(map(string), {})
  })
}