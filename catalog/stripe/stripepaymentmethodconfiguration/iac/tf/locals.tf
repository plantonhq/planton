# Local values for the StripePaymentMethodConfiguration module.
#
# A method the spec leaves out renders as null, so the provider never sends it and Stripe keeps
# the method's current setting (its default on a new configuration). A method removed from the
# manifest therefore keeps its last preference in Stripe -- set it to "none" to hand it back to
# Stripe's default, as the spec's comments tell the manifest author. The preference arrives as its
# enum value name, which is Stripe's own string (on, off, none).
locals {
  methods = [
    "acss_debit", "affirm", "afterpay_clearpay", "alipay", "alma", "amazon_pay", "apple_pay",
    "apple_pay_later", "au_becs_debit", "bacs_debit", "bancontact", "billie", "bizum", "blik",
    "boleto", "card", "cartes_bancaires", "cashapp", "crypto", "customer_balance", "eps", "fpx",
    "giropay", "google_pay", "grabpay", "ideal", "jcb", "kakao_pay", "klarna", "konbini",
    "kr_card", "link", "mb_way", "mobilepay", "multibanco", "naver_pay", "nz_bank_account", "oxxo",
    "p24", "pay_by_bank", "payco", "paynow", "paypal", "payto", "pix", "promptpay", "revolut_pay",
    "samsung_pay", "satispay", "scalapay", "sepa_debit", "sofort", "sunbit", "swish", "twint",
    "upi", "us_bank_account", "wechat_pay", "zip"
  ]

  display = {
    for m in local.methods : m => (
      try(var.spec[m].preference, "") == "" ? null : { display_preference = { preference = var.spec[m].preference } }
    )
  }

  name   = try(var.spec.name, "") != "" ? var.spec.name : null
  parent = try(var.spec.parent, "") != "" ? var.spec.parent : null
  # The spec's default is active; the provider needs the value only to deactivate.
  active = try(var.spec.active, null) == null ? true : var.spec.active
}
