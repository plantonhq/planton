# StripePaymentMethodConfiguration Outputs
# Maps to the StripePaymentMethodConfigurationStackOutputs protobuf message: the configuration a
# payment names and the methods it actually offers.

output "id" {
  description = "The configuration's Stripe id (pmc_...), passed as payment_method_configuration when a Checkout Session or PaymentIntent is created"
  value       = stripe_payment_method_configuration.this.id
}

output "is_default" {
  description = "Whether this is the account's default configuration"
  value       = stripe_payment_method_configuration.this.is_default
}

output "active" {
  description = "Whether payments may use the configuration"
  value       = stripe_payment_method_configuration.this.active
}

output "available_payment_methods" {
  description = "The methods Stripe reports available: set on and with the capability active on the account (Stripe reports no availability for cards or Apple Pay Later)"
  # Every method's computed `available`, in the spec's order. The provider reads no availability
  # for card, and apple_pay_later is write-only, so neither can appear.
  value = compact([
    try(stripe_payment_method_configuration.this.acss_debit.available, false) ? "acss_debit" : "",
    try(stripe_payment_method_configuration.this.affirm.available, false) ? "affirm" : "",
    try(stripe_payment_method_configuration.this.afterpay_clearpay.available, false) ? "afterpay_clearpay" : "",
    try(stripe_payment_method_configuration.this.alipay.available, false) ? "alipay" : "",
    try(stripe_payment_method_configuration.this.alma.available, false) ? "alma" : "",
    try(stripe_payment_method_configuration.this.amazon_pay.available, false) ? "amazon_pay" : "",
    try(stripe_payment_method_configuration.this.apple_pay.available, false) ? "apple_pay" : "",
    try(stripe_payment_method_configuration.this.au_becs_debit.available, false) ? "au_becs_debit" : "",
    try(stripe_payment_method_configuration.this.bacs_debit.available, false) ? "bacs_debit" : "",
    try(stripe_payment_method_configuration.this.bancontact.available, false) ? "bancontact" : "",
    try(stripe_payment_method_configuration.this.billie.available, false) ? "billie" : "",
    try(stripe_payment_method_configuration.this.bizum.available, false) ? "bizum" : "",
    try(stripe_payment_method_configuration.this.blik.available, false) ? "blik" : "",
    try(stripe_payment_method_configuration.this.boleto.available, false) ? "boleto" : "",
    try(stripe_payment_method_configuration.this.cartes_bancaires.available, false) ? "cartes_bancaires" : "",
    try(stripe_payment_method_configuration.this.cashapp.available, false) ? "cashapp" : "",
    try(stripe_payment_method_configuration.this.crypto.available, false) ? "crypto" : "",
    try(stripe_payment_method_configuration.this.customer_balance.available, false) ? "customer_balance" : "",
    try(stripe_payment_method_configuration.this.eps.available, false) ? "eps" : "",
    try(stripe_payment_method_configuration.this.fpx.available, false) ? "fpx" : "",
    try(stripe_payment_method_configuration.this.giropay.available, false) ? "giropay" : "",
    try(stripe_payment_method_configuration.this.google_pay.available, false) ? "google_pay" : "",
    try(stripe_payment_method_configuration.this.grabpay.available, false) ? "grabpay" : "",
    try(stripe_payment_method_configuration.this.ideal.available, false) ? "ideal" : "",
    try(stripe_payment_method_configuration.this.jcb.available, false) ? "jcb" : "",
    try(stripe_payment_method_configuration.this.kakao_pay.available, false) ? "kakao_pay" : "",
    try(stripe_payment_method_configuration.this.klarna.available, false) ? "klarna" : "",
    try(stripe_payment_method_configuration.this.konbini.available, false) ? "konbini" : "",
    try(stripe_payment_method_configuration.this.kr_card.available, false) ? "kr_card" : "",
    try(stripe_payment_method_configuration.this.link.available, false) ? "link" : "",
    try(stripe_payment_method_configuration.this.mb_way.available, false) ? "mb_way" : "",
    try(stripe_payment_method_configuration.this.mobilepay.available, false) ? "mobilepay" : "",
    try(stripe_payment_method_configuration.this.multibanco.available, false) ? "multibanco" : "",
    try(stripe_payment_method_configuration.this.naver_pay.available, false) ? "naver_pay" : "",
    try(stripe_payment_method_configuration.this.nz_bank_account.available, false) ? "nz_bank_account" : "",
    try(stripe_payment_method_configuration.this.oxxo.available, false) ? "oxxo" : "",
    try(stripe_payment_method_configuration.this.p24.available, false) ? "p24" : "",
    try(stripe_payment_method_configuration.this.pay_by_bank.available, false) ? "pay_by_bank" : "",
    try(stripe_payment_method_configuration.this.payco.available, false) ? "payco" : "",
    try(stripe_payment_method_configuration.this.paynow.available, false) ? "paynow" : "",
    try(stripe_payment_method_configuration.this.paypal.available, false) ? "paypal" : "",
    try(stripe_payment_method_configuration.this.payto.available, false) ? "payto" : "",
    try(stripe_payment_method_configuration.this.pix.available, false) ? "pix" : "",
    try(stripe_payment_method_configuration.this.promptpay.available, false) ? "promptpay" : "",
    try(stripe_payment_method_configuration.this.revolut_pay.available, false) ? "revolut_pay" : "",
    try(stripe_payment_method_configuration.this.samsung_pay.available, false) ? "samsung_pay" : "",
    try(stripe_payment_method_configuration.this.satispay.available, false) ? "satispay" : "",
    try(stripe_payment_method_configuration.this.scalapay.available, false) ? "scalapay" : "",
    try(stripe_payment_method_configuration.this.sepa_debit.available, false) ? "sepa_debit" : "",
    try(stripe_payment_method_configuration.this.sofort.available, false) ? "sofort" : "",
    try(stripe_payment_method_configuration.this.sunbit.available, false) ? "sunbit" : "",
    try(stripe_payment_method_configuration.this.swish.available, false) ? "swish" : "",
    try(stripe_payment_method_configuration.this.twint.available, false) ? "twint" : "",
    try(stripe_payment_method_configuration.this.upi.available, false) ? "upi" : "",
    try(stripe_payment_method_configuration.this.us_bank_account.available, false) ? "us_bank_account" : "",
    try(stripe_payment_method_configuration.this.wechat_pay.available, false) ? "wechat_pay" : "",
    try(stripe_payment_method_configuration.this.zip.available, false) ? "zip" : "",
  ])
}
