# StripeShippingRate

Declares a shipping option customers choose in Checkout or on a payment link: its name, a fixed amount in one or more currencies, and how long delivery takes. See Stripe's [charge for shipping](https://docs.stripe.com/payments/during-payment/charge-shipping).

## When to Use

- **Shipping options in files**: standard, express and free, the same in every environment.
- **A payment page and its shipping in one chart**: a StripePaymentLink names its shipping rates by reference.

Declare a shipping rate here only if nothing else owns it.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeShippingRate
metadata:
  name: standard-shipping
  org: acme-corp
  env: production
spec:
  displayName: Standard
  fixedAmount:
    amount: 500
    currency: usd
  deliveryEstimate:
    minimum:
      unit: business_day
      value: 3
    maximum:
      unit: business_day
      value: 5
```

## Fields

| Field | Description |
|---|---|
| `displayName` | The option's name in Checkout (required). **Replaces** |
| `fixedAmount.amount`, `fixedAmount.currency` | What it costs (required); 0 is free. **Replaces** |
| `fixedAmount.currencyOptions` | The amount in other currencies, keyed by currency. Changes in place |
| `deliveryEstimate` | The delivery window: minimum, maximum, or both. **Replaces** |
| `taxBehavior` | `exclusive`, `inclusive` or `unspecified`; fixed once inclusive or exclusive |
| `taxCode` | The Stripe Tax category (`txcd_92010001` is shipping). **Replaces** |
| `active` | Whether new purchases can choose it (default `true`) |
| `metadata` | Key-value pairs stored on the rate |

## Key Behaviors

- **A rate's name, amount and window never change in Stripe.** Changing any field marked **Replaces** deactivates the old rate and creates a new one; a payment link that offers it is replaced with it.
- **Destroy deactivates the rate.** New purchases can't choose it; orders already placed keep it.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The shipping rate's Stripe id (`shr_...`); it changes when the rate is replaced |
| `active` | `false` once deactivated |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
