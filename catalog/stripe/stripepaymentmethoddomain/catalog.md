# Stripe Payment Method Domain

Registers a domain where your own checkout pages may show wallet buttons -- Apple Pay, Google Pay, Link, PayPal, Amazon Pay and Klarna -- and reports, wallet by wallet, whether each can appear and why not. One Cloud Resource per domain.

## What Gets Created

When you deploy this Cloud Resource, the OpenTofu module registers one domain in the Stripe account your Stripe connection's key belongs to:

- **The registration** -- the hostname your checkout page is served from
- **Each wallet's state** -- active or inactive, with Stripe's reason

## Before You Deploy

### Planton Setup

- **Stripe Provider Connection** -- an active connection with a restricted key and its mode (test or live). A key of the other mode is refused before anything runs.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Stripe Account

- **The connection's restricted key** needs **Payment Method Domains: Write** (Stripe Dashboard: Developers, API keys, the key's permissions).
- **For Apple Pay**: Stripe's domain association file served from the domain.

## Deploy

### Console

Open the deployment store, find **Stripe Payment Method Domain**, and click **Deploy**. Start from the **Checkout Domain** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: stripe.planton.dev/v1alpha1
kind: StripePaymentMethodDomain
metadata:
  name: checkout-domain
  org: acme-corp
  env: prod
spec:
  domainName: pay.acme.com
```

```shell
planton apply -f stripe-payment-method-domain.yaml
```

A Stack Job tracks the change in real time.

## Key Configuration

These are the decisions that matter. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The host, not a URL** -- `domainName` is `pay.acme.com`, never `https://pay.acme.com/checkout`.

**Read the wallet outputs** -- an inactive wallet says why in `status.outputs`.

**Destroy only forgets** -- the domain stays registered and enabled in Stripe; apply `enabled: false` first to turn wallets off.

## Outputs and Dependencies

### What This Component Consumes

This component has no foreign key dependencies.

### What This Component Provides

After provisioning, `status.outputs` contains:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `id` | The registration's Stripe id | Audits |
| `enabled` | Whether wallets may appear | Monitoring |
| `apple_pay_status`, `apple_pay_error_message` (and the same for each wallet) | Each wallet's state and reason | Troubleshooting a missing button |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Checkout domain** -- wallets on your checkout host. Start from the **Checkout Domain** preset.

**Wallets paused** -- a registration with wallets off, the step before retiring it. Start from the **Wallets Paused** preset.

## Works With

- [**Stripe Payment Method Configuration**](/cloud-catalog/stripe-payment-method-configuration) -- which methods checkout offers; wallets among them appear on registered domains.
