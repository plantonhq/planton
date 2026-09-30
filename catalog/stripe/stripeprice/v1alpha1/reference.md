# StripePrice

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripePriceSpec declares how much, how often and in which currencies a StripeProduct is
charged: a flat amount, an amount per unit, graduated or volume tiers, or an amount the customer
chooses, one-time or recurring.

A price's amount can never change in Stripe. Changing the amount, currency, product, billing
scheme, tiers or recurrence REPLACES the price: Planton creates the new price first, then
archives the old one, so existing subscriptions keep the old amount and new purchases get the
new one. When the application finds the price by lookup_key, set transfer_lookup_key with the
change so the key moves to the new price in the same call. Nickname, lookup key, metadata,
currency options, tax behavior and active update in place.

One owner per object: declare a price here only if nothing else creates or edits it. When the
application's own code or the Dashboard owns the price list, leave its prices there.

Stripe never deletes a price. Destroy archives it (active = false): subscriptions on it keep
billing, new purchases are refused. A price that cannot be read makes the plan fail (the
provider does not treat a missing price as gone); remove it from state and apply again.

The provider's product_data (a product created inside the price) is not offered: that product
would be invisible to Planton and never archived on destroy. Name a StripeProduct instead.

The key needs "Prices" write (iac/permissions.yaml).

https://docs.stripe.com/products-prices/pricing-models
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/price

## Example

```yaml
# The canonical example: 49 USD a month, found by lookup key. The product is a
# literal id here so the manifest plans on its own; a real manifest references
# its StripeProduct (see the scenarios).
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePrice
metadata:
  name: pro-monthly
  org: e2e-org
  env: testing
spec:
  product:
    value: prod_pro
  currency: usd
  unitAmount: 4900
  recurring:
    interval: month
  lookupKey: pro-monthly
  nickname: Pro, monthly
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.product` | `string \| valueFrom` | yes |  | StripeProduct (`status.outputs.id`) |
| `spec.currency` | `string` | yes |  |  |
| `spec.unitAmount` | `int64` |  |  |  |
| `spec.unitAmountDecimal` | `string` |  |  |  |
| `spec.billingScheme` | `enum` |  |  |  |
| `spec.tiersMode` | `enum` |  |  |  |
| `spec.tiers` | `[]StripePriceTier` |  |  |  |
| `spec.tiers[].upTo` | `string` |  |  |  |
| `spec.tiers[].flatAmount` | `int64` |  |  |  |
| `spec.tiers[].flatAmountDecimal` | `string` |  |  |  |
| `spec.tiers[].unitAmount` | `int64` |  |  |  |
| `spec.tiers[].unitAmountDecimal` | `string` |  |  |  |
| `spec.customUnitAmount` | `StripePriceCustomUnitAmount` |  |  |  |
| `spec.customUnitAmount.minimum` | `int64` |  |  |  |
| `spec.customUnitAmount.maximum` | `int64` |  |  |  |
| `spec.customUnitAmount.preset` | `int64` |  |  |  |
| `spec.transformQuantity` | `StripePriceTransformQuantity` |  |  |  |
| `spec.transformQuantity.divideBy` | `int64` |  |  |  |
| `spec.transformQuantity.round` | `enum` |  |  |  |
| `spec.recurring` | `StripePriceRecurring` |  |  |  |
| `spec.recurring.interval` | `enum` |  |  |  |
| `spec.recurring.intervalCount` | `int64` |  |  |  |
| `spec.recurring.usageType` | `enum` |  |  |  |
| `spec.recurring.meter` | `string \| valueFrom` |  |  | StripeBillingMeter (`status.outputs.id`) |
| `spec.recurring.trialPeriodDays` | `int64` |  |  |  |
| `spec.currencyOptions` | `map<string, StripePriceCurrencyOption>` |  |  |  |
| `spec.currencyOptions.*.unitAmount` | `int64` |  |  |  |
| `spec.currencyOptions.*.unitAmountDecimal` | `string` |  |  |  |
| `spec.currencyOptions.*.taxBehavior` | `enum` |  |  |  |
| `spec.currencyOptions.*.tiers` | `[]StripePriceTier` |  |  |  |
| `spec.currencyOptions.*.tiers[].upTo` | `string` |  |  |  |
| `spec.currencyOptions.*.tiers[].flatAmount` | `int64` |  |  |  |
| `spec.currencyOptions.*.tiers[].flatAmountDecimal` | `string` |  |  |  |
| `spec.currencyOptions.*.tiers[].unitAmount` | `int64` |  |  |  |
| `spec.currencyOptions.*.tiers[].unitAmountDecimal` | `string` |  |  |  |
| `spec.currencyOptions.*.customUnitAmount` | `StripePriceCustomUnitAmount` |  |  |  |
| `spec.currencyOptions.*.customUnitAmount.minimum` | `int64` |  |  |  |
| `spec.currencyOptions.*.customUnitAmount.maximum` | `int64` |  |  |  |
| `spec.currencyOptions.*.customUnitAmount.preset` | `int64` |  |  |  |
| `spec.lookupKey` | `string` |  |  |  |
| `spec.transferLookupKey` | `bool` |  |  |  |
| `spec.nickname` | `string` |  |  |  |
| `spec.taxBehavior` | `enum` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.product

