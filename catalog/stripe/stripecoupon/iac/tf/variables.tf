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
  description = "StripeCoupon specification"
  type = object({
    name                = optional(string, "")
    percent_off         = optional(number)
    amount_off          = optional(number)
    currency            = optional(string, "")
    currency_options    = optional(map(number), {})
    duration            = optional(string, "")
    duration_in_months  = optional(number)
    max_redemptions     = optional(number)
    redeem_by           = optional(number)
    applies_to_products = optional(list(string), [])
    metadata            = optional(map(string), {})
  })
}
