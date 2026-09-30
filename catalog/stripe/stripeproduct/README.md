# StripeProduct

Declares one thing a Stripe account sells -- a plan, a seat, a physical good -- as customers see it in Checkout, Payment Links, subscriptions and pricing tables, and the [entitlement features](https://docs.stripe.com/billing/entitlements) buying it grants.

## When to Use

- **A catalog reviewed like code**: plans and goods live in files, the same in every environment.
- **Plans that grant capabilities**: reference StripeEntitlementFeature resources under `features`, and every subscriber's entitlements follow.
- **Prices declared beside it**: each StripePrice names this product.

Declare a product here only if nothing else owns it. When your application or the Dashboard already manages your price list, leave its products there.

## Quick Start

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeProduct
metadata:
  name: pro-plan
  org: acme-corp
  env: production
spec:
  name: Pro
  description: For growing teams
  unitLabel: seat
  features:
    - valueFrom:
        name: api-access
```

## Fields

| Field | Description |
|---|---|
| `name` | What customers see (required) |
| `description` | The long-form explanation customers see |
| `active` | Whether it can be bought (default `true`); `false` archives it |
| `type` | `service` (Stripe's default) or `good`. Changing it **replaces** the product |
| `images` | Up to 8 public image URLs |
| `marketingFeatures` | Up to 15 display lines for pricing tables |
| `packageDimensions` | Height, length and width in inches, weight in ounces |
| `shippable` | Whether it ships |
| `statementDescriptor` | Card-statement text for subscription payments, up to 22 characters; a service only |
| `taxCode` | The Stripe Tax category (`txcd_...`) |
| `unitLabel` | What a quantity counts ("seat"); a service only |
| `url` | A public page for the product |
| `metadata` | Key-value pairs stored on the product |
| `features` | Entitlement features a purchase grants, by reference to StripeEntitlementFeature |

## Key Behaviors

- **Destroy archives the product.** It can no longer be bought; existing subscriptions keep billing. Stripe keeps it.
- **Features are links.** Adding one attaches it, removing one deletes the link, and subscribers' entitlements follow.
- **A removed setting stays in Stripe**: the provider sends only values that are set. Change a value rather than deleting it.
- **No hidden prices.** Prices are StripePrice resources; the provider's inline default price is not offered, because Planton could never track or archive it.
- **Runs on OpenTofu only.** Planton refuses any other engine for this kind before anything runs.

## Outputs

| Output | Description |
|---|---|
| `id` | The product's Stripe id (`prod_...`), what a StripePrice's `product` references |
| `active` | `false` once archived |
| `default_price` | The price Stripe treats as the default, when one is set (never set by this kind) |
| `product_feature_ids` | Each granted feature's id mapped to the id of the link that attaches it |

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
