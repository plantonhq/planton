variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "StripePaymentMethodConfiguration specification"
  type = object({
    name   = optional(string, "")
    active = optional(bool)
    parent = optional(string, "")
    acss_debit = optional(object({
      preference = optional(string, "")
    }))
    affirm = optional(object({
      preference = optional(string, "")
    }))
    afterpay_clearpay = optional(object({
      preference = optional(string, "")
    }))
    alipay = optional(object({
      preference = optional(string, "")
    }))
    alma = optional(object({
      preference = optional(string, "")
    }))
    amazon_pay = optional(object({
      preference = optional(string, "")
    }))
    apple_pay = optional(object({
      preference = optional(string, "")
    }))
    apple_pay_later = optional(object({
      preference = optional(string, "")
    }))
    au_becs_debit = optional(object({
      preference = optional(string, "")
    }))
    bacs_debit = optional(object({
      preference = optional(string, "")
    }))
    bancontact = optional(object({
      preference = optional(string, "")
    }))
    billie = optional(object({
      preference = optional(string, "")
    }))
    bizum = optional(object({
      preference = optional(string, "")
    }))
    blik = optional(object({
      preference = optional(string, "")
    }))
    boleto = optional(object({
      preference = optional(string, "")
    }))
    card = optional(object({
      preference = optional(string, "")
    }))
    cartes_bancaires = optional(object({
      preference = optional(string, "")
    }))
    cashapp = optional(object({
      preference = optional(string, "")
    }))
    crypto = optional(object({
      preference = optional(string, "")
    }))
    customer_balance = optional(object({
      preference = optional(string, "")
    }))
    eps = optional(object({
      preference = optional(string, "")
    }))
    fpx = optional(object({
      preference = optional(string, "")
    }))
    giropay = optional(object({
      preference = optional(string, "")
    }))
    google_pay = optional(object({
      preference = optional(string, "")
    }))
    grabpay = optional(object({
      preference = optional(string, "")
    }))
    ideal = optional(object({
      preference = optional(string, "")
    }))
    jcb = optional(object({
      preference = optional(string, "")
    }))
    kakao_pay = optional(object({
      preference = optional(string, "")
    }))
    klarna = optional(object({
      preference = optional(string, "")
    }))
    konbini = optional(object({
      preference = optional(string, "")
    }))
    kr_card = optional(object({
      preference = optional(string, "")
    }))
    link = optional(object({
      preference = optional(string, "")
    }))
    mb_way = optional(object({
      preference = optional(string, "")
    }))
    mobilepay = optional(object({
      preference = optional(string, "")
    }))
    multibanco = optional(object({
      preference = optional(string, "")
    }))
    naver_pay = optional(object({
      preference = optional(string, "")
    }))
    nz_bank_account = optional(object({
      preference = optional(string, "")
    }))
    oxxo = optional(object({
      preference = optional(string, "")
    }))
    p24 = optional(object({
      preference = optional(string, "")
    }))
    pay_by_bank = optional(object({
      preference = optional(string, "")
    }))
    payco = optional(object({
      preference = optional(string, "")
    }))
    paynow = optional(object({
      preference = optional(string, "")
    }))
    paypal = optional(object({
      preference = optional(string, "")
    }))
    payto = optional(object({
      preference = optional(string, "")
    }))
    pix = optional(object({
      preference = optional(string, "")
    }))
    promptpay = optional(object({
      preference = optional(string, "")
    }))
    revolut_pay = optional(object({
      preference = optional(string, "")
    }))
    samsung_pay = optional(object({
      preference = optional(string, "")
    }))
    satispay = optional(object({
      preference = optional(string, "")
    }))
    scalapay = optional(object({
      preference = optional(string, "")
    }))
    sepa_debit = optional(object({
      preference = optional(string, "")
    }))
    sofort = optional(object({
      preference = optional(string, "")
    }))
    sunbit = optional(object({
      preference = optional(string, "")
    }))
    swish = optional(object({
      preference = optional(string, "")
    }))
    twint = optional(object({
      preference = optional(string, "")
    }))
    upi = optional(object({
      preference = optional(string, "")
    }))
    us_bank_account = optional(object({
      preference = optional(string, "")
    }))
    wechat_pay = optional(object({
      preference = optional(string, "")
    }))
    zip = optional(object({
      preference = optional(string, "")
    }))
  })
}
