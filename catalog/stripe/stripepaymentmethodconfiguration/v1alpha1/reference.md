# StripePaymentMethodConfiguration

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripePaymentMethodConfigurationSpec declares which payment methods checkout offers, method by
method: on, off, or left to Stripe.

An account holds many configurations and one of them is its default, which Stripe uses when a
checkout or payment names none. This kind always creates a configuration of its own and never
adopts or edits the default: the application names this configuration's id
(status.outputs.id) when it creates a Checkout Session or PaymentIntent, and a portal
configuration can reference it so the portal offers the same methods. A method is shown only
where its capability is active on the account and the payment qualifies (currency, country,
amount); status.outputs.available_payment_methods lists the methods Stripe reports available.

Stripe never deletes a configuration. Destroy deactivates it (active = false) and Stripe keeps
it, inactive, forever. A method removed from the manifest keeps its last preference in Stripe
-- the provider sends only values that are set -- so set it to "none" to hand it back to
Stripe's default. A configuration that cannot be read makes the plan fail until it is removed
from state.

Stripe's API also accepts fr_meal_voucher_conecs (French meal vouchers); the pinned provider
does not expose it, so it is left to Stripe's default here.

The key needs "Payment Method Configurations" write (iac/permissions.yaml).

https://docs.stripe.com/payments/payment-method-configurations
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/payment_method_configuration

## Example

```yaml
# The canonical example: cards and wallets on, Link off. The lanes run
# against the dedicated test sandbox only.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentMethodConfiguration
metadata:
  name: checkout-methods
  org: e2e-org
  env: testing
spec:
  name: Planton catalog example methods
  card:
    preference: "on"
  applePay:
    preference: "on"
  googlePay:
    preference: "on"
  link:
    preference: "off"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.parent` | `string` |  |  |  |
