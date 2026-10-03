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
  description = "AzureFabricCapacity specification"
  type = object({
    resource_group         = string
    name                   = string
    region                 = string
    sku_name               = string
    administration_members = list(string)
    tags                   = optional(map(string), {})
  })
}
