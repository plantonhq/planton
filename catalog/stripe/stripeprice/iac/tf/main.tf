# StripePrice Main Resources
#
# stripe_price is how much, how often and in which currencies a product is charged. A price's
# amount can never change in Stripe, so every field that shapes the charge (product, currency,
# unit_amount, billing_scheme, tiers_mode, tiers, custom_unit_amount, transform_quantity,
# recurring) forces a replacement. create_before_destroy makes that replacement create the new
# price before the old one is archived, so a product always has a purchasable price and existing
# subscriptions keep the old one. nickname, lookup_key, metadata, currency_options, tax_behavior
# and active update in place.
#
# transfer_lookup_key is write-only in the provider: it is never stored, so the provider sends it
# on every create and update while the spec sets it. product_data is never set: the product it
# would create is invisible to this module and never archived -- the product is a StripeProduct.
#
# Destroy archives the price (active = false); Stripe keeps it. The provider has no handling for a
# price it cannot read: the next refresh fails, and the recovery is `tofu state rm` followed by an
# apply, which creates a new price.
resource "stripe_price" "this" {
  product             = var.spec.product
  currency            = var.spec.currency
  unit_amount         = local.unit_amount
  unit_amount_decimal = local.unit_amount_decimal
  billing_scheme      = local.billing_scheme
  tiers_mode          = local.tiers_mode
  transform_quantity  = local.transform_quantity
  lookup_key          = local.lookup_key
  transfer_lookup_key = local.transfer_lookup_key
  nickname            = local.nickname
  tax_behavior        = local.tax_behavior
  active              = local.active
  metadata            = local.metadata

  dynamic "tiers" {
    for_each = local.tiers
    content {
      up_to               = tiers.value.up_to
      flat_amount         = tiers.value.flat_amount
      flat_amount_decimal = tiers.value.flat_amount_decimal
      unit_amount         = tiers.value.unit_amount
      unit_amount_decimal = tiers.value.unit_amount_decimal
    }
  }

  # Present means enabled: the spec models the customer-chosen amount as the block's presence.
  dynamic "custom_unit_amount" {
    for_each = local.custom_unit_amount == null ? [] : [local.custom_unit_amount]
    content {
      enabled = true
      minimum = custom_unit_amount.value.minimum
      maximum = custom_unit_amount.value.maximum
      preset  = custom_unit_amount.value.preset
    }
  }

  dynamic "recurring" {
    for_each = local.recurring == null ? [] : [local.recurring]
    content {
      interval          = recurring.value.interval
      interval_count    = recurring.value.interval_count
      usage_type        = recurring.value.usage_type
      meter             = recurring.value.meter
      trial_period_days = recurring.value.trial_period_days
    }
  }

  # The API takes currency_options as a map keyed by currency; the provider models it as a list
  # whose entries carry the currency in key.
  dynamic "currency_options" {
    for_each = local.currency_options
    content {
      key                 = currency_options.value.key
      unit_amount         = currency_options.value.unit_amount
      unit_amount_decimal = currency_options.value.unit_amount_decimal
      tax_behavior        = currency_options.value.tax_behavior
      tiers               = currency_options.value.tiers

      dynamic "custom_unit_amount" {
        for_each = currency_options.value.custom_unit_amount == null ? [] : [currency_options.value.custom_unit_amount]
        content {
          enabled = true
          minimum = custom_unit_amount.value.minimum
          maximum = custom_unit_amount.value.maximum
          preset  = custom_unit_amount.value.preset
        }
      }
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}
