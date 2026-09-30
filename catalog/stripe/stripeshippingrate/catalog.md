# Stripe Shipping Rate

Declares a shipping option -- standard at 5 dollars in 3 to 5 business days, free express overnight -- that customers choose in Checkout or on a payment link. One Cloud Resource per option.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one shipping rate in the Stripe account your Stripe connection's key belongs to:

- **The price of shipping** -- a fixed amount, in one or more currencies
- **The delivery window** -- shown to customers when they choose it
- **How it is taxed** -- inclusive or exclusive, and its tax category

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Shipping Rates: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **One owner**: declare a shipping rate here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Shipping Rate**, and click **Deploy**. Start from the **Standard Shipping** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeShippingRate
metadata:
  name: standard-shipping
  org: acme-corp
  env: prod
spec:
  displayName: Standard
  fixedAmount:
    amount: 500
    currency: usd
```

```shell
planton apply -f stripe-shipping-rate.yaml
```

A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Name, amount and window are fixed** -- changing them creates a new rate; payment links that offer it are replaced with it.

**Other currencies change in place** -- add a euro or pound amount without a new rate.

**Destroy deactivates** -- new purchases can't choose it; orders already placed keep it.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The shipping rate's Stripe id | A payment link's shipping options, Checkout sessions |
| `active` | `false` once deactivated | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard shipping** -- a paid option in two currencies. Start from the **Standard Shipping** preset.

**Free express** -- free next-day shipping. Start from the **Free Express** preset.

## Works With

- [**Stripe Payment Link**](/cloud-catalog/stripe-payment-link) -- offers the rate on a hosted page, naming it by reference in `shippingOptions`.
- [**Stripe Product**](/cloud-catalog/stripe-product) -- the physical goods it ships.
