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
  description = "StripePromotionCode specification"
  type = object({
    coupon           = string
    code             = optional(string, "")
    customer         = optional(string, "")
    customer_account = optional(string, "")
    expires_at       = optional(number)
    max_redemptions  = optional(number)
    restrictions = optional(object({
      first_time_transaction  = optional(bool)
      minimum_amount          = optional(number)
      minimum_amount_currency = optional(string, "")
      currency_options        = optional(map(number), {})
    }))
    active   = optional(bool)
    metadata = optional(map(string), {})
  })
}
