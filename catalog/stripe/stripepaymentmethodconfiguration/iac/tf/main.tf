# StripePaymentMethodConfiguration Main Resources
#
# stripe_payment_method_configuration is one payment-method configuration in the account the
# provider's key belongs to. It is always a new configuration: the module never adopts the
# account's default, and the application names this one's id when it creates a Checkout Session
# or PaymentIntent.
#
# Each payment method is one nested object whose only settable member is
# display_preference.preference; the rest (available, and display_preference.value and
# overridable) is what Stripe reports back. apple_pay_later is write-only in the provider: it is
# sent on every apply and never read back, so it never shows drift.
#
# Stripe's create call cannot set active, so the provider makes a second, update call right after
# create when active is false. Stripe never deletes a configuration: destroy is an update to
# active = false, and Stripe keeps it forever. parent is create-only, so changing it replaces the
# configuration. The provider has no handling for a configuration that cannot be read; the next
# refresh fails until it is removed from state.
resource "stripe_payment_method_configuration" "this" {
  name   = local.name
  active = local.active
  parent = local.parent

  acss_debit        = local.display["acss_debit"]
  affirm            = local.display["affirm"]
  afterpay_clearpay = local.display["afterpay_clearpay"]
  alipay            = local.display["alipay"]
  alma              = local.display["alma"]
  amazon_pay        = local.display["amazon_pay"]
  apple_pay         = local.display["apple_pay"]
  apple_pay_later   = local.display["apple_pay_later"]
  au_becs_debit     = local.display["au_becs_debit"]
  bacs_debit        = local.display["bacs_debit"]
  bancontact        = local.display["bancontact"]
  billie            = local.display["billie"]
  bizum             = local.display["bizum"]
  blik              = local.display["blik"]
  boleto            = local.display["boleto"]
  card              = local.display["card"]
  cartes_bancaires  = local.display["cartes_bancaires"]
  cashapp           = local.display["cashapp"]
  crypto            = local.display["crypto"]
  customer_balance  = local.display["customer_balance"]
  eps               = local.display["eps"]
  fpx               = local.display["fpx"]
  giropay           = local.display["giropay"]
  google_pay        = local.display["google_pay"]
  grabpay           = local.display["grabpay"]
  ideal             = local.display["ideal"]
  jcb               = local.display["jcb"]
  kakao_pay         = local.display["kakao_pay"]
  klarna            = local.display["klarna"]
  konbini           = local.display["konbini"]
  kr_card           = local.display["kr_card"]
  link              = local.display["link"]
  mb_way            = local.display["mb_way"]
  mobilepay         = local.display["mobilepay"]
  multibanco        = local.display["multibanco"]
  naver_pay         = local.display["naver_pay"]
  nz_bank_account   = local.display["nz_bank_account"]
  oxxo              = local.display["oxxo"]
  p24               = local.display["p24"]
  pay_by_bank       = local.display["pay_by_bank"]
  payco             = local.display["payco"]
  paynow            = local.display["paynow"]
  paypal            = local.display["paypal"]
  payto             = local.display["payto"]
  pix               = local.display["pix"]
  promptpay         = local.display["promptpay"]
  revolut_pay       = local.display["revolut_pay"]
  samsung_pay       = local.display["samsung_pay"]
  satispay          = local.display["satispay"]
  scalapay          = local.display["scalapay"]
  sepa_debit        = local.display["sepa_debit"]
  sofort            = local.display["sofort"]
  sunbit            = local.display["sunbit"]
  swish             = local.display["swish"]
  twint             = local.display["twint"]
  upi               = local.display["upi"]
  us_bank_account   = local.display["us_bank_account"]
  wechat_pay        = local.display["wechat_pay"]
  zip               = local.display["zip"]
}
