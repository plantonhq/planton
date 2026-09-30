# Stripe Product

Declares one thing your Stripe account sells -- a subscription plan, a seat, a physical good -- exactly as customers see it, and the entitlement features a purchase grants. Prices are separate Cloud Resources that name the product. One Cloud Resource per product.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates, in the Stripe account your Stripe connection's key belongs to:

- **The product** -- its name, description, images and pricing-table lines
- **One feature link per granted feature** -- each attaches a Stripe Entitlement Feature, so subscribers hold it

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Products: Write**, and **Entitlements: Write** when the product grants features (Stripe Dashboard: Developers, API keys, the key's permissions).
- **One owner**: declare a product here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Product**, and click **Deploy**. Start from the **SaaS Plan** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeProduct
metadata:
  name: pro-plan
  org: acme-corp
  env: prod
spec:
  name: Pro
  description: For growing teams
  unitLabel: seat
  features:
    - valueFrom:
        name: api-access
```

```shell
planton apply -f stripe-product.yaml
```

A Stack Job tracks the change in real time.

### InfraChart

When the entitlement features and this product deploy together, grant the features with ValueFromRef, so the product attaches exactly the features declared beside it:

```yaml
spec:
  features:
    - valueFrom:
        kind: StripeEntitlementFeature
        name: api-access
        fieldPath: status.outputs.id
```

The InfraPipeline deploys the features first, then this product.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**What customers read** -- `name`, `description` and `marketingFeatures` appear on Stripe-hosted pages and pricing tables.

**What a purchase grants** -- `features` references Stripe Entitlement Features; changing the list changes every subscriber's entitlements.

**Service or good** -- `type` is set once; changing it replaces the product and every price that references it.

**Destroy archives** -- the product can no longer be bought, and existing subscriptions keep billing.

## Outputs and Dependencies

### What This Component Consumes

| Field | Kind | Output |
|-------|------|--------|
| `features` | Stripe Entitlement Feature | `status.outputs.id` |

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The product's Stripe id | A Stripe Price's `product`; a portal's switchable products |
| `active` | `false` once archived | Audits |
| `default_price` | The default price, when set | Audits |
| `product_feature_ids` | Feature id to link id | Import of the feature links |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**SaaS plan** -- a subscription plan that grants entitlement features. Start from the **SaaS Plan** preset.

**Physical good** -- a shippable product with package dimensions. Start from the **Physical Good** preset.

## Works With

- [**Stripe Price**](/cloud-catalog/stripe-price) -- how much, how often and in which currencies the product is charged.
- [**Stripe Entitlement Feature**](/cloud-catalog/stripe-entitlement-feature) -- the capabilities the product grants.
- [**Stripe Billing Portal Configuration**](/cloud-catalog/stripe-billing-portal-configuration) -- lets customers switch between products and their prices.
