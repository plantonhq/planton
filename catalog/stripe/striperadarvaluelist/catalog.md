# Stripe Radar Value List

Declares a Stripe Radar list -- blocked countries, trusted customers, known fraudulent emails -- and every item in it, for Radar rules to reference by alias. Items are checked against the list's type before Stripe sees them. One Infra Component per list.

## What Gets Created

When you deploy this Infra Component, the OpenTofu module creates, in the Stripe account your Stripe connection's key belongs to:

- **The list** -- its alias, name and item type
- **One item per value** -- each its own object in Stripe, created and deleted as the manifest changes

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Radar: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **Radar for Fraud Teams** -- to write the custom rules that use a list.

## Deploy

### Console

Open the deployment store, find **Stripe Radar Value List**, and click **Deploy**. Start from the **Blocked Countries** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeRadarValueList
metadata:
  name: blocked-countries
  org: acme-corp
  env: prod
spec:
  alias: blocked_countries
  name: Blocked countries
  itemType: country
  items:
    - KP
    - IR
```

```shell
planton apply -f stripe-radar-value-list.yaml
```

An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The alias is the contract** -- rules name the list as `@alias`; renaming it means changing every rule.

**Choose the type once** -- changing `itemType` replaces the list and every item.

**Destroy deletes** -- the list and its items; Stripe refuses while a rule still uses the list.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The list's Stripe id | Audits |
| `alias` | The name rules reference | Radar rules |
| `item_ids` | Value to item id | Import of the items |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Blocked countries** -- countries a rule refuses. Start from the **Blocked Countries** preset.

**Trusted customers** -- customers a rule lets through. Start from the **Trusted Customers** preset.

## Works With

- [**Stripe Payment Method Configuration**](/infra-catalog/stripe-payment-method-configuration) -- the methods whose payments Radar screens.
