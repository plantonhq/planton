# StripeCoupon Main Resources
#
# stripe_coupon is a discount. Only name and metadata update in place; every other field (the
# percentage or amount, currency, duration, months, redemption limit, redeem-by date, the products
# and the amounts in other currencies) forces a replacement, which deletes the old coupon and
# creates a new one with a new id. There is no create_before_destroy: a
# coupon's replacement changes its id either way, and a StripePromotionCode that references it is
# replaced on the same apply.
#
# Destroy deletes the coupon. Stripe keeps the discount on every customer and subscription that
# already applied it; nobody new can redeem it. The provider has no handling for a coupon it
# cannot read: the next refresh fails, and the recovery is `tofu state rm` followed by an apply.

# Stripe leaves applies_to and currency_options out of a coupon it returns unless the read
# expands them, and the provider never expands (verified live), so a coupon imported blind holds
# neither in state. The provider would replace it for the products and re-send the amounts for the
# other currencies, which Stripe refuses: an existing currency's amount can't be changed, only new
# currencies added (verified live). The coupon therefore ignores both after create, and this
# tracker carries them instead: a change to either replaces the coupon (replace_triggered_by), the
# one change Stripe accepts for an existing currency. The tracker holds nothing in Stripe and is
# never imported; left out of an import, it is created on the first apply and the coupon is
# untouched.
resource "terraform_data" "replace_triggers" {
  input = {
    applies_to_products = local.applies_to_products
    currency_options    = local.currency_options
  }
}

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

  lifecycle {
    ignore_changes       = [applies_to, currency_options]
    replace_triggered_by = [terraform_data.replace_triggers]
  }
}
