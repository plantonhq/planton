variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "StripeProduct specification"
  type = object({
    name        = string
    description = optional(string, "")
    active      = optional(bool)
    type        = optional(string, "")
    images      = optional(list(string), [])
    marketing_features = optional(list(object({
      name = string
    })), [])
    package_dimensions = optional(object({
      height = optional(number, 0)
      length = optional(number, 0)
      weight = optional(number, 0)
      width  = optional(number, 0)
    }))
    shippable            = optional(bool)
    statement_descriptor = optional(string, "")
    tax_code             = optional(string, "")
    unit_label           = optional(string, "")
    url                  = optional(string, "")
    metadata             = optional(map(string), {})
    features             = optional(list(string), [])
  })
}
