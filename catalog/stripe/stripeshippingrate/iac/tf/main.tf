# StripeShippingRate Main Resources
#
# stripe_shipping_rate is a shipping option Checkout and payment links offer. display_name,
# tax_code, delivery_estimate and fixed_amount's amount and currency force a replacement, which
# deactivates the old rate and creates a new one with a new id; a StripePaymentLink whose shipping
# options reference it is replaced on the same apply. tax_behavior, fixed_amount.currency_options,
# active and metadata update in place.
#
# type is always fixed_amount: it is the only shipping rate type Stripe offers, so the module sets
# it and the spec requires fixed_amount.
#
# Destroy deactivates the rate (active = false); Stripe keeps it. The provider has no handling for
# a rate it cannot read: the next refresh fails, and the recovery is `tofu state rm` followed by
# an apply.
resource "stripe_shipping_rate" "this" {
  display_name = var.spec.display_name
  type         = "fixed_amount"
  tax_behavior = local.tax_behavior
  tax_code     = local.tax_code
  active       = local.active
  metadata     = local.metadata

  fixed_amount {
    amount   = local.fixed_amount.amount
    currency = local.fixed_amount.currency

    # The API takes currency_options as a map keyed by currency; the provider models it as a list
    # whose entries carry the currency in key.
    dynamic "currency_options" {
      for_each = local.fixed_amount.currency_options
      content {
        key          = currency_options.value.key
        amount       = currency_options.value.amount
        tax_behavior = currency_options.value.tax_behavior
      }
    }
  }

  dynamic "delivery_estimate" {
    for_each = local.delivery_estimate == null ? [] : [local.delivery_estimate]
    content {
      dynamic "minimum" {
        for_each = delivery_estimate.value.minimum == null ? [] : [delivery_estimate.value.minimum]
        content {
          unit  = minimum.value.unit
          value = minimum.value.value
        }
      }
      dynamic "maximum" {
        for_each = delivery_estimate.value.maximum == null ? [] : [delivery_estimate.value.maximum]
        content {
          unit  = maximum.value.unit
          value = maximum.value.value
        }
      }
    }
  }
}