| `spec.acssDebit` | `StripePaymentMethodPreference` |  |  |  |
| `spec.acssDebit.preference` | `enum` |  |  |  |
| `spec.affirm` | `StripePaymentMethodPreference` |  |  |  |
| `spec.affirm.preference` | `enum` |  |  |  |
| `spec.afterpayClearpay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.afterpayClearpay.preference` | `enum` |  |  |  |
| `spec.alipay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.alipay.preference` | `enum` |  |  |  |
| `spec.alma` | `StripePaymentMethodPreference` |  |  |  |
| `spec.alma.preference` | `enum` |  |  |  |
| `spec.amazonPay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.amazonPay.preference` | `enum` |  |  |  |
| `spec.applePay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.applePay.preference` | `enum` |  |  |  |
| `spec.applePayLater` | `StripePaymentMethodPreference` |  |  |  |
| `spec.applePayLater.preference` | `enum` |  |  |  |
| `spec.auBecsDebit` | `StripePaymentMethodPreference` |  |  |  |
| `spec.auBecsDebit.preference` | `enum` |  |  |  |
| `spec.bacsDebit` | `StripePaymentMethodPreference` |  |  |  |
| `spec.bacsDebit.preference` | `enum` |  |  |  |
| `spec.bancontact` | `StripePaymentMethodPreference` |  |  |  |
| `spec.bancontact.preference` | `enum` |  |  |  |
| `spec.billie` | `StripePaymentMethodPreference` |  |  |  |
| `spec.billie.preference` | `enum` |  |  |  |
| `spec.bizum` | `StripePaymentMethodPreference` |  |  |  |
| `spec.bizum.preference` | `enum` |  |  |  |
| `spec.blik` | `StripePaymentMethodPreference` |  |  |  |
| `spec.blik.preference` | `enum` |  |  |  |
| `spec.boleto` | `StripePaymentMethodPreference` |  |  |  |
| `spec.boleto.preference` | `enum` |  |  |  |
| `spec.card` | `StripePaymentMethodPreference` |  |  |  |
| `spec.card.preference` | `enum` |  |  |  |
| `spec.cartesBancaires` | `StripePaymentMethodPreference` |  |  |  |
| `spec.cartesBancaires.preference` | `enum` |  |  |  |
| `spec.cashapp` | `StripePaymentMethodPreference` |  |  |  |
| `spec.cashapp.preference` | `enum` |  |  |  |
| `spec.crypto` | `StripePaymentMethodPreference` |  |  |  |
| `spec.crypto.preference` | `enum` |  |  |  |
| `spec.customerBalance` | `StripePaymentMethodPreference` |  |  |  |
| `spec.customerBalance.preference` | `enum` |  |  |  |
| `spec.eps` | `StripePaymentMethodPreference` |  |  |  |
| `spec.eps.preference` | `enum` |  |  |  |
| `spec.fpx` | `StripePaymentMethodPreference` |  |  |  |
| `spec.fpx.preference` | `enum` |  |  |  |
| `spec.giropay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.giropay.preference` | `enum` |  |  |  |
| `spec.googlePay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.googlePay.preference` | `enum` |  |  |  |
| `spec.grabpay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.grabpay.preference` | `enum` |  |  |  |
| `spec.ideal` | `StripePaymentMethodPreference` |  |  |  |
| `spec.ideal.preference` | `enum` |  |  |  |
| `spec.jcb` | `StripePaymentMethodPreference` |  |  |  |
| `spec.jcb.preference` | `enum` |  |  |  |
| `spec.kakaoPay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.kakaoPay.preference` | `enum` |  |  |  |
| `spec.klarna` | `StripePaymentMethodPreference` |  |  |  |
| `spec.klarna.preference` | `enum` |  |  |  |
| `spec.konbini` | `StripePaymentMethodPreference` |  |  |  |
| `spec.konbini.preference` | `enum` |  |  |  |
| `spec.krCard` | `StripePaymentMethodPreference` |  |  |  |
| `spec.krCard.preference` | `enum` |  |  |  |
| `spec.link` | `StripePaymentMethodPreference` |  |  |  |
| `spec.link.preference` | `enum` |  |  |  |
| `spec.mbWay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.mbWay.preference` | `enum` |  |  |  |
| `spec.mobilepay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.mobilepay.preference` | `enum` |  |  |  |
| `spec.multibanco` | `StripePaymentMethodPreference` |  |  |  |
| `spec.multibanco.preference` | `enum` |  |  |  |
| `spec.naverPay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.naverPay.preference` | `enum` |  |  |  |
| `spec.nzBankAccount` | `StripePaymentMethodPreference` |  |  |  |
| `spec.nzBankAccount.preference` | `enum` |  |  |  |
| `spec.oxxo` | `StripePaymentMethodPreference` |  |  |  |
| `spec.oxxo.preference` | `enum` |  |  |  |
| `spec.p24` | `StripePaymentMethodPreference` |  |  |  |
| `spec.p24.preference` | `enum` |  |  |  |
| `spec.payByBank` | `StripePaymentMethodPreference` |  |  |  |
| `spec.payByBank.preference` | `enum` |  |  |  |
| `spec.payco` | `StripePaymentMethodPreference` |  |  |  |
| `spec.payco.preference` | `enum` |  |  |  |
| `spec.paynow` | `StripePaymentMethodPreference` |  |  |  |
| `spec.paynow.preference` | `enum` |  |  |  |
| `spec.paypal` | `StripePaymentMethodPreference` |  |  |  |
| `spec.paypal.preference` | `enum` |  |  |  |
| `spec.payto` | `StripePaymentMethodPreference` |  |  |  |
| `spec.payto.preference` | `enum` |  |  |  |
| `spec.pix` | `StripePaymentMethodPreference` |  |  |  |
| `spec.pix.preference` | `enum` |  |  |  |
| `spec.promptpay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.promptpay.preference` | `enum` |  |  |  |
| `spec.revolutPay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.revolutPay.preference` | `enum` |  |  |  |
| `spec.samsungPay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.samsungPay.preference` | `enum` |  |  |  |
| `spec.satispay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.satispay.preference` | `enum` |  |  |  |
| `spec.scalapay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.scalapay.preference` | `enum` |  |  |  |
| `spec.sepaDebit` | `StripePaymentMethodPreference` |  |  |  |
| `spec.sepaDebit.preference` | `enum` |  |  |  |
| `spec.sofort` | `StripePaymentMethodPreference` |  |  |  |
| `spec.sofort.preference` | `enum` |  |  |  |
| `spec.sunbit` | `StripePaymentMethodPreference` |  |  |  |
| `spec.sunbit.preference` | `enum` |  |  |  |
| `spec.swish` | `StripePaymentMethodPreference` |  |  |  |
| `spec.swish.preference` | `enum` |  |  |  |
| `spec.twint` | `StripePaymentMethodPreference` |  |  |  |
| `spec.twint.preference` | `enum` |  |  |  |
| `spec.upi` | `StripePaymentMethodPreference` |  |  |  |
| `spec.upi.preference` | `enum` |  |  |  |
| `spec.usBankAccount` | `StripePaymentMethodPreference` |  |  |  |
| `spec.usBankAccount.preference` | `enum` |  |  |  |
| `spec.wechatPay` | `StripePaymentMethodPreference` |  |  |  |
| `spec.wechatPay.preference` | `enum` |  |  |  |
| `spec.zip` | `StripePaymentMethodPreference` |  |  |  |
| `spec.zip.preference` | `enum` |  |  |  |