`string | valueFrom` · required

product is the product this price charges for (prod_...). Reference a StripeProduct.
Changing it REPLACES the price.

- references: StripeProduct (`status.outputs.id`)
- rule: product is a Stripe product id (prod_...), or a reference to a StripeProduct
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripeProduct, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.currency

`string` · required

currency is the three-letter ISO code of the price's main currency, in lowercase ("usd").
Changing it REPLACES the price.

- rule: {"required":true,"string":{"pattern":"^[a-z]{3}$"}}

### spec.unitAmount

`int64` · optional (explicit presence)

unit_amount is the charge in the currency's smallest unit (cents for usd), per unit or flat;
0 is a free price. A per-unit price sets exactly one of unit_amount, unit_amount_decimal and
custom_unit_amount. Changing it REPLACES the price.

- rule: {"int64":{"gte":"0"}}

### spec.unitAmountDecimal

`string`

unit_amount_decimal is unit_amount with up to 12 decimal places, for sub-cent unit prices
("0.25" is a quarter of a cent). Changing it REPLACES the price.

- rule: unit_amount_decimal is a non-negative decimal with at most 12 decimal places, such as 0.25

### spec.billingScheme

`enum`

billing_scheme is how the amount is computed. Unset, Stripe uses per_unit. Changing it
REPLACES the price.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `billing_scheme_unspecified`
- `per_unit` -- per_unit charges unit_amount per unit of quantity (or of metered usage).
- `tiered` -- tiered charges by tiers, per tiers_mode.

### spec.tiersMode

`enum`

tiers_mode is how tiers apply; a tiered price sets it. Changing it REPLACES the price.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `tiers_mode_unspecified`
- `graduated` -- graduated bills each unit at the tier it falls in, so the rate changes as quantity grows.
- `volume` -- volume bills every unit at the tier the whole quantity reaches.

### spec.tiers

`[]StripePriceTier`

tiers are the price's steps, lowest first; a tiered price sets them and the last one's up_to
is "inf". Changing them REPLACES the price.

- rule: a tier charges something: set flat_amount or unit_amount (or a decimal form), at most one form of each

### spec.tiers[].upTo

`string`

up_to is the tier's upper bound in units, or "inf" for the last tier. A tier starts one unit
above the previous tier's up_to.

- rule: up_to is a whole number of units, or inf for the last tier

### spec.tiers[].flatAmount

`int64` · optional (explicit presence)

flat_amount is charged once for the whole tier, in the currency's smallest unit.

- rule: {"int64":{"gte":"0"}}

### spec.tiers[].flatAmountDecimal

`string`

flat_amount_decimal is flat_amount as a decimal string.

### spec.tiers[].unitAmount

`int64` · optional (explicit presence)

unit_amount is charged per unit in the tier, in the currency's smallest unit.

- rule: {"int64":{"gte":"0"}}

### spec.tiers[].unitAmountDecimal

`string`

unit_amount_decimal is unit_amount with up to 12 decimal places.

### spec.customUnitAmount

`StripePriceCustomUnitAmount`

custom_unit_amount lets the customer choose the amount in Checkout and Payment Links (a
donation, a pay-what-you-want price). Present means enabled. Changing it REPLACES the price.

- rule: custom_unit_amount keeps minimum <= preset <= maximum

### spec.customUnitAmount.minimum

`int64` · optional (explicit presence)

minimum is the smallest amount the customer may enter, at least Stripe's minimum charge.

- rule: {"int64":{"gte":"0"}}

### spec.customUnitAmount.maximum

`int64` · optional (explicit presence)

