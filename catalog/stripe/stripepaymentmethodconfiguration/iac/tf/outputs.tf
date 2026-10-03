# StripePaymentMethodConfiguration Outputs
# Maps to the StripePaymentMethodConfigurationOutputs protobuf message: the configuration a
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
  # Every method's computed `available`, in the spec's order. Stripe leaves it empty for a method
  # that is off or not offered to the account, so only `true` counts. The provider reads no
  # availability for card, and apple_pay_later is write-only, so neither can appear.
  value = compact([
    try(stripe_payment_method_configuration.this.acss_debit.available == true, false) ? "acss_debit" : "",
    try(stripe_payment_method_configuration.this.affirm.available == true, false) ? "affirm" : "",
    try(stripe_payment_method_configuration.this.afterpay_clearpay.available == true, false) ? "afterpay_clearpay" : "",
    try(stripe_payment_method_configuration.this.alipay.available == true, false) ? "alipay" : "",
    try(stripe_payment_method_configuration.this.alma.available == true, false) ? "alma" : "",
    try(stripe_payment_method_configuration.this.amazon_pay.available == true, false) ? "amazon_pay" : "",
    try(stripe_payment_method_configuration.this.apple_pay.available == true, false) ? "apple_pay" : "",
    try(stripe_payment_method_configuration.this.au_becs_debit.available == true, false) ? "au_becs_debit" : "",
    try(stripe_payment_method_configuration.this.bacs_debit.available == true, false) ? "bacs_debit" : "",
    try(stripe_payment_method_configuration.this.bancontact.available == true, false) ? "bancontact" : "",
    try(stripe_payment_method_configuration.this.billie.available == true, false) ? "billie" : "",
    try(stripe_payment_method_configuration.this.bizum.available == true, false) ? "bizum" : "",
    try(stripe_payment_method_configuration.this.blik.available == true, false) ? "blik" : "",
    try(stripe_payment_method_configuration.this.boleto.available == true, false) ? "boleto" : "",
    try(stripe_payment_method_configuration.this.cartes_bancaires.available == true, false) ? "cartes_bancaires" : "",
    try(stripe_payment_method_configuration.this.cashapp.available == true, false) ? "cashapp" : "",
    try(stripe_payment_method_configuration.this.crypto.available == true, false) ? "crypto" : "",
    try(stripe_payment_method_configuration.this.customer_balance.available == true, false) ? "customer_balance" : "",
    try(stripe_payment_method_configuration.this.eps.available == true, false) ? "eps" : "",
    try(stripe_payment_method_configuration.this.fpx.available == true, false) ? "fpx" : "",
    try(stripe_payment_method_configuration.this.giropay.available == true, false) ? "giropay" : "",
    try(stripe_payment_method_configuration.this.google_pay.available == true, false) ? "google_pay" : "",
    try(stripe_payment_method_configuration.this.grabpay.available == true, false) ? "grabpay" : "",
    try(stripe_payment_method_configuration.this.ideal.available == true, false) ? "ideal" : "",
    try(stripe_payment_method_configuration.this.jcb.available == true, false) ? "jcb" : "",
    try(stripe_payment_method_configuration.this.kakao_pay.available == true, false) ? "kakao_pay" : "",
    try(stripe_payment_method_configuration.this.klarna.available == true, false) ? "klarna" : "",
    try(stripe_payment_method_configuration.this.konbini.available == true, false) ? "konbini" : "",
    try(stripe_payment_method_configuration.this.kr_card.available == true, false) ? "kr_card" : "",
    try(stripe_payment_method_configuration.this.link.available == true, false) ? "link" : "",
    try(stripe_payment_method_configuration.this.mb_way.available == true, false) ? "mb_way" : "",
    try(stripe_payment_method_configuration.this.mobilepay.available == true, false) ? "mobilepay" : "",
    try(stripe_payment_method_configuration.this.multibanco.available == true, false) ? "multibanco" : "",
    try(stripe_payment_method_configuration.this.naver_pay.available == true, false) ? "naver_pay" : "",
    try(stripe_payment_method_configuration.this.nz_bank_account.available == true, false) ? "nz_bank_account" : "",
    try(stripe_payment_method_configuration.this.oxxo.available == true, false) ? "oxxo" : "",
    try(stripe_payment_method_configuration.this.p24.available == true, false) ? "p24" : "",
    try(stripe_payment_method_configuration.this.pay_by_bank.available == true, false) ? "pay_by_bank" : "",
    try(stripe_payment_method_configuration.this.payco.available == true, false) ? "payco" : "",
    try(stripe_payment_method_configuration.this.paynow.available == true, false) ? "paynow" : "",
    try(stripe_payment_method_configuration.this.paypal.available == true, false) ? "paypal" : "",
    try(stripe_payment_method_configuration.this.payto.available == true, false) ? "payto" : "",
    try(stripe_payment_method_configuration.this.pix.available == true, false) ? "pix" : "",
    try(stripe_payment_method_configuration.this.promptpay.available == true, false) ? "promptpay" : "",
    try(stripe_payment_method_configuration.this.revolut_pay.available == true, false) ? "revolut_pay" : "",
    try(stripe_payment_method_configuration.this.samsung_pay.available == true, false) ? "samsung_pay" : "",
    try(stripe_payment_method_configuration.this.satispay.available == true, false) ? "satispay" : "",
    try(stripe_payment_method_configuration.this.scalapay.available == true, false) ? "scalapay" : "",
    try(stripe_payment_method_configuration.this.sepa_debit.available == true, false) ? "sepa_debit" : "",
    try(stripe_payment_method_configuration.this.sofort.available == true, false) ? "sofort" : "",
    try(stripe_payment_method_configuration.this.sunbit.available == true, false) ? "sunbit" : "",
    try(stripe_payment_method_configuration.this.swish.available == true, false) ? "swish" : "",
    try(stripe_payment_method_configuration.this.twint.available == true, false) ? "twint" : "",
    try(stripe_payment_method_configuration.this.upi.available == true, false) ? "upi" : "",
    try(stripe_payment_method_configuration.this.us_bank_account.available == true, false) ? "us_bank_account" : "",
    try(stripe_payment_method_configuration.this.wechat_pay.available == true, false) ? "wechat_pay" : "",
    try(stripe_payment_method_configuration.this.zip.available == true, false) ? "zip" : "",
  ])
}