## Field Details

### spec.name

`string`

name is a label for the configuration, shown only to the account's team.

### spec.active

`bool` · optional (explicit presence)

active is whether payments may use this configuration. Setting it false deactivates the
configuration without destroying the resource; destroy deactivates it too.

- default: `true`

### spec.parent

`string`

parent is, for a Connect platform, the platform configuration (pmc_...) this child
configuration inherits from; its methods start from the parent's and override only what is
set here. Changing it REPLACES the configuration.

- rule: parent is a payment method configuration id (pmc_...)

### spec.acssDebit

`StripePaymentMethodPreference`

acss_debit is ACSS Debit (Canadian pre-authorized debits).

### spec.acssDebit.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.affirm

`StripePaymentMethodPreference`

affirm is Affirm.

### spec.affirm.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.afterpayClearpay

`StripePaymentMethodPreference`

afterpay_clearpay is Afterpay / Clearpay.

### spec.afterpayClearpay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.alipay

`StripePaymentMethodPreference`

alipay is Alipay.

### spec.alipay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.alma

`StripePaymentMethodPreference`

alma is Alma.

### spec.alma.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.amazonPay

`StripePaymentMethodPreference`

amazon_pay is Amazon Pay.

### spec.amazonPay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.applePay

`StripePaymentMethodPreference`

apple_pay is Apple Pay.

### spec.applePay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.applePayLater

`StripePaymentMethodPreference`

apple_pay_later is Apple Pay Later.
Stripe never returns this setting, so it is sent on every apply and never read back.

### spec.applePayLater.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.auBecsDebit

`StripePaymentMethodPreference`

au_becs_debit is BECS Direct Debit (Australia).

### spec.auBecsDebit.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.bacsDebit

`StripePaymentMethodPreference`

bacs_debit is Bacs Direct Debit (UK).

### spec.bacsDebit.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.bancontact

`StripePaymentMethodPreference`

bancontact is Bancontact.

### spec.bancontact.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.billie

`StripePaymentMethodPreference`

billie is Billie.

### spec.billie.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.bizum

`StripePaymentMethodPreference`

bizum is Bizum.

### spec.bizum.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.blik

`StripePaymentMethodPreference`

blik is BLIK.

### spec.blik.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.boleto

`StripePaymentMethodPreference`

boleto is Boleto.

### spec.boleto.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.card

`StripePaymentMethodPreference`

card is cards.

### spec.card.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.cartesBancaires

`StripePaymentMethodPreference`

cartes_bancaires is Cartes Bancaires.

### spec.cartesBancaires.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.cashapp

`StripePaymentMethodPreference`

cashapp is Cash App Pay.

### spec.cashapp.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.crypto

`StripePaymentMethodPreference`

crypto is stablecoin payments.

### spec.crypto.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.customerBalance

`StripePaymentMethodPreference`

customer_balance is bank transfers into the customer balance.

### spec.customerBalance.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.eps

`StripePaymentMethodPreference`

eps is EPS.

### spec.eps.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.fpx

`StripePaymentMethodPreference`

fpx is FPX.

### spec.fpx.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.giropay

`StripePaymentMethodPreference`

giropay is Giropay.

### spec.giropay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.googlePay

`StripePaymentMethodPreference`

google_pay is Google Pay.

### spec.googlePay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.grabpay

`StripePaymentMethodPreference`

grabpay is GrabPay.

### spec.grabpay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.ideal

`StripePaymentMethodPreference`

ideal is iDEAL.

### spec.ideal.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.jcb

`StripePaymentMethodPreference`

jcb is JCB.

### spec.jcb.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.kakaoPay

`StripePaymentMethodPreference`

kakao_pay is Kakao Pay.

### spec.kakaoPay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.klarna

`StripePaymentMethodPreference`

klarna is Klarna.

### spec.klarna.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.konbini

`StripePaymentMethodPreference`

