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
  description = "StripeShippingRate specification"
  type = object({
    display_name = string
    fixed_amount = object({
      amount   = optional(number, 0)
      currency = string
      currency_options = optional(map(object({
        amount       = optional(number, 0)
        tax_behavior = optional(string, "")
      })), {})
    })
    delivery_estimate = optional(object({
      minimum = optional(object({
        unit  = optional(string, "")
        value = optional(number, 0)
      }))
      maximum = optional(object({
        unit  = optional(string, "")
        value = optional(number, 0)
      }))
    }))
    tax_behavior = optional(string, "")
    tax_code     = optional(string, "")
    active       = optional(bool)
    metadata     = optional(map(string), {})
  })
}
