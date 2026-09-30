# StripePromotionCode

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `stripe.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

StripePromotionCodeSpec declares a code customers type to redeem a StripeCoupon -- LAUNCH25 in
Checkout, the customer portal or a payment link -- with its own limits: first-time customers
only, a minimum order, an expiry, a redemption cap, or one customer.

Everything that shapes who may redeem the code is fixed in Stripe. Changing the coupon, code,
customer, expiry, redemption limit or restrictions REPLACES the promotion code: Planton
deactivates the old one and creates the new one. Active and metadata update in place.

A code's expiry can't be later than its coupon's redeem_by, and its redemption limit can't
exceed the coupon's max_redemptions (Stripe's API reference); Stripe refuses either at apply.

One owner per object: declare a code here only if nothing else creates or edits it.

Stripe never deletes a promotion code. Destroy deactivates it (active = false): nobody can
redeem it, discounts already applied stay. A code only has to be unique among active codes, so
the same code can be declared again later. A code that cannot be read makes the plan fail (the
provider does not treat a missing code as gone); remove it from state and apply again.

The key needs "Promotion Codes" write (iac/permissions.yaml).

https://docs.stripe.com/billing/subscriptions/coupons#promotion-codes
https://registry.terraform.io/providers/stripe/stripe/latest/docs/resources/promotion_code

## Example

```yaml
# The canonical example: LAUNCH25 for first-time customers, 500 redemptions
# at most. The coupon is a literal id here so the manifest plans on its own;
# a real manifest references its StripeCoupon (see the scenarios).
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePromotionCode
metadata:
  name: launch25
  org: e2e-org
  env: testing
spec:
  coupon:
    value: Z4OV52SU
  code: LAUNCH25
  maxRedemptions: 500
  expiresAt: 1798761599
  restrictions:
    firstTimeTransaction: true
  metadata:
    campaign: launch
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.coupon` | `string \| valueFrom` | yes |  | StripeCoupon (`status.outputs.id`) |
| `spec.code` | `string` |  |  |  |
| `spec.customer` | `string` |  |  |  |
| `spec.customerAccount` | `string` |  |  |  |
| `spec.expiresAt` | `int64` |  |  |  |
| `spec.maxRedemptions` | `int64` |  |  |  |
| `spec.restrictions` | `StripePromotionCodeRestrictions` |  |  |  |
| `spec.restrictions.firstTimeTransaction` | `bool` |  |  |  |
| `spec.restrictions.minimumAmount` | `int64` |  |  |  |
| `spec.restrictions.minimumAmountCurrency` | `string` |  |  |  |
| `spec.restrictions.currencyOptions` | `map<string, int64>` |  |  |  |
| `spec.active` | `bool` |  | `true` |  |
| `spec.metadata` | `map<string, string>` |  |  |  |

## Field Details

### spec.coupon

`string | valueFrom` · required

coupon is the coupon the code redeems. Reference a StripeCoupon. Changing it REPLACES the
code.

- references: StripeCoupon (`status.outputs.id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: StripeCoupon, name: <that resource's name>, fieldPath: status.outputs.id}} -- a bare string does not parse

### spec.code

`string`

code is what customers type, up to 500 letters, digits and dashes ("LAUNCH25"). Regardless of
case, it is unique among the account's active codes for each customer. Unset, Stripe generates
one (read it from status.outputs.code). Changing it REPLACES the code.

- rule: code is up to 500 letters (a-z, A-Z), digits (0-9) and dashes (-), with no spaces

### spec.customer

`string`

customer limits the code to one customer (cus_...). Unset, any customer can redeem it.
Changing it REPLACES the code.

- rule: customer is a Stripe customer id (cus_...)

### spec.customerAccount

`string`

customer_account limits the code to one customer that is a Stripe account (an Accounts v2
customer) instead of a customer object. Set customer or customer_account, not both. Changing
it REPLACES the code.

### spec.expiresAt

`int64` · optional (explicit presence)

expires_at is the moment after which the code can no longer be redeemed, in Unix seconds
(1798761599 is 2026-12-31T23:59:59Z; on Linux `date -u -d 2026-12-31T23:59:59Z +%s` computes
one). It can't be later than the coupon's redeem_by. Changing it REPLACES the code.

- rule: {"int64":{"gt":"0"}}

### spec.maxRedemptions

`int64` · optional (explicit presence)

max_redemptions is how many times the code can be redeemed. It can't exceed the coupon's
max_redemptions. Unset, no limit beyond the coupon's. Changing it REPLACES the code.

- rule: {"int64":{"gte":"1"}}

### spec.restrictions

`StripePromotionCodeRestrictions`

restrictions limit which orders the code applies to. Changing them REPLACES the code.

- rule: minimum_amount and minimum_amount_currency go together: set both, or neither
- rule: currency_options are minimum_amount in other currencies: they need minimum_amount, and its currency is not one of them

### spec.restrictions.firstTimeTransaction

`bool` · optional (explicit presence)

first_time_transaction limits the code to customers who have never paid before.

### spec.restrictions.minimumAmount

`int64` · optional (explicit presence)

minimum_amount is the smallest order total the code applies to, in the smallest unit of
minimum_amount_currency.

- rule: {"int64":{"gt":"0"}}

### spec.restrictions.minimumAmountCurrency

`string`

minimum_amount_currency is the lowercase three-letter ISO code of minimum_amount ("usd").

- rule: minimum_amount_currency is a lowercase three-letter ISO code, such as usd

### spec.restrictions.currencyOptions

`map<string, int64>`

currency_options are minimum_amount in other currencies, keyed by lowercase currency code
("eur").

- rule: {"map":{"keys":{"string":{"pattern":"^[a-z]{3}$"}},"values":{"int64":{"gt":"0"}}}}

### spec.active

`bool` · optional (explicit presence)

active is whether the code can be redeemed. Setting it false deactivates the code without
destroying the resource; setting it true again reactivates it, if no other active code has
taken the same code meanwhile.

- default: `true`

### spec.metadata

`map<string, string>`

metadata is a set of key-value pairs stored on the code. It updates in place.

## Validation Rules

- `spec.one_customer`: a code is limited to one customer: set customer or customer_account, not both

## Outputs

Reference an output from another manifest as `valueFrom: {kind: StripePromotionCode, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.id` | `string` | id is the promotion code's Stripe id (promo_...). It changes when the code is replaced. |
| `status.outputs.code` | `string` | code is what customers type: the declared code, or the one Stripe generated. |
| `status.outputs.active` | `bool` | active is false once the code is deactivated (by destroy, by spec.active, or in the Dashboard). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.coupon` | StripeCoupon | `status.outputs.id` |

## See Also

- [Overview](../README.md)
