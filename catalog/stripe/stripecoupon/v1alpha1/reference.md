# StripeCoupon

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripeCouponSpec declares a discount: a percentage or a fixed amount off, once, for a number of
months, or forever, optionally only on some products. Customers redeem it through a
StripePromotionCode, or the application applies it to a subscription or Checkout session.

A coupon's discount can never change in Stripe. Changing the amount, percentage, currency,
duration, months, redemption limit, redeem-by date or products REPLACES the coupon: Planton
deletes the old one and creates the new one, and every promotion code that references it is
replaced with it. Name, metadata and currency options update in place.

One owner per object: declare a coupon here only if nothing else creates or edits it. When the
application's own code or the Dashboard owns the account's discounts, leave its coupons there.

Destroy deletes the coupon. Customers who already applied it keep their discount; nobody new
can redeem it (Stripe's API reference, "Delete a coupon"). A coupon that cannot be read makes
the plan fail (the provider does not treat a missing coupon as gone); remove it from state and
apply again.

The key needs "Coupons" write (iac/permissions.yaml).

https://docs.stripe.com/billing/subscriptions/coupons
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/coupon

## Example

```yaml
# The canonical example: 25% off for three months on one product. The product
# is a literal id here so the manifest plans on its own; a real manifest
# references its StripeProduct (see the scenarios).
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeCoupon
metadata:
  name: launch-25
  org: e2e-org
  env: testing
spec:
  name: Launch 25% off
  percentOff: 25
  duration: repeating
  durationInMonths: 3
  maxRedemptions: 500
  appliesToProducts:
    - value: prod_pro
  metadata:
    campaign: launch
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` |  |  |  |
| `spec.percentOff` | `double` |  |  |  |
| `spec.amountOff` | `int64` |  |  |  |
| `spec.currency` | `string` |  |  |  |
| `spec.currencyOptions` | `map<string, int64>` |  |  |  |
| `spec.duration` | `enum` |  |  |  |
| `spec.durationInMonths` | `int64` |  |  |  |
| `spec.maxRedemptions` | `int64` |  |  |  |
| `spec.redeemBy` | `int64` |  |  |  |
| `spec.appliesToProducts` | `[]string \| valueFrom` |  |  | StripeProduct (`status.outputs.id`) |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.name

`string`

name is the coupon's name as customers see it on invoices and receipts. Unset, Stripe shows
the coupon's id. It updates in place.

### spec.percentOff

`double` · optional (explicit presence)

percent_off is the discount as a percentage, above 0 and at most 100. A coupon sets exactly
one of percent_off and amount_off. Changing it REPLACES the coupon.

- rule: {"double":{"lte":100,"gt":0}}

### spec.amountOff

`int64` · optional (explicit presence)

amount_off is the discount in the currency's smallest unit (cents for usd), with currency.
Changing it REPLACES the coupon.

- rule: {"int64":{"gt":"0"}}

### spec.currency

`string`

currency is the three-letter ISO code of amount_off, in lowercase ("usd"). Set it with
amount_off, never with percent_off. Changing it REPLACES the coupon.

- rule: currency is a lowercase three-letter ISO code, such as usd

### spec.currencyOptions

`map<string, int64>`

currency_options are amount_off in other currencies, keyed by lowercase currency code ("eur"),
so a customer paying in theirs gets the same discount. Only with amount_off. Changing them
REPLACES the coupon: Stripe refuses a new amount for a currency the coupon already has.

- rule: {"map":{"keys":{"string":{"pattern":"^[a-z]{3}$"}},"values":{"int64":{"gt":"0"}}}}

### spec.duration

`enum`

duration is how long the discount applies to a subscription. Unset, Stripe uses once.
Changing it REPLACES the coupon.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `duration_unspecified`
- `once` -- once applies to the first invoice only.
- `repeating` -- repeating applies for duration_in_months.
- `forever` -- forever applies to every invoice.

### spec.durationInMonths

`int64` · optional (explicit presence)

duration_in_months is how many months a repeating coupon applies for. Set it with duration
repeating, and only then. Changing it REPLACES the coupon.

- rule: {"int64":{"gte":"1"}}

### spec.maxRedemptions

`int64` · optional (explicit presence)

max_redemptions is how many times the coupon can be redeemed across all customers. Unset, no
limit. Changing it REPLACES the coupon.

- rule: {"int64":{"gte":"1"}}

### spec.redeemBy

`int64` · optional (explicit presence)

redeem_by is the moment after which the coupon can no longer be redeemed, in Unix seconds
(1798761599 is 2026-12-31T23:59:59Z; on Linux `date -u -d 2026-12-31T23:59:59Z +%s` computes
one). A subscription that already has it keeps the discount after this time. Changing it
REPLACES the coupon.

- rule: {"int64":{"gt":"0"}}

### spec.appliesToProducts

`[]string | valueFrom`

applies_to_products limits the discount to these products (prod_...). Reference a
StripeProduct. Unset, the coupon applies to everything. Changing them REPLACES the coupon.

- references: StripeProduct (`status.outputs.id`)
- rule: {"repeated":{"items":{"cel":[{"id":"spec.applies_to_products.format","message":"a product is a Stripe product id (prod_...), or a reference to a StripeProduct","expression":"!has(this.value) || this.value.startsWith('prod_')"}]}}}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripeProduct, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the coupon. It updates in place.

## Validation Rules

- `spec.discount_shape`: a coupon takes exactly one discount: percent_off, or amount_off with its currency
- `spec.currency_with_amount`: currency belongs to amount_off: set both, or neither (a percentage coupon has no currency)
- `spec.currency_options_shape`: currency_options are amount_off in other currencies: they need amount_off, and the main currency is not one of them
- `spec.months_with_repeating`: duration_in_months goes with duration repeating: set both, or neither
- `spec.applies_to_products.unique`: each product is named once: remove the repeated product id

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripeCoupon, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the coupon's Stripe id, the value a StripePromotionCode, a subscription or a Checkout session references. It changes when the coupon is replaced. |
| `status.outputs.valid` | `bool` | valid is whether the coupon can still be applied to new customers: false once it has passed redeem_by or reached max_redemptions. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.appliesToProducts` | StripeProduct | `status.outputs.id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| StripePromotionCode | `spec.coupon` | `status.outputs.id` |

## See Also

- [Overview](../README.md)
