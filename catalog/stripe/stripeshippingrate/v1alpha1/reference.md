# StripeShippingRate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeShippingRateSpec declares a shipping option customers choose in Checkout or on a payment
link: its name, a fixed amount in one or more currencies, and how long delivery takes.

A shipping rate's name, amount, currency, delivery estimate and tax code can never change in
Stripe. Changing any of them REPLACES the rate: Planton deactivates the old one and creates the
new one, and a StripePaymentLink that references it is replaced with it. Tax behavior, the
other currencies' amounts, active and metadata update in place.

One owner per object: declare a shipping rate here only if nothing else creates or edits it.

Stripe never deletes a shipping rate. Destroy deactivates it (active = false): new purchases
can't choose it, and orders already placed keep it. A rate that cannot be read makes the plan
fail (the provider does not treat a missing rate as gone); remove it from state and apply again.

The key needs "Shipping Rates" write (iac/permissions.yaml).

https://docs.stripe.com/payments/during-payment/charge-shipping
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/shipping_rate

## Example

```yaml
# The canonical example: standard shipping at 5 USD (4.50 EUR), 3 to 5
# business days, tax added on top.
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeShippingRate
metadata:
  name: standard-shipping
  org: e2e-org
  env: testing
spec:
  displayName: Standard
  fixedAmount:
    amount: 500
    currency: usd
    currencyOptions:
      eur:
        amount: 450
        taxBehavior: exclusive
  deliveryEstimate:
    minimum:
      unit: business_day
      value: 3
    maximum:
      unit: business_day
      value: 5
  taxBehavior: exclusive
  taxCode: txcd_92010001
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.displayName` | `string` | yes |  |  |
| `spec.fixedAmount` | `StripeShippingRateFixedAmount` | yes |  |  |
| `spec.fixedAmount.amount` | `int64` |  |  |  |
| `spec.fixedAmount.currency` | `string` | yes |  |  |
| `spec.fixedAmount.currencyOptions` | `map<string, StripeShippingRateCurrencyOption>` |  |  |  |
| `spec.fixedAmount.currencyOptions.*.amount` | `int64` |  |  |  |
| `spec.fixedAmount.currencyOptions.*.taxBehavior` | `enum` |  |  |  |
| `spec.deliveryEstimate` | `StripeShippingRateDeliveryEstimate` |  |  |  |
| `spec.deliveryEstimate.minimum` | `StripeShippingRateDeliveryBound` |  |  |  |
| `spec.deliveryEstimate.minimum.unit` | `enum` |  |  |  |
| `spec.deliveryEstimate.minimum.value` | `int64` |  |  |  |
| `spec.deliveryEstimate.maximum` | `StripeShippingRateDeliveryBound` |  |  |  |
| `spec.deliveryEstimate.maximum.unit` | `enum` |  |  |  |
| `spec.deliveryEstimate.maximum.value` | `int64` |  |  |  |
| `spec.taxBehavior` | `enum` |  |  |  |
| `spec.taxCode` | `string` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.displayName

`string` · required

display_name is the option's name as customers see it in Checkout ("Standard", "Express").
Changing it REPLACES the rate.

- rule: {"required":true}

### spec.fixedAmount

`StripeShippingRateFixedAmount` · required

fixed_amount is what the option costs, the only kind of shipping rate Stripe offers.

- rule: {"required":true}
- rule: currency_options are the other currencies: the main currency is not one of them

### spec.fixedAmount.amount

`int64`

amount is the charge in the currency's smallest unit (cents for usd); 0 is free shipping.
Changing it REPLACES the rate.

- rule: {"int64":{"gte":"0"}}

### spec.fixedAmount.currency

`string` · required

currency is the lowercase three-letter ISO code of amount ("usd"). Changing it REPLACES the
rate.

- rule: {"required":true,"string":{"pattern":"^[a-z]{3}$"}}

### spec.fixedAmount.currencyOptions

`map<string, StripeShippingRateCurrencyOption>`

currency_options are the amount in other currencies, keyed by lowercase currency code
("eur"), so a customer paying in theirs sees their price. They update in place.

- rule: {"map":{"keys":{"string":{"pattern":"^[a-z]{3}$"}}}}

### spec.fixedAmount.currencyOptions.*.amount

`int64`

amount is the charge in this currency's smallest unit.

- rule: {"int64":{"gte":"0"}}

### spec.fixedAmount.currencyOptions.*.taxBehavior

`enum`

tax_behavior is whether this amount includes tax.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `tax_behavior_unset`
- `exclusive`
- `inclusive`
- `unspecified`

### spec.deliveryEstimate

`StripeShippingRateDeliveryEstimate`

delivery_estimate is the delivery window shown to customers ("3-5 business days"). Changing
it REPLACES the rate.

- rule: a delivery estimate sets minimum, maximum, or both

### spec.deliveryEstimate.minimum

`StripeShippingRateDeliveryBound`

minimum is the earliest delivery. Unset, no lower bound.

### spec.deliveryEstimate.minimum.unit

`enum`

unit is the unit of time.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `unit_unspecified`
- `business_day`
- `day`
- `hour`
- `week`
- `month`

### spec.deliveryEstimate.minimum.value

`int64`

value is how many units, above 0.

- rule: {"int64":{"gt":"0"}}

### spec.deliveryEstimate.maximum

`StripeShippingRateDeliveryBound`

maximum is the latest delivery. Unset, no upper bound.

### spec.deliveryEstimate.maximum.unit

`enum`

unit is the unit of time.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `unit_unspecified`
- `business_day`
- `day`
- `hour`
- `week`
- `month`

### spec.deliveryEstimate.maximum.value

`int64`

value is how many units, above 0.

- rule: {"int64":{"gt":"0"}}

### spec.taxBehavior

`enum`

tax_behavior is whether the amount includes tax. Unset, the account's default applies. Once
inclusive or exclusive, Stripe refuses to change it.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `tax_behavior_unset`
- `exclusive` -- exclusive adds tax on top of the amount.
- `inclusive` -- inclusive includes tax in the amount.
- `unspecified` -- unspecified leaves the choice open until it is set, once, to inclusive or exclusive.

### spec.taxCode

`string`

tax_code is the Stripe Tax category the shipping is taxed as. Stripe's shipping tax code is
txcd_92010001. Changing it REPLACES the rate.

- rule: tax_code is a Stripe tax code id such as txcd_92010001

### spec.active

`bool` · optional (explicit presence)

active is whether new purchases can choose the rate. Setting it false deactivates the rate
without destroying the resource.

- default: `true`

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the rate. It updates in place.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeShippingRate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the shipping rate's Stripe id (shr_...), the value a StripePaymentLink's shipping options and a Checkout session reference. It changes when the rate is replaced. |
| `status.outputs.active` | `bool` | active is false once the rate is deactivated (by destroy, by spec.active, or in the Dashboard). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| StripePaymentLink | `spec.shippingOptions[].shippingRate` | `status.outputs.id` |

## See Also

- [Overview](../README.md)
