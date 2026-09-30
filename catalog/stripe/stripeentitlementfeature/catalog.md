# Stripe Entitlement Feature

Declares one capability a customer can be entitled to -- API access, single sign-on, extra seats -- identified by a lookup key your application checks. Products grant features, and Stripe tells your application which features each subscribed customer holds. One Cloud Resource per feature.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one entitlement feature in the Stripe account your Stripe connection's key belongs to:

- **The lookup key** -- the name your code checks a customer's active entitlements for
- **The name** -- your team's label in the Dashboard

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Entitlements: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).

## Deploy

### Console

Open the deployment store, find **Stripe Entitlement Feature**, and click **Deploy**. Start from the **API Access** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeEntitlementFeature
metadata:
  name: api-access
  org: acme-corp
  env: prod
spec:
  lookupKey: api-access
  name: API access
```

```shell
planton apply -f stripe-entitlement-feature.yaml
```

A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Choose the lookup key once** -- your code checks it, and changing it replaces the feature.

**Grant it from products** -- a Stripe Product lists the features its purchase grants; the feature itself grants nothing.

**Destroy archives** -- Stripe keeps the feature, archived, and it can no longer be attached to new products.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The feature's Stripe id | A Stripe Product's `features` |
| `lookup_key` | The stable name | Your application's entitlement checks |
| `active` | `false` once archived | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**API access** -- the capability to call your API. Start from the **API Access** preset.

**Seat management** -- the capability to invite teammates. Start from the **Seat Management** preset.

## Works With

- [**Stripe Product**](/cloud-catalog/stripe-product) -- grants the feature to everyone who buys it.
