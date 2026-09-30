# Local values for the StripeCoupon module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. duration arrives as its value name, which is Stripe's
# own string (once, repeating, forever). applies_to_products arrives as resolved product ids: a
# reference to a StripeProduct is resolved before this module runs.
locals {
  name               = try(var.spec.name, "") != "" ? var.spec.name : null
  percent_off        = try(var.spec.percent_off, null)
  amount_off         = try(var.spec.amount_off, null)
  currency           = try(var.spec.currency, "") != "" ? var.spec.currency : null
  duration           = try(var.spec.duration, "") != "" ? var.spec.duration : null
  duration_in_months = try(var.spec.duration_in_months, null)
  max_redemptions    = try(var.spec.max_redemptions, null)
  redeem_by          = try(var.spec.redeem_by, null)
  metadata           = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null

  applies_to_products = try(var.spec.applies_to_products, [])

  # Sorted by currency so the provider's list never reorders between plans.
  currency_options = [
    for currency in sort(keys(try(var.spec.currency_options, {}))) : {
      key        = currency
      amount_off = var.spec.currency_options[currency]
    }
  ]
}
