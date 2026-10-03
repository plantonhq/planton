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
  description = "StripePrice specification"
  type = object({
    product             = string
    currency            = string
    unit_amount         = optional(number)
    unit_amount_decimal = optional(string, "")
    billing_scheme      = optional(string, "")
    tiers_mode          = optional(string, "")
    tiers = optional(list(object({
      up_to               = optional(string, "")
      flat_amount         = optional(number)
      flat_amount_decimal = optional(string, "")
      unit_amount         = optional(number)
      unit_amount_decimal = optional(string, "")
    })), [])
    custom_unit_amount = optional(object({
      minimum = optional(number)
      maximum = optional(number)
      preset  = optional(number)
    }))
    transform_quantity = optional(object({
      divide_by = optional(number, 0)
      round     = optional(string, "")
    }))
    recurring = optional(object({
      interval          = optional(string, "")
      interval_count    = optional(number)
      usage_type        = optional(string, "")
      meter             = optional(string, "")
      trial_period_days = optional(number)
    }))
    currency_options = optional(map(object({
      unit_amount         = optional(number)
      unit_amount_decimal = optional(string, "")
      tax_behavior        = optional(string, "")
      tiers = optional(list(object({
        up_to               = optional(string, "")
        flat_amount         = optional(number)
        flat_amount_decimal = optional(string, "")
        unit_amount         = optional(number)
        unit_amount_decimal = optional(string, "")
      })), [])
      custom_unit_amount = optional(object({
        minimum = optional(number)
        maximum = optional(number)
        preset  = optional(number)
      }))
    })), {})
    lookup_key          = optional(string, "")
    transfer_lookup_key = optional(bool, false)
    nickname            = optional(string, "")
    tax_behavior        = optional(string, "")
    active              = optional(bool)
    metadata            = optional(map(string), {})
  })
}
