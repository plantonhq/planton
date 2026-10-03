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
  description = "AzureComputeGallery specification"
  type = object({
    resource_group = string
    name           = string
    region         = string
    description    = optional(string, "")
    sharing = optional(object({
      permission = string
      community_gallery = optional(object({
        eula            = string
        prefix          = string
        publisher_email = string
        publisher_uri   = string
      }))
    }))
    tags = optional(map(string), {})
  })
}