konbini is Konbini.

### spec.konbini.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.krCard

`StripePaymentMethodPreference`

kr_card is Korean cards.

### spec.krCard.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.link

`StripePaymentMethodPreference`

link is Link, Stripe's saved-details checkout.

### spec.link.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.mbWay

`StripePaymentMethodPreference`

mb_way is MB WAY.

### spec.mbWay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.mobilepay

`StripePaymentMethodPreference`

mobilepay is MobilePay.

### spec.mobilepay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.multibanco

`StripePaymentMethodPreference`

multibanco is Multibanco.

### spec.multibanco.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.naverPay

`StripePaymentMethodPreference`

naver_pay is Naver Pay.

### spec.naverPay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.nzBankAccount

`StripePaymentMethodPreference`

nz_bank_account is New Zealand BECS Direct Debit.

### spec.nzBankAccount.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.oxxo

`StripePaymentMethodPreference`

oxxo is OXXO.

### spec.oxxo.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.p24

`StripePaymentMethodPreference`

p24 is Przelewy24.

### spec.p24.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.payByBank

`StripePaymentMethodPreference`

pay_by_bank is Pay by Bank.

### spec.payByBank.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.payco

`StripePaymentMethodPreference`

payco is PAYCO.

### spec.payco.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.paynow

`StripePaymentMethodPreference`

paynow is PayNow.

### spec.paynow.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.paypal

`StripePaymentMethodPreference`

paypal is PayPal.

### spec.paypal.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.payto

`StripePaymentMethodPreference`

payto is PayTo.

### spec.payto.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.pix

`StripePaymentMethodPreference`

pix is Pix.

### spec.pix.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.promptpay

`StripePaymentMethodPreference`

promptpay is PromptPay.

### spec.promptpay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.revolutPay

`StripePaymentMethodPreference`

revolut_pay is Revolut Pay.

### spec.revolutPay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.samsungPay

`StripePaymentMethodPreference`

samsung_pay is Samsung Pay.

### spec.samsungPay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.satispay

`StripePaymentMethodPreference`

satispay is Satispay.

### spec.satispay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.scalapay

`StripePaymentMethodPreference`

scalapay is Scalapay.

### spec.scalapay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.sepaDebit

`StripePaymentMethodPreference`

sepa_debit is SEPA Direct Debit.

### spec.sepaDebit.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.sofort

`StripePaymentMethodPreference`

sofort is Sofort.

### spec.sofort.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.sunbit

`StripePaymentMethodPreference`

sunbit is Sunbit.

### spec.sunbit.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.swish

`StripePaymentMethodPreference`

swish is Swish.

### spec.swish.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.twint

`StripePaymentMethodPreference`

twint is TWINT.

### spec.twint.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.upi

`StripePaymentMethodPreference`

upi is UPI.

### spec.upi.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.usBankAccount

`StripePaymentMethodPreference`

us_bank_account is ACH Direct Debit (US bank accounts).

### spec.usBankAccount.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.wechatPay

`StripePaymentMethodPreference`

wechat_pay is WeChat Pay.

### spec.wechatPay.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

### spec.zip

`StripePaymentMethodPreference`

zip is Zip.

### spec.zip.preference

`enum`

preference is whether checkout offers the method: "on", "off", or "none" to leave it to
Stripe's default (or, for a Connect child configuration, to the parent's). Planton reads a
bare on or off; quote them ("on") in files that YAML 1.1 tools such as PyYAML also read,
since those read a bare on or off as a boolean.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `preference_unspecified`
- `on`
- `off`
- `none`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripePaymentMethodConfiguration, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the configuration's Stripe id (pmc_...), the value an application passes as `payment_method_configuration` when it creates a Checkout Session or PaymentIntent. |
| `status.outputs.is_default` | `bool` | is_default is whether this is the account's default configuration. A configuration this kind creates is not the default unless someone makes it so in the Dashboard. |
| `status.outputs.active` | `bool` | active is whether payments may use the configuration. |
| `status.outputs.available_payment_methods` | `[]string` | available_payment_methods are the methods Stripe reports available: set on and with the method's capability active on the account. A method set on but missing here needs its capability turned on in the Dashboard. Stripe does not report availability for cards. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| StripeBillingPortalConfiguration | `spec.features.paymentMethodUpdate.paymentMethodConfiguration` | `status.outputs.id` |

## See Also

- [Overview](../README.md)
