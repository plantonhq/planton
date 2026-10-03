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
  description = "StripeRadarValueList specification"
  type = object({
    alias     = string
    name      = string
    item_type = optional(string, "")
    items     = optional(list(string), [])
    metadata  = optional(map(string), {})
  })
}
