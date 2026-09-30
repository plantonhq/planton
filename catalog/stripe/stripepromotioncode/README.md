# StripePromotionCode

Declares a code customers type to redeem a StripeCoupon -- `LAUNCH25` in Checkout, the customer portal or a payment link -- with its own limits: first-time customers only, a minimum order, an expiry, a redemption cap, or one customer. See Stripe's [promotion codes](https://docs.stripe.com/billing/subscriptions/coupons#promotion-codes).

## When to Use

- **A campaign in one chart**: the code names its coupon by reference, so the coupon and its codes deploy together.
- **Codes reviewed like code**: a typo such as `LAUNCH 25` is refused before anything runs.

Declare a code here only if nothing else owns it.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePromotionCode
metadata:
  name: launch25
  org: acme-corp
  env: production
spec:
  coupon:
    valueFrom:
      name: launch-25
  code: LAUNCH25
  maxRedemptions: 500
  restrictions:
    firstTimeTransaction: true
```

## Fields

| Field | Description |
|---|---|
| `coupon` | The coupon it redeems (required), by reference to StripeCoupon. **Replaces** |
| `code` | What customers type: letters, digits and dashes, up to 500. Unset: Stripe generates one. **Replaces** |
| `customer` / `customerAccount` | Only this customer. **Replaces** |
| `expiresAt` | When it stops working, in Unix seconds. **Replaces** |
| `maxRedemptions` | How many times it can be redeemed. **Replaces** |
| `restrictions` | First-time customers, a minimum order (with other currencies). **Replaces** |
| `active` | Whether it can be redeemed (default `true`). Changes in place |
| `metadata` | Key-value pairs stored on the code. Changes in place |

## Key Behaviors

- **Who may redeem a code is fixed in Stripe.** Changing any field marked **Replaces** deactivates the old code and creates a new one; the same code keeps working, because a code only has to be unique among active codes.
- **Destroy deactivates the code.** Nobody can redeem it; discounts already applied stay; the code can be declared again later.
- **A code can't outlast its coupon**: its expiry and cap can't exceed the coupon's.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The promotion code's Stripe id (`promo_...`) |
| `code` | What customers type: the declared code, or the one Stripe generated |
| `active` | `false` once deactivated |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
