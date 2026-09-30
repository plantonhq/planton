# StripeCoupon

Declares a discount: a percentage or an amount off, once, for a number of months, or forever, optionally only on some products. Customers redeem it with a StripePromotionCode, or your application applies it to a subscription or Checkout session. See Stripe's [coupons](https://docs.stripe.com/billing/subscriptions/coupons).

## When to Use

- **A campaign in one chart**: the coupon, the products it discounts, and the codes that redeem it, all by reference.
- **The same discounts in every environment**, reviewed like code.

Declare a coupon here only if nothing else owns it. When your application or the Dashboard already manages your discounts, leave them there.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeCoupon
metadata:
  name: launch-25
  org: acme-corp
  env: production
spec:
  name: Launch 25% off
  percentOff: 25
  duration: repeating
  durationInMonths: 3
```

## Fields

| Field | Description |
|---|---|
| `percentOff` / `amountOff` with `currency` | Exactly one discount. **Replaces** |
| `duration`, `durationInMonths` | `once` (Stripe's default), `repeating` for some months, or `forever`. **Replaces** |
| `maxRedemptions`, `redeemBy` | Limits across all customers; `redeemBy` is Unix seconds. **Replaces** |
| `appliesToProducts` | Only these products, by reference to StripeProduct. **Replaces** |
| `currencyOptions` | `amountOff` in other currencies, keyed by currency. Changes in place |
| `name` | What customers see on invoices. Changes in place |
| `metadata` | Key-value pairs stored on the coupon. Changes in place |

## Key Behaviors

- **A discount never changes in Stripe.** Changing any field marked **Replaces** deletes the coupon and creates a new one, and every promotion code on it is replaced with it.
- **Destroy deletes the coupon.** Customers who already applied it keep their discount; nobody new can redeem it.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The coupon's Stripe id; it changes when the coupon is replaced |
| `valid` | Whether new customers can still redeem it |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
