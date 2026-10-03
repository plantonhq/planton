# Stripe Tax Registration

Declares one place your account is registered to collect tax with Stripe Tax -- VAT in Germany under the EU's One-Stop Shop, sales tax in Texas. One Infra Component per registration.

## What Gets Created

When you deploy this Infra Component, the OpenTofu module creates one tax registration in the Stripe account your Stripe connection's key belongs to:

- **Where** -- a country, and a province or state where the country has them
- **The kind of registration** -- standard, simplified, the EU's One-Stop Shops, or a US state or local tax
- **When** -- the date collection starts, and optionally the date it stops

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **Stripe Tax** -- the account uses Stripe Tax; a registration tells it where to collect.
- **The connection's restricted key** needs **Tax Registrations: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **One owner**: declare a registration here only if your team does not also manage it in the Dashboard.

## Deploy

### Console

Open the deployment store, find **Stripe Tax Registration**, and click **Deploy**. Start from the **EU One-Stop Shop** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripeTaxRegistration
metadata:
  name: de-oss
  org: acme-corp
  env: prod
spec:
  country: DE
  type: oss_union
  activeFrom: 1830297600  # your start: now or later, at most five years ahead, in Unix seconds
```

```shell
planton apply -f stripe-tax-registration.yaml
```

An Infra Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Applying starts collection** -- from the start date, Stripe Tax collects tax in that place on every payment that uses automatic tax. The start date must be now or later when the registration is created.

**Only the dates change in place** -- any other change creates a new registration, and the old one keeps collecting until it expires. Expire the old one first.

**Destroy only forgets** -- Stripe keeps the registration and keeps collecting. Set an expiry date to stop.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies.

### What This Kind Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The registration's Stripe id | Audits and your tax reports |
| `status` | `scheduled`, `active` or `expired` | Knowing whether collection has started |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**EU One-Stop Shop** -- one registration in your EU home country covers sales to consumers across the EU. Start from the **EU One-Stop Shop** preset.

**US state sales tax** -- one registration per state you collect in. Start from the **US State Sales Tax** preset.

## Works With

- [**Stripe Payment Link**](/infra-catalog/stripe-payment-link) -- hosted payment pages that can collect tax automatically.
- [**Stripe Price**](/infra-catalog/stripe-price) -- prices whose tax behavior Stripe Tax reads.
- [**Stripe Tax Rate**](/infra-catalog/stripe-tax-rate) -- manual rates, for accounts that do not use Stripe Tax.
