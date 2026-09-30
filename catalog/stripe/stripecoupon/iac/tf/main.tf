# StripeCoupon Main Resources
#
# stripe_coupon is a discount. The provider's update sends only name, metadata and
# currency_options; every other field (the percentage or amount, currency, duration, months,
# redemption limit, redeem-by date and applies_to.products) forces a replacement, which deletes
# the old coupon and creates a new one with a new id. There is no create_before_destroy: a
# coupon's replacement changes its id either way, and a StripePromotionCode that references it is
# replaced on the same apply.
#
# Destroy deletes the coupon. Stripe keeps the discount on every customer and subscription that
# already applied it; nobody new can redeem it. The provider has no handling for a coupon it
# cannot read: the next refresh fails, and the recovery is `tofu state rm` followed by an apply.
resource "stripe_coupon" "this" {
  name               = local.name
  percent_off        = local.percent_off
  amount_off         = local.amount_off
  currency           = local.currency
  duration           = local.duration
  duration_in_months = local.duration_in_months
  max_redemptions    = local.max_redemptions
  redeem_by          = local.redeem_by
  metadata           = local.metadata

  dynamic "applies_to" {
    for_each = length(local.applies_to_products) > 0 ? [local.applies_to_products] : []
    content {
      products = applies_to.value
    }
  }

  # The API takes currency_options as a map keyed by currency; the provider models it as a list
  # whose entries carry the currency in key.
  dynamic "currency_options" {
    for_each = local.currency_options
    content {
      key        = currency_options.value.key
      amount_off = currency_options.value.amount_off
    }
  }
}
