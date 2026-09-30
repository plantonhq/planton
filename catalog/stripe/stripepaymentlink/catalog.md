# Stripe Payment Link

Declares a Stripe-hosted payment page at a public address -- a one-time purchase, a subscription with a free trial -- that sells declared prices. One Cloud Resource per page.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module creates one payment link in the Stripe account your Stripe connection's key belongs to:

- **The page** -- at a public address that takes payments from the moment it is created
- **What it sells** -- one or more prices, with quantities, and optional extras
- **What it collects** -- addresses, names, phone numbers, tax ids, custom answers, consent
- **What follows payment** -- Stripe's confirmation page or a redirect to your site

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Payment Links: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **The prices** -- Stripe Price Cloud Resources, or existing prices' ids.
- **Terms of service** -- set your terms URL in the Dashboard's public details before requiring terms acceptance.
- **One owner**: declare a link here only if your application and the Dashboard do not also manage it.

## Deploy

### Console

Open the deployment store, find **Stripe Payment Link**, and click **Deploy**. Start from the **Subscription With a Trial** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentLink
metadata:
  name: pro-monthly-link
  org: acme-corp
  env: prod
spec:
  lineItems:
    - price:
        valueFrom:
          name: pro-monthly
      quantity: 1
  allowPromotionCodes: true
```

```shell
planton apply -f stripe-payment-link.yaml
```

A Stack Job tracks the change in real time. The page takes payments as soon as the job finishes.

### InfraChart

When the prices and this link deploy together, name each price with ValueFromRef, and let your site read the link's address the same way:

```yaml
# StripePaymentLink
spec:
  lineItems:
    - price:
        valueFrom:
          kind: StripePrice
          name: pro-monthly
          fieldPath: status.outputs.id
      quantity: 1
---
# Your site's deployment
env:
  variables:
    - name: PRO_SIGNUP_URL
      valueFrom:
        kind: StripePaymentLink
        name: pro-monthly-link
        fieldPath: status.outputs.url
```

The InfraPipeline deploys the prices first, then the link, then your site. When a change replaces the link, your site picks up the new address on the same apply.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**It takes payments on apply** -- anyone with the address can pay for the declared prices.

**Prices and quantities are fixed** -- changing them creates a new link with a new address before the old one is deactivated; visitors to the old address see `inactiveMessage`.

**Connect fields move money** -- `transferData`, `onBehalfOf` and the application fees decide where future buyers' money goes. Changing them creates a new link.

**Destroy deactivates** -- the address shows `inactiveMessage`; payments already taken stay.

## Outputs and Dependencies

### What This Component Consumes

| Field | Kind | Output |
|-------|------|--------|
| `lineItems[].price`, `optionalItems[].price` | Stripe Price | `status.outputs.id` |
| `shippingOptions[].shippingRate` | Stripe Shipping Rate | `status.outputs.id` |

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The payment link's Stripe id | Audits, your code's lookups |
| `url` | The page's public address | Your site's buttons and emails |
| `active` | `false` once deactivated | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One-time purchase** -- a single price with invoices and a confirmation message. Start from the **One-Time Purchase** preset.

**Subscription with a trial** -- a monthly plan with 14 free days and a redirect to your site. Start from the **Subscription With a Trial** preset.

## Works With

- [**Stripe Price**](/cloud-catalog/stripe-price) -- the prices the page sells.
- [**Stripe Shipping Rate**](/cloud-catalog/stripe-shipping-rate) -- the shipping options it offers.
- [**Stripe Promotion Code**](/cloud-catalog/stripe-promotion-code) -- codes buyers enter when promotion codes are allowed.