maximum is the largest amount the customer may enter.

- rule: {"int64":{"gte":"0"}}

### spec.customUnitAmount.preset

`int64` · optional (explicit presence)

preset is the amount the field starts at.

- rule: {"int64":{"gte":"0"}}

### spec.transformQuantity

`StripePriceTransformQuantity`

transform_quantity divides the quantity (or reported usage) before billing -- "per 1,000
requests". It cannot be combined with tiers. Changing it REPLACES the price.

### spec.transformQuantity.divideBy

`int64`

divide_by is the number the quantity (or usage) is divided by.

- rule: {"int64":{"gt":"0"}}

### spec.transformQuantity.round

`enum`

round is how the division's remainder is rounded.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `round_unspecified`
- `up` -- up bills the partial unit as a whole one.
- `down` -- down drops the partial unit.

### spec.recurring

`StripePriceRecurring`

recurring makes the price a subscription price; unset, the price is one-time. Changing it
REPLACES the price.

- rule: a billing period is at most three years (3 years, 36 months, 156 weeks or 1095 days)
- rule: a meter tracks usage, so it needs usage_type metered
- rule: a metered price bills what a meter counts: name its meter (a reference to a StripeBillingMeter)

### spec.recurring.interval

`enum`

interval is the billing period's unit.

- rule: {"enum":{"definedOnly":true,"notIn":[0]}}

Allowed values (use exactly as shown):

- `interval_unspecified`
- `day`
- `week`
- `month`
- `year`

### spec.recurring.intervalCount

`int64` · optional (explicit presence)

interval_count is how many intervals make one billing period ("month" and 3 bills
quarterly), up to three years. Unset, 1.

- rule: {"int64":{"gte":"1"}}

### spec.recurring.usageType

`enum`

usage_type is how the quantity per period is found. Unset, Stripe uses licensed.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `usage_type_unspecified`
- `licensed` -- licensed bills the quantity set on the subscription.
- `metered` -- metered bills the usage reported during the period.

### spec.recurring.meter

`string | valueFrom`

meter is the billing meter (mtr_...) whose usage a metered price bills. Reference a
StripeBillingMeter; a metered price needs one, and a licensed price takes none. Changing it
REPLACES the price, and so does a replacement of the referenced meter.

