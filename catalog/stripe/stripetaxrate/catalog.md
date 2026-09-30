# Stripe Tax Rate

Declares a manual tax rate -- German VAT at 19%, New York sales tax -- that invoices, subscriptions, Checkout sessions and payment links apply. One Cloud Resource per rate.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one tax rate in the Stripe account your Stripe connection's key belongs to:

- **The rate** -- a percentage, included in the amount or added on top
- **Where it applies** -- a country, a state, a jurisdiction
- **How it reads on invoices** -- its display name and tax type

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Tax Rates: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **One owner**: declare a tax rate here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Tax Rate**, and click **Deploy**. Start from the **EU VAT, Exclusive** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeTaxRate
metadata:
  name: de-vat
  org: acme-corp
  env: prod
spec:
  displayName: VAT
  percentage: 19
  country: DE
  taxType: vat
```

```shell
planton apply -f stripe-tax-rate.yaml
```

A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The rate is fixed** -- changing the percentage or inclusiveness creates a new tax rate; subscriptions keep the old one until you move them.

**Everything else changes in place** -- name, description, country, state, jurisdiction and tax type.

**Destroy deactivates** -- the rate still applies wherever it is already used.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The tax rate's Stripe id | Your code's invoices, subscriptions and Checkout sessions |
| `active` | `false` once deactivated | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EU VAT** -- a country's standard VAT rate, added on top. Start from the **EU VAT, Exclusive** preset.

**US sales tax** -- a state and city rate. Start from the **US Sales Tax** preset.

## Works With

- [**Stripe Price**](/cloud-catalog/stripe-price) -- the prices the rate is applied to.
- [**Stripe Shipping Rate**](/cloud-catalog/stripe-shipping-rate) -- shipping that may be taxed too.
