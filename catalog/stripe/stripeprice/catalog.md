# Stripe Price

Declares how much, how often and in which currencies a product is charged -- a flat monthly amount, a per-seat price, graduated usage tiers, or an amount the customer chooses. One Infra Component per price.

## What Gets Created

When you deploy this Infra Component, the OpenTofu module creates one price on the product it names, in the Stripe account your Stripe connection's key belongs to:

- **The charge** -- amount, currency, and how quantity or usage turns into an amount
- **The schedule** -- one-time, or recurring at an interval
- **Other currencies** -- the same price in euros, pounds or any supported currency

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Prices: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **The product** -- a Stripe Product Infra Component, or an existing product's id.
- **One owner**: declare a price here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Price**, and click **Deploy**. Start from the **Monthly Subscription** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePrice
metadata:
  name: pro-monthly
  org: acme-corp
  env: prod
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

```shell
planton apply -f stripe-price.yaml
```

An Infra Job tracks the change in real time.

### InfraChart

When the product and this price deploy together, name the product with ValueFromRef:

```yaml
spec:
  product:
    valueFrom:
      kind: StripeProduct
      name: pro-plan
      fieldPath: status.outputs.id
```

The InfraPipeline deploys the product first, then this price. A portal or checkout that references the price by `status.outputs.id` follows it when an amount change replaces it.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**An amount change is a new price** -- Stripe never changes a price's amount. Planton creates the new price, then archives the old one; existing subscribers keep the old amount.

**Keep the lookup key across a change** -- set `transferLookupKey` with the change, so the key moves to the new price in the same call.

**Pick the model once** -- per unit, tiered, or customer-chosen; changing it replaces the price.

**Destroy archives** -- subscriptions on the price keep billing; new purchases are refused.

## Outputs and Dependencies

### What This Kind Consumes

| Field | Kind | Output |
|-------|------|--------|
| `product` | Stripe Product | `status.outputs.id` |
| `recurring.meter` | Stripe Billing Meter | `status.outputs.id` |

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The price's Stripe id | Checkout sessions, subscriptions, a portal's switchable prices |
| `type` | `one_time` or `recurring` | Audits |
| `active` | `false` once archived | Audits |
| `lookup_key` | The stable name | Your checkout code |
| `product` | The product's id | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Monthly subscription** -- a flat monthly price in three currencies. Start from the **Monthly Subscription** preset.

**Graduated usage** -- metered usage in tiers that get cheaper with volume, billing what a Stripe Billing Meter counts. Start from the **Graduated Usage** preset.

## Works With

- [**Stripe Product**](/infra-catalog/stripe-product) -- what the price charges for.
- [**Stripe Billing Meter**](/infra-catalog/stripe-billing-meter) -- counts the usage a metered price bills.
- [**Stripe Payment Link**](/infra-catalog/stripe-payment-link) -- a hosted page that sells the price.
- [**Stripe Billing Portal Configuration**](/infra-catalog/stripe-billing-portal-configuration) -- lets customers switch between prices.