- references: StripeBillingMeter (`status.outputs.id`)
- rule: meter is a Stripe billing meter id (mtr_...), or a reference to a StripeBillingMeter
- rule: write as {value: <literal>} or {valueFrom: {kind: StripeBillingMeter, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.recurring.trialPeriodDays

`int64` · optional (explicit presence)

trial_period_days is the trial a subscription gets when it is created with trial_from_plan.

- rule: {"int64":{"gte":"0"}}

### spec.currencyOptions

`map<string, StripePriceCurrencyOption>`

currency_options are the price's amounts in other currencies, keyed by lowercase currency code
("eur", "gbp"), so a customer can pay in theirs. They update in place. Stripe does not report
them back, so a change made in the Dashboard is not detected, and removing a currency here
does not remove it in Stripe.

- rule: {"map":{"keys":{"string":{"pattern":"^[a-z]{3}$"}}}}

### spec.currencyOptions.*.unitAmount

`int64` · optional (explicit presence)

unit_amount is the charge in this currency's smallest unit.

- rule: {"int64":{"gte":"0"}}

### spec.currencyOptions.*.unitAmountDecimal

`string`

unit_amount_decimal is unit_amount with up to 12 decimal places.

### spec.currencyOptions.*.taxBehavior

`enum`

tax_behavior is whether this amount includes tax.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `tax_behavior_unset`
- `exclusive`
- `inclusive`
- `unspecified`

### spec.currencyOptions.*.tiers

`[]StripePriceTier`

tiers are this currency's steps, for a tiered price.

- rule: a tier charges something: set flat_amount or unit_amount (or a decimal form), at most one form of each

### spec.currencyOptions.*.tiers[].upTo

`string`

up_to is the tier's upper bound in units, or "inf" for the last tier. A tier starts one unit
above the previous tier's up_to.

- rule: up_to is a whole number of units, or inf for the last tier

### spec.currencyOptions.*.tiers[].flatAmount

`int64` · optional (explicit presence)

flat_amount is charged once for the whole tier, in the currency's smallest unit.

- rule: {"int64":{"gte":"0"}}

### spec.currencyOptions.*.tiers[].flatAmountDecimal

`string`

flat_amount_decimal is flat_amount as a decimal string.

### spec.currencyOptions.*.tiers[].unitAmount

`int64` · optional (explicit presence)

unit_amount is charged per unit in the tier, in the currency's smallest unit.

- rule: {"int64":{"gte":"0"}}

### spec.currencyOptions.*.tiers[].unitAmountDecimal

`string`

unit_amount_decimal is unit_amount with up to 12 decimal places.

### spec.currencyOptions.*.customUnitAmount

`StripePriceCustomUnitAmount`

custom_unit_amount lets the customer choose the amount in this currency.

- rule: custom_unit_amount keeps minimum <= preset <= maximum

### spec.currencyOptions.*.customUnitAmount.minimum

`int64` · optional (explicit presence)

minimum is the smallest amount the customer may enter, at least Stripe's minimum charge.

- rule: {"int64":{"gte":"0"}}

### spec.currencyOptions.*.customUnitAmount.maximum

`int64` · optional (explicit presence)

maximum is the largest amount the customer may enter.

- rule: {"int64":{"gte":"0"}}

### spec.currencyOptions.*.customUnitAmount.preset

`int64` · optional (explicit presence)

preset is the amount the field starts at.

- rule: {"int64":{"gte":"0"}}

### spec.lookupKey

`string`

lookup_key is a stable name the application retrieves the price by ("pro-monthly"), up to
200 characters, unique among the account's prices. It updates in place.

- rule: {"string":{"maxLen":"200"}}

### spec.transferLookupKey

`bool`

transfer_lookup_key moves lookup_key to this price from whichever price holds it, in the same
call. Set it when a replacement price must take over the key (an amount change on a price the
application finds by lookup_key); without it, creating the replacement fails while the old
price still holds the key. The provider sends it on every apply while it is set.

### spec.nickname

`string`

nickname is a note about the price for the account's team, hidden from customers.

### spec.taxBehavior

`enum`

tax_behavior is whether the amount includes tax. Unset, the account's default tax behavior
applies. Once inclusive or exclusive, Stripe refuses to change it.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `tax_behavior_unset`
- `exclusive` -- exclusive adds tax on top of the amount.
- `inclusive` -- inclusive includes tax in the amount.
- `unspecified` -- unspecified leaves the choice open until it is set, once, to inclusive or exclusive.

### spec.active

`bool` · optional (explicit presence)

active is whether the price can be used for new purchases. Setting it false archives the price
without destroying the resource; subscriptions on it keep billing.

- default: `true`

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the price. A key removed here is removed in
Stripe on the next apply; removing the whole map leaves Stripe's keys in place.

## Validation Rules

- `spec.tiered_shape`: a tiered price sets tiers_mode and tiers (the last tier's up_to is inf) and no unit_amount, unit_amount_decimal or custom_unit_amount
- `spec.per_unit_shape`: a per-unit price sets exactly one of unit_amount, unit_amount_decimal and custom_unit_amount, and no tiers or tiers_mode (set billing_scheme tiered for tiers)
- `spec.transform_quantity_without_tiers`: transform_quantity cannot be combined with tiers: use a per-unit price, or drop transform_quantity
- `spec.currency_options_shape`: currency_options are the other currencies: the main currency is not one of them, and each option follows the price's scheme (tiers for a tiered price, an amount for a per-unit one)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripePrice, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the price's Stripe id (price_...), the value Checkout sessions, subscriptions, payment links and a StripeBillingPortalConfiguration's switchable prices reference. It changes when the price is replaced. |
| `status.outputs.type` | `string` | type is one_time or recurring, as Stripe derives it from spec.recurring. |
| `status.outputs.active` | `bool` | active is false once the price is archived (by destroy, by spec.active, or in the Dashboard). |
| `status.outputs.lookup_key` | `string` | lookup_key is the stable name the application retrieves the price by, when one is set. |
| `status.outputs.product` | `string` | product is the id of the product the price charges for (prod_...). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.product` | StripeProduct | `status.outputs.id` |
| `spec.recurring.meter` | StripeBillingMeter | `status.outputs.id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| StripeBillingPortalConfiguration | `spec.features.subscriptionUpdate.products[].prices` | `status.outputs.id` |
| StripePaymentLink | `spec.lineItems[].price` | `status.outputs.id` |
| StripePaymentLink | `spec.optionalItems[].price` | `status.outputs.id` |

## See Also

- [Overview](../README.md)
