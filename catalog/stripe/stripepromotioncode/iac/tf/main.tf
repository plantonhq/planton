# StripePromotionCode Main Resources
#
# stripe_promotion_code is a customer-facing code on a coupon. The provider's update sends only
# active and metadata; every other field (the coupon, code, customer, expiry, redemption limit and
# restrictions) forces a replacement, which deactivates the old code and creates a new one. A
# replacement that keeps the same code works because Stripe requires a code to be unique only
# among active codes, and the old one is deactivated first.
#
# promotion.type is always coupon: it is the only promotion Stripe offers, so the module sets it
# and the spec names only the coupon.
#
# Destroy deactivates the code (active = false); Stripe keeps it. The provider has no handling for
# a code it cannot read: the next refresh fails, and the recovery is `tofu state rm` followed by
# an apply.
resource "stripe_promotion_code" "this" {
  code             = local.code
  customer         = local.customer
  customer_account = local.customer_account
  expires_at       = local.expires_at
  max_redemptions  = local.max_redemptions
  active           = local.active
  metadata         = local.metadata

  promotion {
    type   = "coupon"
    coupon = var.spec.coupon
  }

  dynamic "restrictions" {
    for_each = local.restrictions == null ? [] : [local.restrictions]
    content {
      first_time_transaction  = restrictions.value.first_time_transaction
      minimum_amount          = restrictions.value.minimum_amount
      minimum_amount_currency = restrictions.value.minimum_amount_currency

      # The API takes currency_options as a map keyed by currency; the provider models it as a
      # list whose entries carry the currency in key.
      dynamic "currency_options" {
        for_each = restrictions.value.currency_options
        content {
          key            = currency_options.value.key
          minimum_amount = currency_options.value.minimum_amount
        }
      }
    }
  }
}
