# Local values for the StripeShippingRate module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. Enums arrive as their value names, which are Stripe's
# own strings (business_day, exclusive).
locals {
  tax_behavior = try(var.spec.tax_behavior, "") != "" ? var.spec.tax_behavior : null
  tax_code     = try(var.spec.tax_code, "") != "" ? var.spec.tax_code : null
  metadata     = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  # The spec's default is active; the provider needs the value only to deactivate.
  active = try(var.spec.active, null) == null ? true : var.spec.active

  # A free rate's amount is 0, which the manifest may leave out; the provider requires it.
  fixed_amount = {
    amount   = try(var.spec.fixed_amount.amount, 0)
    currency = var.spec.fixed_amount.currency
    # Sorted by currency so the provider's list never reorders between plans.
    currency_options = [
      for currency in sort(keys(try(var.spec.fixed_amount.currency_options, {}))) : {
        key          = currency
        amount       = try(var.spec.fixed_amount.currency_options[currency].amount, 0)
        tax_behavior = try(var.spec.fixed_amount.currency_options[currency].tax_behavior, "") != "" ? var.spec.fixed_amount.currency_options[currency].tax_behavior : null
      }
    ]
  }

  delivery_estimate = try(var.spec.delivery_estimate, null) == null ? null : {
    minimum = try(var.spec.delivery_estimate.minimum, null)
    maximum = try(var.spec.delivery_estimate.maximum, null)
  }
}
