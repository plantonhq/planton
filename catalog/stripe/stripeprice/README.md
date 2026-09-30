# StripePrice

Declares how much, how often and in which currencies a StripeProduct is charged: a flat amount, an amount per unit, graduated or volume tiers, or an amount the customer chooses, one-time or recurring. See Stripe's [pricing models](https://docs.stripe.com/products-prices/pricing-models).

## When to Use

- **A plan's prices in files**: monthly and yearly, in several currencies, the same in every environment.
- **Price changes that keep your code working**: an amount change creates a new price first, and `transferLookupKey` moves your lookup key to it.
- **Usage-based pricing**: per unit, tiered, metered, or divided before billing.

Declare a price here only if nothing else owns it. When your application or the Dashboard already manages your price list, leave its prices there.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePrice
metadata:
  name: pro-monthly
  org: acme-corp
  env: production
spec:
  product:
    valueFrom:
      name: pro-plan
  currency: usd
  unitAmount: 4900
  recurring:
    interval: month
  lookupKey: pro-monthly
```

## Fields

| Field | Description |
|---|---|
| `product` | The product charged for (required), by reference to a StripeProduct. **Replaces** |
| `currency` | Lowercase ISO code (required). **Replaces** |
| `unitAmount` / `unitAmountDecimal` / `customUnitAmount` | Exactly one, for a per-unit price. **Replaces** |
| `billingScheme`, `tiersMode`, `tiers` | A tiered price: its mode and steps, the last `upTo: inf`. **Replaces** |
| `transformQuantity` | Divide the quantity before billing; not with tiers. **Replaces** |
| `recurring` | Interval, count, usage type, trial days, and a metered price's meter by reference to a StripeBillingMeter. Unset: one-time. **Replaces** |
| `currencyOptions` | Amounts in other currencies, keyed by currency. Changes in place |
| `lookupKey` | A stable name your code retrieves the price by. Changes in place |
| `transferLookupKey` | Move the lookup key to this price from whichever price holds it |
| `nickname` | A note for your team, hidden from customers |
| `taxBehavior` | `exclusive`, `inclusive` or `unspecified`; fixed once inclusive or exclusive |
| `active` | Whether it can be used for new purchases (default `true`) |
| `metadata` | Key-value pairs stored on the price |

## Key Behaviors

- **An amount never changes in Stripe.** Changing any field marked **Replaces** creates the new price, then archives the old one. Existing subscriptions keep the old price.
- **Destroy archives the price.** Subscriptions on it keep billing; new purchases are refused.
- **Currency options are write-only in effect**: Stripe does not report them back, so a Dashboard change is not detected and a removed currency stays.
- **No hidden products.** The product is a StripeProduct; the provider's inline product is not offered, because Planton could never track or archive it.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The price's Stripe id (`price_...`); it changes when the price is replaced |
| `type` | `one_time` or `recurring` |
| `active` | `false` once archived |
| `lookup_key` | The stable name, when set |
| `product` | The product's id |
